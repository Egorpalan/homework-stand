package storage

import (
	"database/sql"
	"time"

	"analytic-service/internal/pkg/outbox/message"
)

type Message struct {
	ID          int64          `db:"id"`
	EntityID    string         `db:"entity_id"`
	Topic       string         `db:"topic"`
	Key         []byte         `db:"key"`
	Body        []byte         `db:"body"`
	Headers     []byte         `db:"headers"`
	Metadata    []byte         `db:"metadata"`
	CreatedAt   time.Time      `db:"created_at"`
	ProcessedAt sql.NullTime   `db:"processed_at"`
	ErrorsCount int32          `db:"errors_count"`
	ErrorDesc   sql.NullString `db:"error_desc"`
}

func (m *Message) Convert() *message.Message {
	opts := []message.Option{
		message.WithID(m.ID),
		message.WithHeaders(m.Headers),
		message.WithMetadata(m.Metadata),
		message.WithErrorsCount(m.ErrorsCount),
	}
	if m.ProcessedAt.Valid {
		opts = append(opts, message.WithProcessedAt(m.ProcessedAt.Time))
	}
	if m.ErrorDesc.Valid {
		opts = append(opts, message.WithErrDesc(m.ErrorDesc.String))
	}
	return message.NewMessage(m.EntityID, m.Topic, m.Key, m.Body, opts...)
}
