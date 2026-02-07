package chaos

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"profile-service/internal/pkg/chaos/mode"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var slowSince sync.Map // map[string]*atomic.Int64

func getSlowSince(method string) *atomic.Int64 {
	v, _ := slowSince.LoadOrStore(method, &atomic.Int64{})
	return v.(*atomic.Int64)
}

func resetSlow(method string) {
	if v, ok := slowSince.Load(method); ok {
		v.(*atomic.Int64).Store(0)
	}
}

var inflight sync.Map // map[string]*atomic.Int64

func getInflight(method string) *atomic.Int64 {
	v, _ := inflight.LoadOrStore(method, &atomic.Int64{})
	return v.(*atomic.Int64)
}

func resetInflight(method string) {
	if v, ok := inflight.Load(method); ok {
		v.(*atomic.Int64).Store(0)
	}
}

func resetState(method string) {
	resetSlow(method)
	resetInflight(method)
}

func ModeInterceptor(store *mode.Store) grpc.UnaryServerInterceptor {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := info.FullMethod

		md := store.Get(method)

		switch md {

		// ---------------- SLOW ----------------

		case mode.Slow:
			since := getSlowSince(method)

			now := time.Now().UnixNano()
			if since.Load() == 0 {
				since.Store(now)
			}

			elapsed := time.Duration(now - since.Load())
			steps := elapsed / (500 * time.Millisecond)

			delay := 200*time.Millisecond + steps*100*time.Millisecond

			select {
			case <-time.After(delay):
				return handler(ctx, req)
			case <-ctx.Done():
				return nil, status.Error(codes.DeadlineExceeded, "request deadline exceeded")
			}

		// ---------------- ERROR ----------------

		case mode.Error:
			resetState(method)
			return nil, status.Error(codes.Unavailable, "analytic in error mode")

		// ---------------- RARE ERROR ----------------

		case mode.RareError:
			resetState(method)
			if r.Intn(100) < 5 {
				return nil, status.Error(codes.Unavailable, "analytic rare error")
			}
			return handler(ctx, req)

		// ---------------- FLAKY ----------------

		case mode.Flaky:
			resetState(method)
			if r.Intn(100) < 15 {
				return nil, status.Error(codes.Unavailable, "analytic flaky error")
			}
			return handler(ctx, req)

		// ---------------- OK ----------------

		// ---------------- OVERLOAD ----------------
		case mode.Overload:
			counter := getInflight(method)

			cur := counter.Add(1)
			defer counter.Add(-1)

			slog.Info(fmt.Sprintf("Inflight: %d", cur))
			if cur < 10 {
				select {
				case <-time.After(time.Millisecond * 45):
					return handler(ctx, req)
				case <-ctx.Done():
					return nil, status.Error(codes.DeadlineExceeded, "overloaded: timeout")
				}
			}

			var extra time.Duration
			if cur >= 11 && cur <= 100 {
				if cur < 20 {
					extra = time.Duration(cur) * 5 * time.Millisecond
				} else {
					extra = time.Duration(cur) * 20 * time.Millisecond
				}

				select {
				case <-time.After(extra):
					return handler(ctx, req)
				case <-ctx.Done():
					return nil, status.Error(codes.DeadlineExceeded, "overloaded: timeout")
				}
			}

			select {
			case <-time.After(time.Second * 5):
				return nil, status.Error(codes.DeadlineExceeded, "overloaded: timeout")
			case <-ctx.Done():
				return nil, status.Error(codes.DeadlineExceeded, "overloaded: timeout")
			}
		default:
			resetSlow(method)
			return handler(ctx, req)
		}
	}
}
