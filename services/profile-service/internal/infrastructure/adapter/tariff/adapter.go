package tariff

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"profile-service/config"
	"profile-service/internal/domain/entity"
	"profile-service/internal/domain/value_object"
	"profile-service/internal/infrastructure/adapter/tariff/payload"
	"profile-service/internal/pkg/cache"
	"profile-service/internal/pkg/ctxutil"
	"profile-service/internal/pkg/jitter"
	"profile-service/internal/pkg/perror"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type TaskCounter interface {
	GetUserTaskCount(ctx context.Context, userID int64) (uint64, error)
}

type Adapter struct {
	client *cache.Client[payload.Tariff, *payload.Tariff]
	origin TaskCounter
	group  singleflight.Group
}

func NewAdapter(client *cache.Client[payload.Tariff, *payload.Tariff], origin TaskCounter) *Adapter {
	return &Adapter{
		client: client,
		origin: origin,
	}
}

func (a *Adapter) GetTariff(ctx context.Context, userID int64) (value_object.Tariff, error) {
	key := payload.Key(userID)
	entry, err := a.client.Get(ctx, key)
	if err == nil {
		if entry.Value.Expired(config.Instance().Cache.SoftTTL) {
			go a.refreshSkipErr(ctxutil.Detach(ctx), userID)
		}
		return entry.Value.Domain(), nil
	}

	if !errors.Is(err, redis.Nil) {
		slog.Error("tariff cache get failed", "user_id", userID, "error", err.Error())
	}

	tariff, loadErr := a.loadOnce(ctx, userID)
	if loadErr != nil {
		slog.Error("tariff origin failed, falling back to unknown", "user_id", userID, "error", loadErr.Error())
		return value_object.TariffUnknown, nil
	}

	return tariff, nil
}

func (a *Adapter) Invalidate(ctx context.Context, userID int64) error {
	return a.client.Invalidate(ctx, payload.Key(userID), payload.GenerationKey(userID))
}

func (a *Adapter) loadOnce(ctx context.Context, userID int64) (value_object.Tariff, error) {
	key := strconv.FormatInt(userID, 10)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err, _ := a.group.Do(key, func() (any, error) {
		return a.refresh(ctx, userID)
	})
	if err != nil {
		a.group.Forget(key)
		return value_object.TariffUnknown, err
	}

	return result.(value_object.Tariff), nil
}

func (a *Adapter) refresh(ctx context.Context, userID int64) (value_object.Tariff, error) {
	gen, err := a.client.Generation(ctx, payload.GenerationKey(userID))
	if err != nil {
		return value_object.TariffUnknown, err
	}

	taskCount, err := a.origin.GetUserTaskCount(ctx, userID)
	if err != nil {
		return value_object.TariffUnknown, err
	}

	profile := &entity.Profile{}
	profile.WithTariff(taskCount)

	value := payload.Convert(userID, profile.Tariff, taskCount)
	ttl := jitter.Deterministic(
		payload.Key(userID),
		config.Instance().Cache.TTL,
		config.Instance().Cache.JitterPercentage,
	)

	body, err := value.MarshalJSON()
	if err != nil {
		return value_object.TariffUnknown, err
	}

	written, err := a.client.SetIfGeneration(ctx, payload.Key(userID), payload.GenerationKey(userID), body, gen, ttl)
	if err != nil {
		slog.Error("tariff cache write failed", "user_id", userID, "error", err.Error())
	} else if !written {
		slog.Info("skip stale tariff write after invalidation", "user_id", userID)
	}

	return profile.Tariff, nil
}

func (a *Adapter) refreshSkipErr(ctx context.Context, userID int64) {
	perror.LogWrap(errors.WithMessage(a.refreshInBackground(ctx, userID), fmt.Sprintf("refresh tariff %d", userID)))
}

func (a *Adapter) refreshInBackground(ctx context.Context, userID int64) error {
	_, err := a.loadOnce(ctx, userID)
	return err
}
