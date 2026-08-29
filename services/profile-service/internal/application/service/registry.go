package service

import (
	"profile-service/internal/application/service/profile/check_permission"
	"profile-service/internal/application/service/profile/create_profile"
	"profile-service/internal/application/service/profile/get_profile"
	"profile-service/internal/application/service/profile/get_profiles"
	"profile-service/internal/infrastructure/adapter"
	"profile-service/internal/infrastructure/storage"
)

type Registry struct {
	CheckPermission *check_permission.Service
	GetProfile      *get_profile.Service
	GetProfiles     *get_profiles.Service
	CreateProfile   *create_profile.Service
}

func NewRegistry(storage *storage.Registry, adapter *adapter.Registry) *Registry {
	return &Registry{
		CheckPermission: check_permission.NewService(storage.Profile),
		GetProfile:      get_profile.NewService(storage.Profile, adapter.Tariff),
		GetProfiles:     get_profiles.NewService(adapter.Profile),
		CreateProfile:   create_profile.NewService(adapter.Profile),
	}
}
