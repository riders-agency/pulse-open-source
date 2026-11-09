package providers

import "context"

type ModelGateway interface {
	FetchModels(context.Context, ModelDependency) ([]ModelDescriptor, error)
	Chat(context.Context, ModelDependency, string, []ChatMessage, int) (string, ModelUsage, error)
}
