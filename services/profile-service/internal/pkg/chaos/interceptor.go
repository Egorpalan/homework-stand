package chaos

import (
	"context"
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

func ModeInterceptor(store *mode.Store) grpc.UnaryServerInterceptor {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		method := info.FullMethod
		slog.Info("method invoked", "method", method)

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
			resetSlow(method)
			return nil, status.Error(codes.Unavailable, "analytic in error mode")

		// ---------------- RARE ERROR ----------------

		case mode.RareError:
			resetSlow(method)
			if r.Intn(100) < 5 {
				return nil, status.Error(codes.Unavailable, "analytic rare error")
			}
			return handler(ctx, req)

		// ---------------- FLAKY ----------------

		case mode.Flaky:
			resetSlow(method)
			if r.Intn(100) < 15 {
				return nil, status.Error(codes.Unavailable, "analytic flaky error")
			}
			return handler(ctx, req)

		// ---------------- OK ----------------

		default:
			resetSlow(method)
			return handler(ctx, req)
		}
	}
}
