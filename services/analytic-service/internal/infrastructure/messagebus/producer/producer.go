package producer

import (
	"analytic-service/internal/pkg/msgbus/producer"
)

type Producers struct {
	UserTariffInvalidate *producer.MessageProducer
}

func NewProducers() Producers {
	return Producers{
		UserTariffInvalidate: producer.NewMessageProducer(),
	}
}
