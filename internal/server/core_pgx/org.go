package core_pgx

import (
	"context"

	database "github.com/goodieshq/onus/internal/database/pgx"
	"github.com/goodieshq/onus/internal/server/core"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetOrganizationByUserID retrieves the organization associated with a user
func (c *CorePGX) GetOrganizationByUserID(ctx context.Context, userID uuid.UUID) (*core.Organization, error) {
	org, err := c.q.GetOrganizationByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return convertOrganization(&org), nil
}

// UpdateOrganizationName updates the name of an organization
// Should require a role >= admin
func (c *CorePGX) UpdateOrganization(ctx context.Context, orgID uuid.UUID, name *string) (*core.Organization, error) {
	var dbName pgtype.Text
	if name != nil {
		dbName.String = *name
		dbName.Valid = true
	}

	org, err := c.q.UpdateOrganization(ctx, database.UpdateOrganizationParams{
		OrganizationID: orgID,
		Name:           dbName,
	})
	if err != nil {
		return nil, err
	}

	return convertOrganization(&org), nil
}
