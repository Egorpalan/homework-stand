package task_created

import (
	"context"
	"fmt"
	"log/slog"

	"analytic-service/internal/applicaton/service/task/accept_task"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

const eventTypeTaskCreated = "task-created"

type TaskCreator interface {
	Create(ctx context.Context, request accept_task.CreateTaskRequest) error
}

type Handler struct {
	creator TaskCreator
}

func NewMessageHandler(creator TaskCreator) *Handler {
	return &Handler{creator: creator}
}

// EventType возвращает тип ивента
func (h *Handler) EventType() string {
	return eventTypeTaskCreated
}

func (h *Handler) HandleEvent(ctx context.Context, payload []byte) error {
	deserialized, err := deserialize(payload)
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка десереализации сообщения: %s", err.Error()))
		return err
	}

	amount, err := decimal.NewFromString(deserialized.Price)
	if err != nil {
		return err
	}

	return h.creator.Create(ctx, accept_task.NewCreateTaskRequest(
		uuid.FromStringOrNil(deserialized.TaskID),
		deserialized.UserID,
		deserialized.CategoryID,
		deserialized.Status,
		deserialized.Comment,
		deserialized.ExecutionTime,
		deserialized.CreatedAt,
		amount,
	))
}
