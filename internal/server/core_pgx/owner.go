package core_pgx

import (
	"context"
	"fmt"

	database "github.com/goodieshq/onus/internal/database/pgx"
	"github.com/google/uuid"
)

func (c *CorePGX) TransferOwnership(ctx context.Context, oldOwnerUserID, newOwnerUserID, orgID uuid.UUID) error {
	assignments, err := c.q.TransferOrganizationOwnership(ctx, database.TransferOrganizationOwnershipParams{
		OrganizationID: orgID,
		ActorUserID:    oldOwnerUserID,
		NewOwnerUserID: newOwnerUserID,
	})
	if err != nil {
		return err
	}

	if len(assignments) != 2 {
		return fmt.Errorf("unexpected number of assignments updated: %d", len(assignments))
	}

	return nil
}
