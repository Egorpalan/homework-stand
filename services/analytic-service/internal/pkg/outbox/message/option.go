package message

import (
	"encoding/json"
	"time"
)

type Option func(opt *Message)

func WithID(id int64) Option {
	return func(message *Message) {
		message.id = id
	}
}

func WithHeaders(headers json.RawMessage) Option {
	return func(opt *Message) {
		opt.headers = headers
	}
}

func WithMetadata(meta json.RawMessage) Option {
	return func(message *Message) {
		message.metadata = meta
	}
}

func WithProcessedAt(processedAt time.Time) Option {
	return func(message *Message) {
		message.processedAt = &processedAt
	}
}

func WithErrorsCount(errorsCount int32) Option {
	return func(message *Message) {
		message.errorsCount = errorsCount
	}
}

func WithErrDesc(errorsDesc string) Option {
	return func(message *Message) {
		message.errDesc = &errorsDesc
	}
}
