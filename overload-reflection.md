# README-2: Overload и rate-limiter

**Контекст:** `task-service` → `profile-service`, метод `CheckPermission`

---

## Задание 1 — защита от перегрузки

### Контекст экспериментов

| Параметр | Значение |
|---|---|
| Нагрузка | `task-service` CreateTask → `CheckPermission` |
| Framer | 10 clients, `line 120 1000` (120→1000 RPS) |
| Режим | `overload` на `CheckPermission` |
| SLA | ≤ 400 ms на `CheckPermission` |

### Конфигурация rate-limiter

```yaml
rate_limit:
  rules:
    - clients: ["task-service"]
      handlers: ["/profile_service.order.v1.ProfileService/CheckPermission"]
      limit: 40
      burst: 60
      timeout: 280ms
```

> В `task-service` исправлен `AppName` на `"task-service"` (был `"profile-service"`), чтобы правило `clients: ["task-service"]` корректно матчилось.

### Сравнение: без rate-limiter vs с rate-limiter

| Метрика | Без limiter (15s) | С limiter `limit=40` (25s) |
|---|---|---|
| Успех CreateTask | 299 / 8390 (3.6%) | 1070 / 14000 (7.6%) |
| Ошибки | 8091 × `DeadlineExceeded` | 12930 × `ResourceExhausted` |
| Max inflight (логи profile) | ~4269 | ~7 |
| Avg latency успешных | ~128 ms | ~308 ms |
| Max latency успешных | — | ~355 ms (≤ 400 ms SLA ✓) |

### Без limiter

- При росте RPS inflight растёт лавинообразно (`overload`: inflight > 100 → `DeadlineExceeded`).
- Сервис «ломается» — очередь запросов, таймауты, деградация всего `profile-service`.

### С limiter

- Inflight удерживается на низком уровне — overload не уходит в зону > 100.
- Избыточные запросы быстро получают `ResourceExhausted` (не ждут 5s в overload).
- Успешные запросы проходят через `Wait` (до `timeout` 280 ms) + обработку (~45–120 ms).
- Больше «полезных» успехов, чем без limiter, при той же перегрузке.

### Выбор параметров

| Параметр | Значение | Обоснование |
|---|---|---|
| `limit` | 40 RPS | Ниже RPS деградации (~120+), но достаточно для burst |
| `burst` | 60 | Кратковременный пик без мгновенного отказа |
| `timeout` | 280 ms | Wait + обработка ≈ 355 ms max (факт: avg 308 ms, max 355 ms) |

**RPS деградации (без limiter):** наблюдается при ~120+ RPS — inflight начинает расти, latency растёт, при inflight > 100 — `DeadlineExceeded`.

---

## Задание 2 — экспериментальные вопросы

### 1. Что произойдёт, если `limit` слишком высоко?

Если `limit` установлен значительно выше реальной пропускной способности ручки в режиме `Overload`, limiter **перестаёт защищать** сервис — поведение приближается к сценарию без limiter.

**Наблюдение из эксперимента (фактически «limit = ∞»):**

- inflight вырос до **~4269**;
- массовые **`DeadlineExceeded`** (8091 из 8390 запросов);
- успешных запросов меньше (3.6%), чем с `limit=40` (7.6%).

**Почему так:** при высоком `limit` token bucket пропускает слишком много RPS. Inflight накапливается, chaos-режим `Overload` увеличивает задержку пропорционально inflight, и при > 100 запрос «умирает» по deadline. Limiter не выполняет свою задачу — сервис снова перегружен.

**Практический вывод:** `limit` должен быть **ниже RPS, при котором начинается деградация** (в нашем стенде — ниже ~120 RPS). Мы выбрали **40 RPS** — консервативно, с запасом до зоны `inflight > 10`.

---

### 2. Что произойдёт, если `burst` слишком велик?

`burst` — ёмкость token bucket: сколько запросов можно **мгновенно** пропустить сверх sustained rate, прежде чем сработает `limit`.

**Если `burst` слишком велик** (при том же `limit`):

