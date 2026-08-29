package payload

import (
	"encoding/json"
	"fmt"
	"time"

	"profile-service/internal/domain/value_object"
)

type Tariff struct {
	UserID    int64     `json:"user_id"`
	Tariff    int       `json:"tariff"`
	TaskCount uint64    `json:"task_count"`
	CachedAt  time.Time `json:"cached_at"`
}

func (t Tariff) MarshalJSON() ([]byte, error) {
	type alias Tariff
	return json.Marshal(alias(t))
}

func (t *Tariff) UnmarshalJSON(data []byte) error {
	type alias Tariff
	var parsed alias
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*t = Tariff(parsed)
	return nil
}

func (t *Tariff) Expired(ttl time.Duration) bool {
	return time.Since(t.CachedAt) >= ttl
}

func (t *Tariff) Domain() value_object.Tariff {
	return value_object.Tariff(t.Tariff)
}

func Convert(userID int64, tariff value_object.Tariff, taskCount uint64) Tariff {
	return Tariff{
		UserID:    userID,
		Tariff:    int(tariff),
		TaskCount: taskCount,
		CachedAt:  time.Now().UTC(),
	}
}

func Key(userID int64) string {
	return fmt.Sprintf("user:{%d}:tariff", userID)
}

func GenerationKey(userID int64) string {
	return fmt.Sprintf("user:{%d}:tariff:gen", userID)
}
