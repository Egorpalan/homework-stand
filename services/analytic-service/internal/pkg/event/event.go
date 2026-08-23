package event

import (
	"context"

	"analytic-service/internal/pkg/pipe"
)

// Raw бинарное представление содержимого
type Raw []byte

// Events список событий
type Events []Event

// Event событие системы
type Event struct {
	EntityID string
	Key, Body, Headers Raw
	Schema   string
}

// Add добавляет в буфер из контекста событие.
func Add(ctx context.Context, eventCallback pipe.Func[Events]) {
	Extract(ctx).Add(eventCallback)
}
