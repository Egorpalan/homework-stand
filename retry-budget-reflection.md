# Задание 2 — retry-бюджет (throttler) при полном даунтайме

**Контекст:** `profile-service` → `analytic-service`, метод `GetUserTaskCount`

## Контекст экспериментов

| Параметр | Значение |
|---|---|
| Нагрузка | ~10 RPS (framer: `const 10`, 10s, 1 client) |
| Режим analytic-service | `error` (всегда `Unavailable`) |
| `max_attempts` | 3 |
| `throttling_enabled` | true |
| `throttle_max_tokens` | 10 |
| `throttle_token_ratio` | 0.1 |
| Расчёт | `allowed_retry_rps = 10 × 0.1 = 1 retry/сек`; бюджет = 1 × 10 сек = **10 токенов** |

---

## 1. Почему важно ограничивать количество повторов при полном даунтайме?

При `mode=error` analytic-service **не восстановится** от retry — каждая повторная попытка гарантированно вернёт `Unavailable`. Без ограничений `profile-service` превращает каждый входящий запрос в `max_attempts` вызовов downstream:

```
100 RPS × 3 попытки = 300 RPS на уже «мёртвый» сервис
```

Это **retry storm**: мы не помогаем пользователю (ошибка всё равно придёт), но увеличиваем нагрузку на `analytic-service`, занимаем connection pool, CPU и ресурсы `profile-service` (goroutine, backoff-таймеры). Восстановление после реального инцидента затягивается — **золотое правило**: чем сильнее давим на страдающий сервис, тем дольше он встаёт.

Throttler ограничивает дополнительный RPS retry'ей, оставляя основной поток (первые попытки) и небольшой «бюджет» на случай кратковременного сбоя.

---

## 2. Что произойдёт, если лимит повторов не установлен?

**Сценарий:** analytic в `error`, `max_attempts=3`, throttler выключен.

- Каждый `GetProfile` делает до 3 вызовов `GetUserTaskCount`.
- При 10 RPS → **~30 RPC/сек** на analytic вместо 10 (×3 паразитная нагрузка).
- `profile-service` тратит время на backoff (30–100 ms) без пользы для клиента.
- Latency растёт: пользователь ждёт все попытки, прежде чем получить ошибку.
- Метрики: `retries_total` ≈ 2 × success_rps, `throttled_total = 0`, `success_total = 0`.

При масштабировании (100+ RPS) downstream получает кратный трафик и может не восстановиться даже после починки — classic **retry storm / death spiral**.

---

## 3. Какой ещё подход (паттерн) способен защитить проблемный сервис от перегрузки?

### Circuit Breaker

- При серии ошибок «открывается» и быстро отклоняет запросы без вызова downstream.
- Периодически пробует «полуоткрытое» состояние для проверки восстановления.
- Реализован в `fault-tolerant-stand` (`ad-service` → `review-service`).

### Другие паттерны

| Паттерн | Назначение |
|---|---|
| **Bulkhead** | Изоляция пулов/лимитов, чтобы сбой одной зависимости не утянул весь сервис |
| **Load shedding** | Отбрасывание части входящих запросов при перегрузке (README-2) |
| **Timeout + deadline** | Не дают запросу бесконечно ретраить и держать ресурсы |
| **Fallback / degraded response** | Вернуть профиль с `TariffUnknown` без вызова analytic |
| **Hedging** | Дублирование запроса (осторожно: удваивает нагрузку, нужен budget) |
| **Server pushback** | Downstream сам говорит «не ретраить» или «жди N ms» (`grpc-retry-pushback-ms`) |

Throttler и circuit breaker дополняют друг друга: throttler ограничивает retry-бюджет, breaker — полностью прекращает вызовы при устойчивом отказе.

---

## Анализ метрик (mode=error, throttler включён)

**Результат framer:** 100/100 ошибок (`ok=0`) — ожидаемо, analytic мёртв.

### Метрики (`:8092/metrics`)

| Метрика | Значение | Смысл |
|---|---|---|
| `grpc_client_attempts_total` | 202 | Итерации цикла retry (включая throttled) |
| `grpc_client_retries_total` | 4 | Retry, дошедшие до invoker |
| `grpc_client_throttled_total` | 98 | Retry отброшены throttler'ом |
| `grpc_client_success_total` | 0 | Успехов нет |

**Реальных RPC к analytic:** ≈ 100 (первые попытки) + 4 (retry) = **104 вызова**.  
Без throttler было бы ≈ 100 × `max_attempts` = **300 вызовов** — снижение **~3×**.

### Алгоритм throttler (homework-stand)

1. Бак стартует полным (`tokens = max_tokens`), порог `thresh = max / 2`.
2. Перед retry (`attempt > 0`): `Throttle()` вычитает токен; если `tokens ≤ thresh` — retry запрещён.
3. При успешном RPC: `tokens += token_ratio` (в режиме `error` пополнения нет).

### Подтверждённое поведение (`max_tokens=10`, `thresh=5`)

- Все 100 запросов делают первую попытку → 100 вызовов analytic.
- На retry глобальный бак позволил **4** дополнительных попытки, **98** — throttled.
- Итого **~104 RPC** вместо **~300** без throttler.

Пользователь получает ошибки (analytic недоступен — это ожидаемо), но паразитная нагрузка на downstream существенно ниже, чем без бюджета.
