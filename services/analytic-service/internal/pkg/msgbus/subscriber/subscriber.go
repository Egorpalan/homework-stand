package subscriber

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"analytic-service/internal/pkg/connector/kafka"
	"analytic-service/internal/pkg/connector/kafka/consumer"
	"analytic-service/internal/pkg/msgbus/subscriber/retry"

	"github.com/IBM/sarama"
)

const deliverAtHeader = "x-deliver-at"

type MessageSubscriber struct {
	subscriber   *consumer.TopicConsumer
	retryHandler *retry.Handler
}

func NewMessageSubscriber(topic string) *MessageSubscriber {
	return &MessageSubscriber{subscriber: consumer.NewTopicConsumer(topic, kafka.MustConsumerGroup())}
}

func (m *MessageSubscriber) WithRetry(handler *retry.Handler) *MessageSubscriber {
	m.retryHandler = handler
	return m
}

func (m *MessageSubscriber) Subscribe(ctx context.Context, handler consumer.MessageHandler) {
	m.subscriber.Subscribe(ctx, m.handle(handler))
}

func (m *MessageSubscriber) Close() error {
	return m.subscriber.Close()
}

func (m *MessageSubscriber) Stop() {
	m.subscriber.Stop()
}

func (m *MessageSubscriber) Errors() <-chan error {
	return m.subscriber.Errors()
}

func (m *MessageSubscriber) handle(handler consumer.MessageHandler) consumer.MessageHandler {
	return func(ctx context.Context, session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) error {
		deliverAt := m.deliverAt(message)
		if !deliverAt.IsZero() {
			delay := time.Until(deliverAt)
			if delay > 0 {
				select {
				case <-session.Context().Done():
					return nil
				case <-time.After(delay):
				}
			}
		}

		err := handler(ctx, session, message)
		if err != nil && m.retryHandler != nil {
			if retryErr := m.retryHandler.Handle(ctx, message); retryErr != nil {
				return retryErr
			}
			return nil
		}

		return err
	}
}

func (m *MessageSubscriber) deliverAt(message *sarama.ConsumerMessage) time.Time {
	headerValue := ""
	for _, header := range message.Headers {
		if string(header.Key) == deliverAtHeader {
			headerValue = string(header.Value)
			break
		}
	}

	if headerValue == "" {
		return time.Time{}
	}

	deliverAt, err := time.Parse(time.RFC3339, headerValue)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to parse deliverAt: %s", err.Error()))
		return time.Time{}
	}

	return deliverAt
}
