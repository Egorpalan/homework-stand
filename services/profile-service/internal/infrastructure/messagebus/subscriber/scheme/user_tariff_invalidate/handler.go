package user_tariff_invalidate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
)

type Invalidator interface {
	Invalidate(ctx context.Context, userID int64) error
}

type Event struct {
	EventID    string `json:"event_id"`
	EventType  string `json:"event_type"`
	UserID     int64  `json:"user_id"`
	OccurredAt string `json:"occurred_at"`
}

type Handler struct {
	invalidator Invalidator
}

func NewHandler(invalidator Invalidator) *Handler {
	return &Handler{invalidator: invalidator}
}

func (h *Handler) Handle(ctx context.Context, _ sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) error {
	var event Event
	if err := json.Unmarshal(message.Value, &event); err != nil {
		slog.Error("invalid tariff invalidate payload", "error", err.Error())
		return nil
	}

	if event.UserID == 0 {
		slog.Error("tariff invalidate event without user_id")
		return nil
	}

	if err := h.invalidator.Invalidate(ctx, event.UserID); err != nil {
		return fmt.Errorf("invalidate tariff cache: %w", err)
	}

	slog.Info("tariff cache invalidated", "user_id", event.UserID, "event_id", event.EventID)
	return nil
}
