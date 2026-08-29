package jitter

import (
	"hash/fnv"
	"math/rand"
	"time"
)

func Deterministic(key string, baseTTL time.Duration, jitterPercent float64) time.Duration {
	if baseTTL <= 0 || jitterPercent <= 0 {
		return baseTTL
	}

	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	r := rand.New(rand.NewSource(int64(h.Sum64())))

	jitterRange := float64(baseTTL) * jitterPercent
	return baseTTL + time.Duration(r.Float64()*jitterRange)
}
