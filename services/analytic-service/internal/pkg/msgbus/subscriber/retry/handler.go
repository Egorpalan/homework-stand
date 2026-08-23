package retry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"analytic-service/internal/pkg/event"

	"github.com/IBM/sarama"
)

var ErrHandler = errors.New("retry handler")

type Handler struct {
	backoff time.Duration
	topic   string
	flusher event.Flusher
}

func NewHandler(backoff time.Duration, topic string, flusher event.Flusher) *Handler {
	return &Handler{
		backoff: backoff,
		topic:   topic,
		flusher: flusher,
	}
}

func (h *Handler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
	if h == nil {
		return nil
	}

	buf, ctx := event.WithContext(ctx, h.flusher)
	event.Add(ctx, func(ctx context.Context, batch event.Events) (event.Events, error) {
		e, err := h.messageToEvent(message)
		if err != nil {
			return nil, fmt.Errorf("%w: event.Add: %v", ErrHandler, err)
		}
		return append(batch, e), nil
	})

	if err := buf.Flush(ctx); err != nil {
		return fmt.Errorf("%w: buf.Flush: %v", ErrHandler, err)
	}

	return nil
}

func (h *Handler) messageToEvent(message *sarama.ConsumerMessage) (event.Event, error) {
	headers, err := h.createHeaders(message)
	if err != nil {
		return event.Event{}, err
	}

	return event.Event{
		Key:     message.Key,
		Body:    message.Value,
		Headers: headers,
		Schema:  h.topic,
	}, nil
}

func (h *Handler) createHeaders(message *sarama.ConsumerMessage) ([]byte, error) {
	headers := make(map[string]string, len(message.Headers)+4)
	for _, header := range message.Headers {
		if header != nil {
			headers[string(header.Key)] = string(header.Value)
		}
	}

	headers["x-deliver-at"] = time.Now().Add(h.backoff).Format(time.RFC3339)
	headers["x-original-topic"] = message.Topic
	headers["x-original-partition"] = strconv.Itoa(int(message.Partition))
	headers["x-original-offset"] = strconv.Itoa(int(message.Offset))

	headersRaw, err := json.Marshal(headers)
	if err != nil {
		return nil, fmt.Errorf("%w: json.Marshal: %v", ErrHandler, err)
	}

	return headersRaw, nil
}
