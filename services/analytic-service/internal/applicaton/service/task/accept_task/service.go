package accept_task

import (
	"context"

	"analytic-service/internal/domain/entity"
	"analytic-service/internal/events/tariff_invalidated"
	"analytic-service/internal/pkg/event"
	"analytic-service/internal/pkg/transaction"
)

type Creator interface {
	InsertTask(ctx context.Context, task *entity.Task) error
}

type Service struct {
	storage Creator
	flusher event.Flusher
}

func NewService(storage Creator, flusher event.Flusher) *Service {
	return &Service{storage: storage, flusher: flusher}
}

func (s *Service) Create(ctx context.Context, request CreateTaskRequest) error {
	task := entity.NewTask(
		request.taskID,
		request.userID,
		request.categoryID,
		request.status,
		request.comment,
		request.executionTime,
		request.createdAt,
		request.price,
	)

	buf, ctx := event.WithContext(ctx, s.flusher)
	event.Add(ctx, tariff_invalidated.New(request.userID))

	return transaction.Exec(ctx, func(ctx context.Context) error {
		if err := s.storage.InsertTask(ctx, task); err != nil {
			return err
		}
		return buf.Flush(ctx)
	})
}
