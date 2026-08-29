package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Worker interface {
	Start(ctx context.Context) error
	Stop() error
}

type settings struct {
	interval             func(context.Context) time.Duration
	concurrentTasksCount func(context.Context) int
}

func NewWorker(
	ctx context.Context,
	task func(ctx context.Context),
	interval func(context.Context) time.Duration,
	concurrentTasksCount func(context.Context) int,
) Worker {
	return &worker{
		ctx:  ctx,
		task: task,
		settings: &settings{
			interval:             interval,
			concurrentTasksCount: concurrentTasksCount,
		},
	}
}

type worker struct {
	settings     *settings
	ctx          context.Context
	task         func(ctx context.Context)
	concurrentCh chan struct{}
	stopCh       chan struct{}
	stopChMutex  sync.Mutex
	taskWg       sync.WaitGroup
}

func (w *worker) Start(ctx context.Context) error {
	if w.stopCh != nil {
		return w.Stop()
	}

	w.concurrentCh = make(chan struct{}, w.settings.concurrentTasksCount(ctx))
	w.stopCh = make(chan struct{})

	go func() {
		for {
			select {
			case <-w.ctx.Done():
				return
			case <-w.stopCh:
				return
			case <-time.After(w.settings.interval(ctx)):
				if len(w.concurrentCh) < cap(w.concurrentCh) {
					w.concurrentCh <- struct{}{}
					w.taskWg.Add(1)
					go w.runTask()
				}
			}
		}
	}()

	return nil
}

func (w *worker) runTask() {
	defer func() {
		<-w.concurrentCh
		w.taskWg.Done()
	}()
	defer func() {
		if r := recover(); r != nil {
			slog.Error(fmt.Sprintf("recovered from panic in worker: %v", r))
		}
	}()
	w.task(w.ctx)
}

func (w *worker) Stop() error {
	w.stopChMutex.Lock()
	defer w.stopChMutex.Unlock()

	if w.stopCh == nil {
		return nil
	}

	close(w.stopCh)
	w.taskWg.Wait()
	w.stopCh = nil
	return nil
}
