package adapter

import (
	"profile-service/internal/infrastructure/adapter/profile"
	"profile-service/internal/infrastructure/adapter/profile/payload"
	"profile-service/internal/infrastructure/adapter/tariff"
	tariffpayload "profile-service/internal/infrastructure/adapter/tariff/payload"
	"profile-service/internal/infrastructure/gateway"
	"profile-service/internal/infrastructure/storage"
	"profile-service/internal/pkg/cache"
	pkgredis "profile-service/internal/pkg/connector/redis"
)

type Registry struct {
	Profile *profile.Adapter
	Tariff  *tariff.Adapter
}

func NewRegistry(redis *pkgredis.ShardedClient, dal *storage.Registry, gw *gateway.Registry) *Registry {
	return &Registry{
		Profile: profile.NewAdapter(cache.NewClient[payload.Profile, *payload.Profile](redis), dal.Profile),
		Tariff:  tariff.NewAdapter(cache.NewClient[tariffpayload.Tariff, *tariffpayload.Tariff](redis), gw.Analytic),
	}
}
