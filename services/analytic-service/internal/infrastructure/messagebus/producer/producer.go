package producer

import (
	"analytic-service/internal/pkg/msgbus/producer"
)

type Producers struct {
	TaskEventsRetry *producer.MessageProducer
}

func NewProducers() Producers {
	return Producers{
		TaskEventsRetry: producer.NewMessageProducer(),
	}
}
