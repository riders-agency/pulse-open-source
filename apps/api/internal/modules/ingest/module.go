package ingest

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/executor"
	ingesthttp "riders.agency/pulse/api/internal/modules/ingest/transport/http"
	runscontracts "riders.agency/pulse/api/internal/modules/runs/contracts"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterHTTP(r chi.Router, deps httpserver.Dependencies) {
	MountHookHTTP(r, deps.Pool, deps.Engine, deps.MasterKey, deps.Runs)
}

func MountHookHTTP(
	r chi.Router,
	pool *pgxpool.Pool,
	engine *executor.Engine,
	masterKey string,
	receipts runscontracts.RunReceiptReader,
) {
	ingesthttp.MountHookRoutes(r, ingesthttp.Webhook{
		P: pool, E: engine, MasterHEX: masterKey, Receipts: receipts,
	})
}
