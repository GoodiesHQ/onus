package core_pgx

import (
	"context"
	"errors"
	"strings"

	database "github.com/goodieshq/onus/internal/database/pgx"
	"github.com/goodieshq/onus/internal/server/core"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ListEnabledUsers lists all enabled users in the organization, optionally filtered by a search string
func (c *CorePGX) ListEnabledUsers(ctx context.Context, orgID uuid.UUID, search string) ([]*core.User, error) {
	search = strings.TrimSpace(search)
	var dbSearch *string
	if search != "" {
		dbSearch = &search
	}

	users, err := c.q.ListEnabledUsersByOrganizationID(ctx, database.ListEnabledUsersByOrganizationIDParams{
		OrganizationID: orgID,
		Search:         dbSearch,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	if len(users) == 0 {
		return []*core.User{}, nil
	}
	result := make([]*core.User, 0, len(users))
	for _, u := range users {
		result = append(result, &core.User{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			CreatedAt: u.CreatedAt,
		})
	}
	return result, nil
}

// ListAllUsersWithAssignments lists all users in the organization along with their assignments
// Should require role >= admin
func (c *CorePGX) ListAllUsersWithAssignments(ctx context.Context, orgID uuid.UUID) ([]*core.UserWithAssignment, error) {
	users, err := c.q.ListAllUsersWithAssignmentsByOrganizationID(ctx, orgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	if len(users) == 0 {
		return []*core.UserWithAssignment{}, nil
	}

	results := make([]*core.UserWithAssignment, 0, len(users))
	for _, user := range users {
		results = append(results, convertUserWithAssignment(&user))
	}

	return results, nil
}

// EnableUser enables a disabled user in the organization
// Should require role >= admin
func (c *CorePGX) EnableUser(ctx context.Context, userID, orgID uuid.UUID) (*core.UserOrganizationAssignment, error) {
	assignment, err := c.q.EnableUser(ctx, database.EnableUserParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	return convertUserOrganizationAssignment(&assignment), nil
}

// DisableUser disables a user in the organization with a given reason
// Should require role >= admin
func (c *CorePGX) DisableUser(ctx context.Context, userID, orgID uuid.UUID, reason string) (*core.UserOrganizationAssignment, error) {
	assignment, err := c.q.DisableUser(ctx, database.DisableUserParams{
		UserID:         userID,
		OrganizationID: orgID,
		DisabledReason: reason,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	return convertUserOrganizationAssignment(&assignment), nil
}

// UpdateUserName updates the name of a user
func (c *CorePGX) UpdateUserName(ctx context.Context, userID uuid.UUID, name string) (*core.User, error) {
	user, err := c.q.UpdateUserName(ctx, database.UpdateUserNameParams{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	return convertUser(&user), nil
}

// UpdateUserRole updates the role of a user in the organization
// Should require role >= admin
func (c *CorePGX) UpdateUserRole(ctx context.Context, userID, orgID uuid.UUID, role core.UserOrgRole) (*core.UserOrganizationAssignment, error) {
	assignment, err := c.q.UpdateUserRole(ctx, database.UpdateUserRoleParams{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           int16(role),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	return convertUserOrganizationAssignment(&assignment), nil
}
