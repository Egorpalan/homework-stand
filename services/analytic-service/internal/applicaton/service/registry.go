package service

import (
	"analytic-service/internal/applicaton/service/task/accept_task"
	"analytic-service/internal/applicaton/service/task/get_task_count"
	"analytic-service/internal/infrastructure/storage"
	"analytic-service/internal/pkg/outbox"
)

type Registry struct {
	AcceptTask   *accept_task.Service
	GetTaskCount *get_task_count.Service
}

func NewRegistry(storage *storage.Registry, box *outbox.Outbox) *Registry {
	return &Registry{
		AcceptTask:   accept_task.NewService(storage.TaskStorage, box),
		GetTaskCount: get_task_count.NewService(storage.TaskStorage),
	}
}
