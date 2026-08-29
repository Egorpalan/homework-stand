package get_profile

import (
	"context"

	"profile-service/internal/domain/entity"
	"profile-service/internal/domain/value_object"
)

type ProfileProvider interface {
	GetProfile(ctx context.Context, userID int64) (*entity.Profile, error)
}

type TariffProvider interface {
	GetTariff(ctx context.Context, userID int64) (value_object.Tariff, error)
}

type Service struct {
	profileProvider ProfileProvider
	tariffProvider  TariffProvider
}

func NewService(profileProvider ProfileProvider, tariffProvider TariffProvider) *Service {
	return &Service{profileProvider: profileProvider, tariffProvider: tariffProvider}
}

func (s *Service) GetProfile(ctx context.Context, userID int64) (*entity.Profile, error) {
	profile, err := s.profileProvider.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	tariff, err := s.tariffProvider.GetTariff(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile.Tariff = tariff
	return profile, nil
}
