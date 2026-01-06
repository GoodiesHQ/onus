package core_pgx

import (
	"context"
	"errors"
	"fmt"

	database "github.com/goodieshq/onus/internal/database/pgx"
	"github.com/goodieshq/onus/internal/server/core"
	"github.com/goodieshq/onus/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Authenticate the user by creating an organization/account or logging them in if their email already exists
func (c *CorePGX) SignupOrLogin(ctx context.Context, email, name string) (*core.User, error) {
	// Extract normalized domain from the email address
	domain, err := util.ExtractDomain(email)
	if err != nil {
		return nil, err
	}

	var user database.User

	// wrap in a single transaction
	err = c.withTx(ctx, func(ctx context.Context, q *database.Queries) error {
		u, err := q.UpsertUser(ctx, database.UpsertUserParams{
			Email: email,
			Name:  name,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert user: %w", err)
		}

		org, err := q.UpsertOrganization(ctx, database.UpsertOrganizationParams{
			Name:   domain, // default placeholder, editable later
			Domain: domain,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert organization: %w", err)
		}

		// Attempt to claim owner, no-op if owner already exists
		if err := q.AssignUserToOrganizationTryOwner(ctx, database.AssignUserToOrganizationTryOwnerParams{
			UserID:         u.ID,
			OrganizationID: org.ID,
		}); err != nil {
			return fmt.Errorf("failed to assign owner to organization: %w", err)
		}

		// Ensure baseline membership exists (no-op if already a member)
		if err := q.AssignUserToOrganization(ctx, database.AssignUserToOrganizationParams{
			UserID:         u.ID,
			OrganizationID: org.ID,
			Role:           int16(core.RoleMember),
		}); err != nil {
			return fmt.Errorf("failed to assign user to organization: %w", err)
		}

		user = u
		return nil
	})
	if err != nil {
		return nil, err
	}

	return convertUser(&user), nil
}

// GetAuthSnapshot retrieves an authentication snapshot for a user in an organization
func (c *CorePGX) GetAuthSnapshot(ctx context.Context, userID, orgID uuid.UUID) (*core.AuthSnapshot, error) {
	snapshot, err := c.q.GetAuthSnapshot(ctx, database.GetAuthSnapshotParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return convertAuthSnapshot(&snapshot), nil
}

// GetAssignment retrieves the role and disabled status of a user in an organization
func (c *CorePGX) GetAssignment(ctx context.Context, userID, orgID uuid.UUID) (core.UserOrgRole, bool, error) {
	row, err := c.q.GetOrgUserRole(ctx, database.GetOrgUserRoleParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core.RoleNone, false, core.ErrNotFound
		}
		return core.RoleNone, false, err
	}
	return core.UserOrgRole(row.Role), row.Disabled, nil
}
