package runtime

import (
	"context"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

var clientActionAdapters = []clientActionAdapter{
	{typeKey: "clients.manual_payment.trigger", execute: executeManualPaymentAdapter},
	{typeKey: "clients.payment_due.trigger", execute: executePaymentDueAdapter},
	{typeKey: "clients.payment_cycle.collect", execute: executePaymentCycleCollectAdapter},
}

func executeManualPaymentAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeClientAction(ctx, req, runner.executeManualPaymentTrigger)
}

func executePaymentDueAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeClientAction(ctx, req, runner.executePaymentDueTrigger)
}

func executePaymentCycleCollectAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeClientAction(ctx, req, runner.executePaymentCycleCollect)
}
