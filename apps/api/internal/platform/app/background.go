package app

import "context"

// StartBackground starts the process-owned background components assembled by Build.
func (a *Application) StartBackground(ctx context.Context) {
	if a == nil {
		return
	}
	if a.RuntimeKernel != nil {
		go a.RuntimeKernel.Run(ctx)
	}
	if a.TelegramPoller != nil {
		go a.TelegramPoller.Run(ctx)
	}
}
