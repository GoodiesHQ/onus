package core_pgx

import (
	"context"
	"errors"
	"fmt"
	"time"

	database "github.com/goodieshq/onus/internal/database/pgx"
	"github.com/goodieshq/onus/internal/server/core"
	"github.com/goodieshq/onus/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid"
)

// ListTasks retrieves tasks based on the provided filters and scope
func (c *CorePGX) ListTasks(ctx context.Context, userID, orgID uuid.UUID, scope core.TaskListScope, assignerID, assigneeID *uuid.UUID, since *time.Time, includeComplete *bool, pastDue *bool, priorityMin *core.TaskPriority, limit int) ([]*core.Task, error) {
	var dbPriorityMin *int16
	if priorityMin != nil {
		dbPriorityMin = util.Ptr(int16(*priorityMin))
	}

	params := database.ListTasksParams{
		OrganizationID:  orgID,
		UserID:          userID,
		Scope:           string(scope),
		PriorityMin:     dbPriorityMin,
		PastDue:         pastDue,
		Since:           since,
		Limit:           int32(limit),
		IncludeComplete: includeComplete,
		AssignerID:      assignerID,
		AssigneeID:      assigneeID,
	}
	tasks, err := c.q.ListTasks(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(tasks) == 0 {
		return []*core.Task{}, nil
	}

	result := make([]*core.Task, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, convertTask(&t))
	}
	return result, nil
}

// GetTask retrieves a task by its ID and organization ID
func (c *CorePGX) GetTask(ctx context.Context, taskID ulid.ULID, orgID uuid.UUID) (*core.Task, error) {
	task, err := c.q.GetTaskByID(ctx, database.GetTaskByIDParams{
		ID:             taskID.String(),
		OrganizationID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return convertTask(&task), nil
}

// UpdateTaskAsAssignee updates a task's notes and progress as the assignee
func (c *CorePGX) UpdateTaskAsAssignee(ctx context.Context, taskID ulid.ULID, orgID uuid.UUID, params *core.TaskUpdateParamsAssignee) (*core.Task, error) {
	var dbProgress *int16
	if params.Progress != nil {
		dbProgress = util.Ptr(int16(*params.Progress))
	}

	task, err := c.q.UpdateTaskAsAssignee(ctx, database.UpdateTaskAsAssigneeParams{
		TaskID:         taskID.String(),
		OrganizationID: orgID,
		Notes:          params.Notes,
		Progress:       dbProgress,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return convertTask(&task), nil
}

// UpdateTaskAsManager updates a task's details as a manager (role >= admin or task assigner)
func (c *CorePGX) UpdateTaskAsManager(ctx context.Context, taskID ulid.ULID, orgID uuid.UUID, params *core.TaskUpdateParamsManager) (*core.Task, error) {
	var clearDueBy = false
	if params.ClearDueBy != nil {
		clearDueBy = *params.ClearDueBy
	}

	var dbPriority, dbProgress *int16

	if params.Priority != nil {
		dbPriority = util.Ptr(int16(*params.Priority))
	}

	if params.Progress != nil {
		dbProgress = util.Ptr(int16(*params.Progress))
	}

	task, err := c.q.UpdateTaskAsManager(ctx, database.UpdateTaskAsManagerParams{
		ID:             taskID.String(),
		OrganizationID: orgID,
		Title:          params.Title,
		Description:    params.Description,
		Notes:          params.Notes,
		ClearDueBy:     clearDueBy,
		DueBy:          params.DueBy,
		Priority:       dbPriority,
		Progress:       dbProgress,
		AssigneeUserID: params.AssigneeUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return convertTask(&task), nil
}

// CreateTask creates a new task in the organization, requires assigner and assignee IDs
func (c *CorePGX) CreateTask(ctx context.Context, orgID uuid.UUID, params *core.TaskCreateParams) (*core.Task, error) {
	ulid, err := util.GenerateULID()
	if err != nil {
		return nil, err
	}

	var priority int16
	if params.Priority != nil {
		priority = int16(*params.Priority)
	} else {
		// default priority
		priority = int16(core.TaskPriorityMedium)
	}

	task, err := c.q.CreateTask(ctx, database.CreateTaskParams{
		ID:             ulid,
		Title:          params.Title,
		Description:    params.Description,
		Priority:       priority,
		DueBy:          params.DueBy,
		AssigneeUserID: params.AssigneeUserID,
		AssignerUserID: params.AssignerUserID,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create task: %w", err)
	}

	return convertTask(&task), nil
}

// DeleteTaskByID deletes a task by its ID and organization ID
func (c *CorePGX) DeleteTaskByID(ctx context.Context, taskID ulid.ULID, orgID uuid.UUID) error {
	err := c.q.DeleteTask(ctx, database.DeleteTaskParams{
		ID:             taskID.String(),
		OrganizationID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core.ErrNotFound
		}
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}
