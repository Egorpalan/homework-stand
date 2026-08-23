package service

import (
	"task-service/internal/application/service/task/create_task"
	"task-service/internal/application/service/task/get_categories"
	"task-service/internal/application/service/task/get_task"
	"task-service/internal/application/service/task/list_tasks"
	"task-service/internal/infrastructure/gateway"
	"task-service/internal/infrastructure/storage"
	"task-service/internal/pkg/outbox"
)

type Registry struct {
	CreateTask    *create_task.Service
	GetTask       *get_task.Service
	ListTasks     *list_tasks.Service
	GetCategories *get_categories.Service
}

func NewRegistry(storage *storage.Registry, gateway *gateway.Registry, outbox *outbox.Outbox) *Registry {
	return &Registry{
		CreateTask: create_task.NewService(
			storage.Task,
			storage.Category,
			gateway.Profile,
			outbox,
		),
		GetTask:       get_task.NewService(storage.Task),
		ListTasks:     list_tasks.NewService(storage.Task),
		GetCategories: get_categories.NewService(storage.Category),
	}
}
