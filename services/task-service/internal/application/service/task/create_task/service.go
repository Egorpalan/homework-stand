package create_task

import (
	"context"

	"task-service/internal/domain/entity"
	"task-service/internal/events/task_created"
	"task-service/internal/pkg/event"
	"task-service/internal/pkg/terror"
	"task-service/internal/pkg/transaction"
)

type Creator interface {
	CreateTask(ctx context.Context, task *entity.Task) error
}

type Category interface {
	GetCategory(_ context.Context, id string) (entity.Category, error)
}

type PermissionChecker interface {
	CheckPermission(ctx context.Context, userID int64) (bool, error)
}

type Service struct {
	creator    Creator
	category   Category
	permission PermissionChecker

	flusher event.Flusher
}

func NewService(creator Creator, category Category, permission PermissionChecker, flusher event.Flusher) *Service {
	return &Service{creator: creator, category: category, permission: permission, flusher: flusher}
}

func (s *Service) Create(ctx context.Context, request CreateTaskDTO) (*entity.Task, error) {
	// проверяем, может ли пользователь создать задачу
	allowed, err := s.permission.CheckPermission(ctx, request.UserID)
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, terror.NewBusinessErr("настройки пользователя запрещают создавать задачи")
	}

	category, err := s.category.GetCategory(ctx, request.CategoryID)
	if err != nil {
		return nil, err
	}

	task := entity.NewTask(
		request.UserID,
		request.CategoryID,
		request.Comment,
		request.ExecutionTime,
	)

	task.SetPrice(category.Price)

	// создаем буфер событий
	buf, ctx := event.WithContext(ctx, s.flusher)
	// событие о создании задачи
	event.Add(ctx, task_created.New(task))

	err = transaction.Exec(ctx, func(ctx context.Context) error {
		if err := s.creator.CreateTask(ctx, task); err != nil {
			return err
		}

		// отправляем событие о создании
		return buf.Flush(ctx)
	})
	if err != nil {
		return nil, err
	}

	return task, nil
}
