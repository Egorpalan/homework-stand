package tariff_invalidated

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"analytic-service/config"
	"analytic-service/internal/pkg/event"
	"analytic-service/internal/pkg/pipe"

	"github.com/gofrs/uuid"
)

const eventTypeTariffInvalidated = "TariffInvalidated"

type Event struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	UserID     int64     `json:"user_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func New(userID int64) pipe.Func[event.Events] {
	return func(_ context.Context, batch event.Events) (event.Events, error) {
		eventID, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}

		payload := Event{
			EventID:    eventID.String(),
			EventType:  eventTypeTariffInvalidated,
			UserID:     userID,
			OccurredAt: time.Now().UTC(),
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		headers, err := json.Marshal(map[string]string{
			"x-app-name":   "analytic-service",
			"x-event-type": eventTypeTariffInvalidated,
		})
		if err != nil {
			return nil, err
		}

		return append(batch, event.Event{
			EntityID: strconv.FormatInt(userID, 10),
			Key:      event.Raw(strconv.FormatInt(userID, 10)),
			Body:     body,
			Headers:  headers,
			Schema:   config.UserTariffInvalidateTopic,
		}), nil
	}
}
