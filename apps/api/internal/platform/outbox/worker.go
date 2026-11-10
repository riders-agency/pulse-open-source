package outbox

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type Worker struct {
	Repo          *Repository
	Registry      *Registry
	Logger        *slog.Logger
	BatchSize     int
	LeaseDuration time.Duration
	PollInterval  time.Duration
}

func (w *Worker) Run(ctx context.Context) error {
	if w == nil || w.Repo == nil {
		return errors.New("outbox worker is not configured")
	}
	if w.Registry == nil {
		w.Registry = NewRegistry()
	}
	logger := w.Logger
	if logger == nil {
		logger = slog.Default()
	}
	batchSize := w.BatchSize
	if batchSize <= 0 {
		batchSize = 10
	}
	lease := w.LeaseDuration
	if lease <= 0 {
		lease = 60 * time.Second
	}
	poll := w.PollInterval
	if poll <= 0 {
		poll = 2 * time.Second
	}

	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	for {
		if err := w.runOnce(ctx, batchSize, lease, logger); err != nil {
			logger.Error("outbox worker batch failed", slog.Any("err", err))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (w *Worker) runOnce(ctx context.Context, batchSize int, lease time.Duration, logger *slog.Logger) error {
	events, err := w.Repo.Claim(ctx, batchSize, lease)
	if err != nil {
		return err
	}
	if len(events) > 0 {
		logger.Info("outbox events claimed",
			slog.Int("count", len(events)),
			slog.Duration("lease", lease),
		)
	}
	for _, ev := range events {
		w.handle(ctx, ev, logger)
	}
	return nil
}

func (w *Worker) handle(ctx context.Context, ev Event, logger *slog.Logger) {
	handler, ok := w.Registry.Handler(ev.TypeKey)
	if !ok {
		msg := "outbox handler is not registered"
		logger.Error(msg, slog.String("event_id", ev.ID), slog.String("event_type", ev.TypeKey))
		_ = w.Repo.MarkFailed(ctx, ev.ID, msg, retryAt(ev.Attempts))
		return
	}
	started := time.Now()
	if err := handler.Handle(ctx, ev); err != nil {
		logger.Error("outbox handler failed",
			slog.String("event_id", ev.ID),
			slog.String("event_type", ev.TypeKey),
			slog.Int("attempt", ev.Attempts),
			slog.Any("err", err),
		)
		_ = w.Repo.MarkFailed(ctx, ev.ID, err.Error(), retryAt(ev.Attempts))
		return
	}
	if err := w.Repo.MarkProcessed(ctx, ev.ID); err != nil {
		logger.Error("outbox mark processed failed", slog.String("event_id", ev.ID), slog.Any("err", err))
		return
	}
	logger.Info("outbox event processed",
		slog.String("event_id", ev.ID),
		slog.String("event_type", ev.TypeKey),
		slog.Duration("latency", time.Since(started)),
	)
}

func retryAt(attempt int) time.Time {
	if attempt < 1 {
		attempt = 1
	}
	delay := time.Duration(attempt*attempt) * time.Second
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}
	return time.Now().UTC().Add(delay)
}
