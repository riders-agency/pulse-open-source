package runtime

import (
	"context"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

func executeYClientsCertificateAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeProviderAction(ctx, req, runner.executeYClientsCertificate)
}
