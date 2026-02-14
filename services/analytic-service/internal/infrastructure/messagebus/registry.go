package messagebus

import (
	"context"

	"analytic-service/internal/infrastructure/messagebus/subscriber"
	"analytic-service/internal/infrastructure/messagebus/subscriber/scheme/task_events"
	"analytic-service/internal/pkg/closer"
	eventrouter "analytic-service/internal/pkg/event-router"
)

type handlers struct {
	TaskEvents *task_events.MessageHandler
}

type Registry struct {
	handlers    handlers
	subscribers subscriber.Subscribers
}

func NewRegistry(router *eventrouter.EventRouter[string, []byte]) *Registry {
	registry := &Registry{
		subscribers: subscriber.NewSubscribers(),
		handlers: handlers{
			TaskEvents: task_events.NewMessageHandler(router),
		},
	}

	closer.Add(registry.subscribers.TaskEvents.Close)
	return registry
}

func (r *Registry) Run(ctx context.Context) {
	// Можно добавить полноценный multiplexer для событий этого топика, но пока сосредоточимся на одном
	go r.subscribers.TaskEvents.Subscribe(ctx, r.handlers.TaskEvents.Handle)
}
