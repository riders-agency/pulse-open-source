package modelgateway

import (
	"context"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"riders.agency/pulse/api/internal/platform/demomode"
)

// Demo credentials are separate from the normal project/provider resolution.
func LoadDemoConfigFromEnv() Config {
	return Config{
		ProviderKey:      os.Getenv("PULSE_DEMO_MODEL_PROVIDER"),
		BaseURL:          os.Getenv("PULSE_DEMO_MODEL_BASE_URL"),
		APIKey:           os.Getenv("PULSE_DEMO_MODEL_API_KEY"),
		DefaultModelID:   os.Getenv("PULSE_DEMO_MODEL_MODEL"),
		AllowedModels:    parseAllowedModels(os.Getenv("PULSE_DEMO_MODEL_ALLOWED_MODELS")),
		DefaultMaxTokens: 700, HardMaxTokens: 1800, Timeout: 20 * time.Second,
	}.normalize()
}

func DemoProviderConfigured() bool {
	c := LoadDemoConfigFromEnv()
	return c.ProviderKey != "" && c.APIKey != "" && c.BaseURL != "" && c.DefaultModelID != ""
}

func WithDemoAccess(check func(context.Context, string, string) (bool, error), config Config) Option {
	return func(s *Server) { s.demoAccess = check; s.demoConfig = config.normalize() }
}

func (s *Server) providerForActor(ctx context.Context, userID, projectID, sessionID string) (Config, Provider, error) {
	if s.demoAccess != nil {
		demo, err := s.demoAccess(ctx, userID, projectID)
		if err != nil {
			return Config{}, nil, status.Error(codes.PermissionDenied, "model_access_denied")
		}
		if demo {
			if err := demomode.ReserveTenantModelCall(userID); err != nil {
				return Config{}, nil, status.Error(codes.ResourceExhausted, "demo_model_budget_exhausted")
			}
			p, err := providerForConfig(s.demoConfig, s.client)
			return s.demoConfig, p, err
		}
	}
	if err := demomode.ReserveModelCall(userID); err != nil {
		return Config{}, nil, status.Error(codes.ResourceExhausted, "demo_model_budget_exhausted")
	}
	return s.providerForRequest(ctx, projectID, sessionID)
}
