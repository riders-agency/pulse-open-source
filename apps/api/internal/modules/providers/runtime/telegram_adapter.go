package runtime

import (
	"context"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

func executeTelegramSendMessageAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeProviderAction(ctx, req, runner.executeTelegram)
}

func executeTelegramEditMessageTextAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeProviderAction(ctx, req, runner.executeTelegramEditMessageText)
}

func executeTelegramAnswerCallbackQueryAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeProviderAction(ctx, req, runner.executeTelegramAnswerCallbackQuery)
}

func executeTelegramPinMessageAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeProviderAction(ctx, req, runner.executeTelegramPinMessage)
}

func executeTelegramUnpinMessageAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeProviderAction(ctx, req, runner.executeTelegramUnpinMessage)
}

func executeTelegramSendDocumentAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeProviderAction(ctx, req, runner.executeTelegramDocument)
}
