package consumer

import (
	"log/slog"

	"github.com/IBM/sarama"
)

type groupSubscriber struct {
	messageHandler MessageHandler
}

func (g groupSubscriber) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (g groupSubscriber) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (g groupSubscriber) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			if err := g.messageHandler(ctx, session, message); err != nil {
				slog.Error("handle kafka message", "error", err.Error())
			}

			session.MarkMessage(message, "")
			session.Commit()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
