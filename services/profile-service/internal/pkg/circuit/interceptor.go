package circuit

import (
	"context"
	"errors"
	"log/slog"

	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc"

	"profile-service/internal/pkg/circuit/metrics"
)

// UnaryClientInterceptor унарный клиентский перехватчик
func (b *Breaker) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		state, ok := b.getBreakerByMethod(method)
		if !ok || state == nil {
			slog.Debug("no breaker found for method or disabled", "method", method)
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		_, err := state.breaker.Execute(func() (interface{}, error) {
			return nil, invoker(ctx, method, req, reply, cc, opts...)
		})

		if err != nil {
			switch {
			case errors.Is(err, gobreaker.ErrOpenState):
				metrics.IncRequestOutcome(method, "circuit_open")
				return ErrCircuitIsOpen
			case errors.Is(err, gobreaker.ErrTooManyRequests):
				metrics.IncRequestOutcome(method, "too_many_requests")
				return ErrTooManyRequests
			default:
				metrics.IncRequestOutcome(method, "upstream_error")
				return err
			}
		}
		metrics.IncRequestOutcome(method, "success")
		return nil
	}
}
