package circuit

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCircuitIsOpen = errors.New("circuit is open")

	ErrTooManyRequests = errors.New("too many requests through circuit")
)

// errorCodeMap отображение строк -> codes.Code
var errorCodeMap = map[string]codes.Code{
	"Unknown":           codes.Unknown,
	"DeadlineExceeded":  codes.DeadlineExceeded,
	"ResourceExhausted": codes.ResourceExhausted,
	"Internal":          codes.Internal,
	"Unavailable":       codes.Unavailable,
}

// triggerOnError возвращает true, если ошибку нужно учитывать как сбой для CB.
// Отмена клиентом (Canceled) и чистый context.Canceled не считаются сбоем внешнего сервиса.
func triggerOnError(err error, failureCodeNames []string, codesSet map[codes.Code]struct{}) bool {

	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) {
		return false
	}

	st, ok := status.FromError(err)
	if ok {
		if st.Code() == codes.Canceled {
			return false
		}
	} else {
		// Не gRPC-ошибка: как в исходной логике — сбой только если нет списка кодов
		// (иначе нельзя корректно сопоставить с failure_codes).
		return len(failureCodeNames) == 0
	}

	code := st.Code()

	// Если фильтрация не задана — считаем всё ошибками
	if len(failureCodeNames) == 0 {
		return true
	}

	// Проверяем вхождение в список "плохих" кодов
	_, allowed := codesSet[code]
	return allowed
}
