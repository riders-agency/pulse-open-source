package kernel

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultRuntimeWorkerCount = 8
	maximumRuntimeWorkerCount = 64
)

func RuntimeWorkerCountFromEnv(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 1 || value > maximumRuntimeWorkerCount {
		return defaultRuntimeWorkerCount
	}
	return value
}

func WithWorkerCount(count int) Option {
	return func(k *Kernel) {
		if count >= 1 && count <= maximumRuntimeWorkerCount {
			k.workerCount = count
		}
	}
}

func (k *Kernel) runWorkerPool(ctx context.Context) {
	workerCount := k.workerCount
	if workerCount < 1 || workerCount > maximumRuntimeWorkerCount {
		workerCount = defaultRuntimeWorkerCount
	}
	workerSignals := make([]chan struct{}, workerCount)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for index := 0; index < workerCount; index++ {
		workerSignals[index] = make(chan struct{}, 1)
		go func(workerID int, signal <-chan struct{}) {
			defer workers.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case <-signal:
					k.Drain(ctx)
				}
			}
		}(index+1, workerSignals[index])
	}
	slog.Info("runtime kernel worker pool started", slog.Int("workers", workerCount))
	defer func() {
		workers.Wait()
		slog.Info("runtime kernel worker pool stopped", slog.Int("workers", workerCount))
	}()

	processTick := time.NewTicker(k.processInterval)
	defer processTick.Stop()
	triggerTick := time.NewTicker(k.triggerInterval)
	defer triggerTick.Stop()
	signalRuntimeWorkers(workerSignals)
	for {
		select {
		case <-ctx.Done():
			return
		case <-processTick.C:
			signalRuntimeWorkers(workerSignals)
		case <-triggerTick.C:
			k.EnqueueDue(ctx)
			signalRuntimeWorkers(workerSignals)
		case <-k.ch:
			k.EnqueueDue(ctx)
			signalRuntimeWorkers(workerSignals)
		}
	}
}

func signalRuntimeWorkers(workerSignals []chan struct{}) {
	for _, signal := range workerSignals {
		select {
		case signal <- struct{}{}:
		default:
		}
	}
}
