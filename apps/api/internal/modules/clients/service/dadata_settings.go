package service

import (
	"context"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/modules/action_dependencies/contracts"
	depssvc "riders.agency/pulse/api/internal/modules/action_dependencies/service"
)

func (c *Commands) GetDaDataIntegration(ctx context.Context, companyID, userID uuid.UUID) (*contracts.DaDataIntegrationRow, error) {
	return depssvc.New(c.P, c.MasterHEX).GetDaDataIntegration(ctx, companyID, userID)
}
func (c *Commands) GetDaDataProjectIntegration(ctx context.Context, projectID, userID uuid.UUID) (*contracts.DaDataIntegrationRow, error) {
	return depssvc.New(c.P, c.MasterHEX).GetDaDataProjectIntegration(ctx, projectID, userID)
}
