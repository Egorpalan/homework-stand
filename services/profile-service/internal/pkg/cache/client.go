package cache

import (
	"context"
	"time"

	pkgredis "profile-service/internal/pkg/connector/redis"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

var setIfGeneration = redis.NewScript(`
local gen = redis.call('GET', KEYS[2])
if not gen then
  gen = '0'
end
if tostring(gen) ~= tostring(ARGV[2]) then
  return 0
end
redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[3])
return 1
`)

var invalidateScript = redis.NewScript(`
redis.call('INCR', KEYS[2])
redis.call('DEL', KEYS[1])
return 1
`)

type Client[TValue Marshaller, TValuePtr UnMarshaller] struct {
	client *pkgredis.ShardedClient
}

func NewClient[TValue Marshaller, TValuePtr UnMarshaller](client *pkgredis.ShardedClient) *Client[TValue, TValuePtr] {
	return &Client[TValue, TValuePtr]{
		client: client,
	}
}

//  ------------------  GET/SET ------------------

func (c *Client[TValue, TValuePtr]) Get(ctx context.Context, key string) (Entry[TValue, TValuePtr], error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return Entry[TValue, TValuePtr]{}, errors.WithMessagef(err, "get entry for key %s", key)
	}
	return From[TValue, TValuePtr](key, value)
}

func (c *Client[TValue, TValuePtr]) Set(ctx context.Context, entry Entry[TValue, TValuePtr]) error {
	return c.client.Set(ctx, entry.Key, entry.marshall(), entry.TTL()).Err()
}

func (c *Client[TValue, TValuePtr]) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *Client[TValue, TValuePtr]) Generation(ctx context.Context, genKey string) (string, error) {
	gen, err := c.client.Get(ctx, genKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "0", nil
		}
		return "", err
	}
	return gen, nil
}

func (c *Client[TValue, TValuePtr]) SetIfGeneration(
	ctx context.Context,
	key, genKey string,
	value []byte,
	generation string,
	ttl time.Duration,
) (bool, error) {
	res, err := setIfGeneration.Run(ctx, c.client, []string{key, genKey}, value, generation, ttl.Milliseconds()).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func (c *Client[TValue, TValuePtr]) Invalidate(ctx context.Context, key, genKey string) error {
	return invalidateScript.Run(ctx, c.client, []string{key, genKey}).Err()
}

func (c *Client[TValue, TValuePtr]) MGet(ctx context.Context, keys ...string) (Entries[TValue, TValuePtr], []string, error) {
	var partErr *pkgredis.PartialError
	// получаем частичный результат из кеша
	cmd := c.client.PartialMGet(ctx, keys...)

	if cmd.Err() != nil && !errors.As(cmd.Err(), &partErr) {
		return nil, keys, cmd.Err()
	}

	values := cmd.Val()
	entries := make(Entries[TValue, TValuePtr], 0, len(values))
	missed := make([]string, 0)

	for i, raw := range values {
		key := keys[i]
		switch asserted := raw.(type) {
		case error, nil:
			missed = append(missed, key)
			continue
		case string:
			if len(asserted) == 0 {
				continue
			}

			entry, err := From[TValue, TValuePtr](key, []byte(asserted))
			if err != nil {
				missed = append(missed, key)
				continue
			}

			entries = append(entries, entry)
		}
	}

	return entries, missed, nil
}