- в начале нагрузки или при резком скачке RPS limiter **пропускает большую волну** запросов разом;
- inflight **скачет вверх** — ровно то, чего мы избегаем в `Overload`;
- даже при умеренном `limit=40` burst=500 позволил бы ~500 одновременных «входов» до начала отсечения;
- сервис на короткое время попадает в зону `inflight 20–100+` с растущей задержкой или `DeadlineExceeded`.

**Сравнение с нашим конфигом (`burst=60`, `limit=40`):**

- max inflight ≈ **7** — сервис остаётся в «зелёной» зоне (< 10, задержка ~45 ms);
- burst чуть выше limit даёт небольшой запас на микропики, но не открывает шлюз на сотни запросов.

**Практический вывод:** `burst` обычно ставят **чуть выше `limit`** (1.2–2×), но не на порядки больше. Цель — сгладить микропики, а не пропустить всю лавину.

---

### 3. Как `timeout` rate-limiter влияет на стабильность ручки?

`timeout` — максимальное время, которое запрос **ждёт токен** в `limiter.Wait()` перед тем, как получить `ResourceExhausted`.

**Сравнение из наших прогонов:**

| `timeout` | Avg latency (OK) | Max latency (OK) | Поведение |
|---|---|---|---|
| 350 ms | ~421 ms | ~994 ms | Часть успешных **выходит за SLA 400 ms** — запрос долго стоит в очереди за токеном |
| 280 ms | ~308 ms | ~355 ms | Укладываемся в SLA; лишние запросы быстрее получают отказ |

**Большой `timeout`:**

- (+) больше запросов дождутся токена и завершатся успешно;
- (−) растёт latency успешных запросов (очередь);
- (−) goroutine и inflight дольше «висят» в ожидании → хуже предсказуемость, риск выйти за SLA.

**Малый `timeout`:**

- (+) быстрый отказ (`ResourceExhausted`) — клиент не ждёт зря;
- (+) inflight не копится в очереди limiter'а;
- (−) больше ошибок на клиенте, даже когда «чуть-чуть» не хватило времени на токен.

**Практический вывод:** `timeout` нужно подбирать так, чтобы **Wait + обработка ≤ SLA** (400 ms). Формула: `timeout ≈ SLA − expected_handler_latency`. У нас: 400 − ~75 ms ≈ **280–300 ms**.

---

### 4. Можно ли использовать этот limiter для других ручек в сервисе?

**Да.** Реализация в `profile-service/internal/pkg/ratelimit` поддерживает **несколько правил** с фильтрацией по:

- `clients` — имя клиента из metadata (`x-client-name`);
- `handlers` — полное имя gRPC-метода.

Пример для нескольких ручек:

```yaml
rate_limit:
  rules:
    - clients: ["task-service"]
      handlers: ["/profile_service.order.v1.ProfileService/CheckPermission"]
      limit: 40
      burst: 60
      timeout: 280ms

    - clients: ["api-gateway"]
      handlers: ["/profile_service.order.v1.ProfileService/GetProfile"]
      limit: 100
      burst: 150
      timeout: 250ms

    - clients: []          # любой клиент
      handlers: ["/profile_service.order.v1.ProfileService/GetProfileList"]
      limit: 50
      burst: 80
      timeout: 300ms
```

**На что обратить внимание:**

- у разных ручек **разная стоимость** и разный SLA — параметры нужно подбирать отдельно;
- `GetProfile` ходит в analytic — limiter на входе profile-service защищает сам profile, но не analytic;
- пустой `clients: []` применяет правило ко **всем** клиентам — использовать осторожно;
- порядок правил: первое совпавшее правило побеждает (см. `match()` в `limiter.go`).

Limiter универсален для любого unary gRPC handler на сервере, где уже подключён `ExtractClientNameInterceptor`.

---

## Итог

| Задача | Решение |
|---|---|
| Защита от overload | Rate-limiter на `CheckPermission` для клиента `task-service` |
| SLA ≤ 400 ms | `limit=40`, `burst=60`, `timeout=280ms` |
| Контроль inflight | Max ~7 вместо ~4269 без limiter |
| Баланс ошибок | `ResourceExhausted` (быстро) вместо `DeadlineExceeded` (после долгого ожидания) |
