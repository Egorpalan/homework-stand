package message

import (
	"time"

	"github.com/samber/lo"
)

type Message struct {
	id          int64
	entityID    string
	topic       string
	key         []byte
	body        []byte
	headers     []byte
	metadata    []byte
	createdAt   time.Time
	processedAt *time.Time
	errorsCount int32
	errDesc     *string
}

func NewMessage(entityID, topic string, key, body []byte, options ...Option) *Message {
	message := &Message{
		entityID:  entityID,
		topic:     topic,
		key:       key,
		body:      body,
		createdAt: time.Now(),
	}
	for _, opt := range options {
		opt(message)
	}
	return message
}

func (m *Message) GetID() int64          { return m.id }
func (m *Message) GetEntityID() string   { return m.entityID }
func (m *Message) GetTopic() string      { return m.topic }
func (m *Message) GetKey() []byte        { return m.key }
func (m *Message) GetBody() []byte       { return m.body }
func (m *Message) GetHeaders() []byte    { return m.headers }
func (m *Message) GetMetadata() []byte {
	return lo.Ternary(m.metadata == nil, []byte("{}"), m.metadata)
}
func (m *Message) GetCreatedAt() time.Time    { return m.createdAt }
func (m *Message) GetProcessedAt() *time.Time { return m.processedAt }
func (m *Message) GetErrorsCount() int32      { return m.errorsCount }
func (m *Message) GetErrorDesc() *string      { return m.errDesc }

func (m *Message) MarkAsProcessed() {
	m.processedAt = lo.ToPtr(time.Now().UTC())
	m.errorsCount = 0
	m.errDesc = nil
}

func (m *Message) SetError(err error) {
	m.errorsCount++
	m.errDesc = lo.ToPtr(err.Error())
}
