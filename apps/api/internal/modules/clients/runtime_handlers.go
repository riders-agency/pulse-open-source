package clients

import (
	"github.com/jackc/pgx/v5/pgxpool"

	clientsruntime "riders.agency/pulse/api/internal/modules/clients/runtime"
	"riders.agency/pulse/api/internal/runtime/registry"
)

func RegisterRuntimeHandlers(reg *registry.Registry, pool *pgxpool.Pool, notifier clientsruntime.Notifier) error {
	return clientsruntime.RegisterHandlers(reg, pool, notifier)
}
