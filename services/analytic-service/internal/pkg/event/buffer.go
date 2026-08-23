package event

import (
	"context"
	"sync"

	"analytic-service/internal/pkg/pipe"
)

// Flusher функция для обработки содержимого буфера
type Flusher interface {
	Flush(ctx context.Context, events Events) error
}

// Buffer аккумулирует события
type Buffer struct {
	buffer          pipe.Pipe[Events]
	lock            *sync.Mutex
	flushCallback   Flusher
	flushCtxOptions []Option
}

// Option опция, применяемая к контексту перед вызовом Flusher
type Option func(ctx context.Context) context.Context

// New конструктор
func New(flusher Flusher, ctxOpts ...Option) *Buffer {
	return &Buffer{flushCallback: flusher, lock: &sync.Mutex{}, flushCtxOptions: ctxOpts}
}

// Add добавляет в буфер событие
func (b *Buffer) Add(eventCallback pipe.Func[Events]) {
	b.lock.Lock()
	b.buffer = b.buffer.With(eventCallback)
	b.lock.Unlock()
}

// Flush сбрасывает событие в функцию flusher для обработки и очищает буфер
func (b *Buffer) Flush(ctx context.Context) error {
	b.lock.Lock()
	events, err := b.buffer.Run(ctx, Events{}).Get()
	b.buffer = nil
	b.lock.Unlock()

	if err != nil {
		return err
	}

	for _, option := range b.flushCtxOptions {
		ctx = option(ctx)
	}

	return b.flushCallback.Flush(ctx, events)
}
