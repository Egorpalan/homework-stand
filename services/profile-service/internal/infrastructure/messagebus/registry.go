package messagebus

import (
	"context"

	"profile-service/internal/infrastructure/adapter"
	"profile-service/internal/infrastructure/messagebus/subscriber"
	"profile-service/internal/infrastructure/messagebus/subscriber/scheme/user_tariff_invalidate"
	"profile-service/internal/pkg/closer"
)

type Registry struct {
	subscribers subscriber.Subscribers
	handler     *user_tariff_invalidate.Handler
}

func NewRegistry(adapters *adapter.Registry) *Registry {
	registry := &Registry{
		subscribers: subscriber.NewSubscribers(),
		handler:     user_tariff_invalidate.NewHandler(adapters.Tariff),
	}
	closer.Add(registry.subscribers.UserTariffInvalidate.Close)
	return registry
}

func (r *Registry) Run(ctx context.Context) {
	go r.subscribers.UserTariffInvalidate.Subscribe(ctx, r.handler.Handle)
}
