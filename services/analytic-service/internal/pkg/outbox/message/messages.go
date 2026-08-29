package message

import (
	"time"

	"analytic-service/internal/pkg/event"

	"github.com/samber/lo"
)

type Messages []*Message

func (b Messages) Convert() event.Events {
	return lo.Map(b, func(msg *Message, _ int) event.Event {
		return event.Event{
			EntityID: msg.GetEntityID(),
			Key:      msg.GetKey(),
			Body:     msg.GetBody(),
			Headers:  msg.GetHeaders(),
			Schema:   msg.GetTopic(),
		}
	})
}

func (b Messages) IsEmpty() bool { return len(b) == 0 }

func (b Messages) MarkAsProcessed() {
	for _, msg := range b {
		msg.MarkAsProcessed()
	}
}

func (b Messages) AnErrorOccurred(err error) {
	for _, m := range b {
		m.SetError(err)
	}
}

func (b Messages) IDs() []int64 {
	return lo.Map(b, func(msg *Message, _ int) int64 { return msg.GetID() })
}

func (b Messages) EntitiesIDs() []string {
	return lo.Map(b, func(msg *Message, _ int) string { return msg.GetEntityID() })
}

func (b Messages) Topics() []string {
	return lo.Map(b, func(msg *Message, _ int) string { return msg.GetTopic() })
}

func (b Messages) Keys() [][]byte {
	return lo.Map(b, func(msg *Message, _ int) []byte { return msg.GetKey() })
}

func (b Messages) Bodies() [][]byte {
	return lo.Map(b, func(msg *Message, _ int) []byte { return msg.GetBody() })
}

func (b Messages) Headers() []string {
	return lo.Map(b, func(msg *Message, _ int) string { return string(msg.GetHeaders()) })
}

func (b Messages) Metadata() []string {
	return lo.Map(b, func(msg *Message, _ int) string { return string(msg.GetMetadata()) })
}

func (b Messages) CreatedAt() []time.Time {
	return lo.Map(b, func(msg *Message, _ int) time.Time { return msg.GetCreatedAt() })
}

func (b Messages) ProcessedAt() []*time.Time {
	return lo.Map(b, func(msg *Message, _ int) *time.Time { return msg.GetProcessedAt() })
}

func (b Messages) ErrorsCounts() []int32 {
	return lo.Map(b, func(msg *Message, _ int) int32 { return msg.GetErrorsCount() })
}

func (b Messages) ErrorsDesc() []*string {
	return lo.Map(b, func(msg *Message, _ int) *string { return msg.GetErrorDesc() })
}
