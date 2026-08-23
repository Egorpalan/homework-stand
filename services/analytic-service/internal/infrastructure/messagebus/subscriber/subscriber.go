package subscriber

import (
	"time"

	"analytic-service/config"
	"analytic-service/internal/pkg/event"
	"analytic-service/internal/pkg/msgbus/subscriber"
	"analytic-service/internal/pkg/msgbus/subscriber/retry"
)

type Subscribers struct {
	TaskEvents      *subscriber.MessageSubscriber
	TaskEventsRetry *subscriber.MessageSubscriber
}

func NewSubscribers(flusher event.Flusher) Subscribers {
	const retryBackoff = 30 * time.Second

	taskEventsSubscriber := subscriber.NewMessageSubscriber(config.TaskEventsTopic).
		WithRetry(retry.NewHandler(retryBackoff, config.TaskEventsRetry30sTopic, flusher))

	taskEventsRetrySubscriber := subscriber.NewMessageSubscriber(config.TaskEventsRetry30sTopic).
		WithRetry(retry.NewHandler(retryBackoff, config.TaskEventsDLQTopic, flusher))

	return Subscribers{
		TaskEvents:      taskEventsSubscriber,
		TaskEventsRetry: taskEventsRetrySubscriber,
	}
}
