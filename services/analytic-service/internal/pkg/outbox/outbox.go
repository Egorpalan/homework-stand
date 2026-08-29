package outbox

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"analytic-service/internal/pkg/event"
	"analytic-service/internal/pkg/outbox/message"
	"analytic-service/internal/pkg/outbox/storage"
	"analytic-service/internal/pkg/transaction"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
)

type Topic string

type Handler interface {
	Handle(ctx context.Context, events event.Events) error
	BatchSize(ctx context.Context) int
	Topic() string
}

type Outbox struct {
	storage  *storage.Storage
	handlers map[Topic]Handler
	config   Config
}

func NewOutbox(pool *pgxpool.Pool, config Config, opts ...Option) *Outbox {
	box := &Outbox{
		storage:  storage.NewStorage(pool),
		handlers: make(map[Topic]Handler),
		config:   config,
	}
	for _, opt := range opts {
		opt(box)
	}
	return box
}

func (s *Outbox) RegisterHandler(handler Handler) {
	s.handlers[Topic(handler.Topic())] = handler
}

func (s *Outbox) Flush(ctx context.Context, events event.Events) error {
	messages := make(message.Messages, 0, len(events))
	for _, ev := range events {
		messages = append(messages, message.NewMessage(
			ev.EntityID,
			ev.Schema,
			ev.Key,
			ev.Body,
			message.WithHeaders(json.RawMessage(ev.Headers)),
		))
	}
	if len(messages) == 0 {
		return nil
	}
	return s.storage.SaveMessages(ctx, messages)
}

func (s *Outbox) HandlePendingMessages(ctx context.Context) {
	_ = transaction.Exec(ctx, func(ctx context.Context) error {
		pending, err := s.storage.GetPendingMessages(ctx, s.config.LockedKeysLimit, s.config.MessagesLimit)
		if err != nil {
			slog.Error("ошибка получения pending сообщений", "error", err.Error())
			return err
		}
		if pending.IsEmpty() {
			return nil
		}

		var wg sync.WaitGroup
		for topic, messages := range lo.GroupBy(pending, func(m *message.Message) string { return m.GetTopic() }) {
			wg.Add(1)
			go func(topic string, messages message.Messages) {
				defer wg.Done()
				s.processMessagesByTopic(ctx, topic, messages)
			}(topic, messages)
		}
		wg.Wait()

		return s.storage.MarkAsProcessed(ctx, pending)
	})
}

func (s *Outbox) processMessagesByTopic(ctx context.Context, topic string, messages message.Messages) {
	handler := s.handlers[Topic(topic)]
	if handler == nil {
		slog.Error("outbox handler not registered", "topic", topic)
		return
	}

	for _, chunk := range lo.Chunk(messages, handler.BatchSize(ctx)) {
		if err := handler.Handle(ctx, chunk.Convert()); err != nil {
			chunk.AnErrorOccurred(err)
			if lo.ContainsBy(chunk, func(msg *message.Message) bool {
				return msg.GetErrorsCount() >= int32(s.config.MaxErrCountForMessage)
			}) {
				slog.Error("outbox message exceeded error limit", "topic", topic, "error", err.Error())
			}
			continue
		}
		chunk.MarkAsProcessed()
	}
}
