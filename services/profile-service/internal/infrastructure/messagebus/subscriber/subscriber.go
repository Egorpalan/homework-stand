package subscriber

import (
	"profile-service/config"
	"profile-service/internal/pkg/msgbus/subscriber"
)

type Subscribers struct {
	UserTariffInvalidate *subscriber.MessageSubscriber
}

func NewSubscribers() Subscribers {
	return Subscribers{
		UserTariffInvalidate: subscriber.NewMessageSubscriber(config.UserTariffInvalidateTopic),
	}
}
