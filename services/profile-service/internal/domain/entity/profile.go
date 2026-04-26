package entity

import (
	"time"

	"profile-service/internal/domain/value_object"
)

type Profile struct {
	UserID        int64
	Name          string
	Email         string
	IsTaskAllowed bool
	CreatedAt     time.Time

	Tariff         value_object.Tariff
	TariffDegraded bool
}

// NewProfile конструктор для сущности
func NewProfile(userID int64, name, email string) *Profile {
	now := time.Now()
	return &Profile{
		UserID:        userID,
		Name:          name,
		Email:         email,
		IsTaskAllowed: true,
		CreatedAt:     now,
	}
}

func (p *Profile) AllowTask(value bool) {
	p.IsTaskAllowed = value
}

func (p *Profile) WithTariff(taskCount uint64, countOK bool) {
	if !countOK {
		p.Tariff = value_object.TariffBase
		p.TariffDegraded = true
		return
	}
	p.TariffDegraded = false
	switch {
	case taskCount >= 0 && taskCount <= 5:
		p.Tariff = value_object.TariffBase
	case taskCount > 5:
		p.Tariff = value_object.TariffMax
	default:
		p.Tariff = value_object.TariffUnknown
	}
}
