package service

import (
	"context"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/modules/action_dependencies/probes"
)

func (c *Commands) TestConnection(ctx context.Context, pid, depID, uid uuid.UUID) (map[string]any, error) {
	dep, err := c.Deps.LoadConnectionProbe(ctx, pid, depID, uid)
	if err != nil {
		return nil, err
	}
	result := probes.New().Check(ctx, dep)
	if err := c.Deps.MarkConnectionCheck(ctx, pid, depID); err != nil {
		return nil, err
	}
	return result, nil
}
