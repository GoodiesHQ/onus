package core_pgx

import (
	"context"
	"database/sql"
	"errors"

	database "github.com/goodieshq/onus/internal/database/pgx"
	"github.com/goodieshq/onus/internal/server/core"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

// CorePGX is the Postgres implementation of the core.Core interface
type CorePGX struct {
	pool *pgxpool.Pool
	q    *database.Queries
}

// NewCorePGX creates a new CorePGX instance with a connection pool to the Postgres database
func NewCorePGX(ctx context.Context, connStr string) (*CorePGX, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &CorePGX{
		pool: pool,
		q:    database.New(pool),
	}, nil
}

// GetDB returns a standard sql.DB instance for compatibility with libraries that require it
func (c *CorePGX) GetDB() (*sql.DB, func() error, error) {
	db := stdlib.OpenDBFromPool(c.pool)
	return db, db.Close, nil
}

// convertOrganization converts a postgres database.Organization to a core.Organization
func convertOrganization(organization *database.Organization) *core.Organization {
	if organization == nil {
		return nil
	}
	return &core.Organization{
		ID:        organization.ID,
		Name:      organization.Name,
		Domain:    organization.Domain,
		CreatedAt: organization.CreatedAt,
	}
}

// convertUser converts a postgres database.User to a core.User
func convertUser(user *database.User) *core.User {
	if user == nil {
		return nil
	}
	return &core.User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}
}

// convertAuthSnapshot converts a postgres database.GetAuthSnapshotRow to a core.AuthSnapshot
func convertAuthSnapshot(snapshot *database.GetAuthSnapshotRow) *core.AuthSnapshot {
	if snapshot == nil {
		return nil
	}
	return &core.AuthSnapshot{
		UserID:             snapshot.UserID.String(),
		Email:              snapshot.Email,
		Name:               snapshot.Name,
		OrganizationID:     snapshot.OrganizationID.String(),
		OrganizationName:   snapshot.OrganizationName,
		OrganizationDomain: snapshot.OrganizationDomain,
		Role:               core.UserOrgRole(snapshot.Role),
		Disabled:           snapshot.Disabled,
	}
}

// convertUserOrganizationAssignment converts a postgres database.UserOrganizationAssignment to a core.UserOrganizationAssignment
func convertUserOrganizationAssignment(assignment *database.UserOrganizationAssignment) *core.UserOrganizationAssignment {
	if assignment == nil {
		return nil
	}

	return &core.UserOrganizationAssignment{
		UserID:         assignment.UserID,
		OrganizationID: assignment.OrganizationID,
		Role:           core.UserOrgRole(assignment.Role),
		DisabledAt:     assignment.DisabledAt,
		DisabledReason: assignment.DisabledReason,
		CreatedAt:      assignment.CreatedAt,
	}
}

// convertUserWithAssignment converts a postgres database.ListAllUsersWithAssignmentsByOrganizationIDRow to a core.UserWithAssignment
func convertUserWithAssignment(u *database.ListAllUsersWithAssignmentsByOrganizationIDRow) *core.UserWithAssignment {
	if u == nil {
		return nil
	}

	return &core.UserWithAssignment{
		UserID:         u.UserID,
		OrganizationID: u.OrganizationID,
		Email:          u.Email,
		Name:           u.Name,
		Role:           core.UserOrgRole(u.Role),
		DisabledAt:     u.DisabledAt,
		DisableReason:  u.DisabledReason,
	}
}

// convertTask converts a postgres database.Task to a core.Task
func convertTask(task *database.Task) *core.Task {
	if task == nil {
		return nil
	}

	return &core.Task{
		ID:             task.ID,
		Title:          task.Title,
		Description:    task.Description,
		Notes:          task.Notes,
		DueBy:          task.DueBy,
		Priority:       core.TaskPriority(task.Priority),
		Progress:       int(task.Progress),
		Status:         int(task.Status),
		Assignee:       task.AssigneeUserID,
		Assigner:       task.AssignerUserID,
		OrganizationID: task.OrganizationID,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}
}

// withTx executes the provided function within a database transaction.
func (c *CorePGX) withTx(
	ctx context.Context,
	fn func(ctx context.Context, q *database.Queries) error,
) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}

	qtx := c.q.WithTx(tx)

	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				log.Error().Err(err).Msg("failed to rollback transaction")
			}
		}
	}()

	if err := fn(ctx, qtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}
