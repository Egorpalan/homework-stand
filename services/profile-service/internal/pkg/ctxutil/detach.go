package ctxutil

import "context"

func Detach(ctx context.Context) context.Context {
	return context.WithoutCancel(ctx)
}
