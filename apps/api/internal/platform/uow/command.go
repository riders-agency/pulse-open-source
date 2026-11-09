package uow

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
	"riders.agency/pulse/api/internal/platform/domainevent"
	"riders.agency/pulse/api/internal/platform/outbox"
)

type transactionKey struct{}
type transactionState struct {
	tx          pgx.Tx
	afterCommit []func()
}

// Run gives composed domain commands one atomic transaction. Only the outer command commits.
func Run(ctx context.Context, database dbpkg.DB, operation func(context.Context, pgx.Tx) error) error {
	if state, ok := ctx.Value(transactionKey{}).(*transactionState); ok {
		return operation(ctx, state.tx)
	}
	tx, err := database.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	state := &transactionState{tx: tx}
	events := &domainevent.Collector{}
	ctx = domainevent.WithCollector(ctx, events)
	if err := operation(context.WithValue(ctx, transactionKey{}, state), tx); err != nil {
		return err
	}
	for _, event := range events.Events {
		if _, err := outbox.EnqueueTx(ctx, tx, outbox.EnqueueRequest{TypeKey: event.TypeKey, AggregateType: event.AggregateType, AggregateID: event.AggregateID, IdempotencyKey: event.IdempotencyKey, PayloadJSON: event.PayloadJSON}); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	for _, notify := range state.afterCommit {
		notify()
	}
	return nil
}

// AfterCommit defers a wake-up until the data it references is durable.
func AfterCommit(ctx context.Context, notify func()) {
	if state, ok := ctx.Value(transactionKey{}).(*transactionState); ok {
		state.afterCommit = append(state.afterCommit, notify)
		return
	}
	notify()
}

// Savepoint keeps failed domain operations and their pending events out of the parent transaction.
func Savepoint(ctx context.Context, operation func(context.Context, pgx.Tx) error) error {
	parent, ok := ctx.Value(transactionKey{}).(*transactionState)
	if !ok {
		return fmt.Errorf("savepoint requires an application transaction")
	}
	tx, err := parent.tx.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	state := &transactionState{tx: tx}
	events := &domainevent.Collector{}
	child := domainevent.WithCollector(context.WithValue(ctx, transactionKey{}, state), events)
	if err := operation(child, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if err := domainevent.Merge(ctx, events.Events); err != nil {
		return err
	}
	parent.afterCommit = append(parent.afterCommit, state.afterCommit...)
	return nil
}
