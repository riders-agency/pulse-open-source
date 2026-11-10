package providers

import (
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	providerruntime "riders.agency/pulse/api/internal/modules/providers/runtime"
	"riders.agency/pulse/api/internal/runtime/registry"
)

func RegisterRuntimeHandlers(reg *registry.Registry, pool *pgxpool.Pool, masterHEX string, client, maxClient *http.Client) error {
	documentBasicAuth, err := providerruntime.ParseDocumentBasicAuth(os.Getenv("PULSE_DOCUMENT_BASIC_AUTH_JSON"))
	if err != nil {
		return err
	}
	return providerruntime.RegisterHandlers(reg, &providerruntime.Runner{
		P:                 pool,
		MasterHEX:         masterHEX,
		HTTPClient:        client,
		MaxHTTPClient:     maxClient,
		DocumentBasicAuth: documentBasicAuth,
	})
}
