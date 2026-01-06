package core

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid"
)

type UserOrgRole int

func (r UserOrgRole) IsValid() bool {
	switch r {
	case RoleMember, RoleAdmin, RoleOwner:
		return true
	default:
		return false
	}
}

func (r UserOrgRole) String() string {
	switch r {
	case RoleMember:
		return "member"
	case RoleAdmin:
		return "admin"
	case RoleOwner:
		return "owner"
	default:
		return "<none>"
	}
}

const (
	RoleNone   UserOrgRole = -1
	RoleMember UserOrgRole = 1
	RoleAdmin  UserOrgRole = 2
	RoleOwner  UserOrgRole = 3
)

type TaskPriority int

func (p TaskPriority) IsValid() bool {
	switch p {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityUrgent:
		return true
	default:
		return false
	}
}

const (
	TaskPriorityNone   TaskPriority = -1
	TaskPriorityLow    TaskPriority = 1
	TaskPriorityMedium TaskPriority = 2
	TaskPriorityHigh   TaskPriority = 3
	TaskPriorityUrgent TaskPriority = 4
)

type TaskStatus int

func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusToDo, TaskStatusInProgress, TaskStatusComplete:
		return true
	default:
		return false
	}
}

const (
	TaskStatusNone       TaskStatus = -1
	TaskStatusToDo       TaskStatus = 0
	TaskStatusInProgress TaskStatus = 1
	TaskStatusComplete   TaskStatus = 2
)

type TaskListScope string

func (s TaskListScope) IsValid() bool {
	switch s {
	case TaskListScopeAssigned, TaskListScopeRequested:
		return true
	default:
		return false
	}
}

const (
	TaskListScopeAssigned  TaskListScope = "assigned"
	TaskListScopeRequested TaskListScope = "requested"
)

type TaskListDirection int

const (
	TaskListDirectionNext TaskListDirection = iota
	TaskListDirectionPrev
)

type TaskListOrder string

func (o TaskListOrder) IsValid() bool {
	switch o {
	case TaskListOrderAsc, TaskListOrderDesc:
		return true
	default:
		return false
	}
}

const (
	TaskListOrderAsc  TaskListOrder = "asc"
	TaskListOrderDesc TaskListOrder = "desc"
)

// Common data structures used regardless of database implementation

type User struct {
	ID        uuid.UUID `json:"id"`         // unique identifier
	Email     string    `json:"email"`      // email address (must end with @ + organization domain name)
	Name      string    `json:"name"`       // full name of the user (for display purposes)
	CreatedAt time.Time `json:"created_at"` // when the user was created
}

type UserWithAssignment struct {
	UserID         uuid.UUID   `json:"user_id"`         // user unique identifier
	Name           string      `json:"name"`            // user full name
	Email          string      `json:"email"`           // user email address
	OrganizationID uuid.UUID   `json:"organization_id"` // organization unique identifier
	Role           UserOrgRole `json:"role"`            // role of the user in the organization ( member, admin, owner )
	DisabledAt     *time.Time  `json:"disabled_at"`     // timestamp when the user was disabled, nil if enabled
	DisableReason  *string     `json:"disabled_reason"` // reason for disabling the user, nil if enabled
}

type Task struct {
	ID             string       `json:"id"`              // unique identifier, ULID
	Title          string       `json:"title"`           // required title of the task
	Description    string       `json:"description"`     // optional detailed description
	Notes          string       `json:"notes"`           // optional notes for the task added during progress
	DueBy          *time.Time   `json:"due_by"`          // optional target date of task completion
	Priority       TaskPriority `json:"priority"`        // 1 = low, 2 = medium, 3 = high, 4 = urgent
	Progress       int          `json:"progress"`        // 0-100
	Status         int          `json:"status"`          // 0 = to do, 1 = in progress, 2 = done
	Assignee       uuid.UUID    `json:"assignee"`        // user who is assigned to complete the task
	Assigner       uuid.UUID    `json:"assigner"`        // user who assigned/created the task
	OrganizationID uuid.UUID    `json:"organization_id"` // organization to which the task belongs
	CreatedAt      time.Time    `json:"created_at"`      // when the task was created
	UpdatedAt      time.Time    `json:"updated_at"`      // when the task was last updated
}

type TaskCreateParams struct {
	Title          string        `json:"title"`                 // required title of the task
	Description    string        `json:"description,omitempty"` // optional detailed description
	DueBy          *time.Time    `json:"due_by,omitempty"`      // optional target date of task completion
	Priority       *TaskPriority `json:"priority,omitempty"`    // defaults to 2 (medium) if nil
	AssigneeUserID *uuid.UUID    `json:"assignee"`              // defauls to self if nil
	AssignerUserID uuid.UUID     `json:"assigner"`              // user who is creating the task
}

type TaskUpdateParamsManager struct {
	Title          *string       `json:"title,omitempty"`        // optional title of the task
	Description    *string       `json:"description,omitempty"`  // optional detailed description
	Notes          *string       `json:"notes,omitempty"`        // optional notes for the task
	ClearDueBy     *bool         `json:"clear_due_by,omitempty"` // if true, clears the due date
	DueBy          *time.Time    `json:"due_by,omitempty"`       // optional due date of task completion
	Progress       *int          `json:"progress,omitempty"`     // optional progress of the task
	Priority       *TaskPriority `json:"priority,omitempty"`     // optional priority of the task
	AssigneeUserID *uuid.UUID    `json:"assignee,omitempty"`     // optional user who is assigned to complete the task
}

type TaskUpdateParamsAssignee struct {
	Notes    *string `json:"notes,omitempty"`    // optional notes for the task
	Progress *int    `json:"progress,omitempty"` // optional progress of the task
}

type Organization struct {
	ID        uuid.UUID `json:"id"`         // unique identifier
	Name      string    `json:"name"`       // organization name
	Domain    string    `json:"domain"`     // organization domain name (for email restrictions)
	CreatedAt time.Time `json:"created_at"` // when the organization was created
}

type UserOrganizationAssignment struct {
	UserID         uuid.UUID   `json:"user_id"`         // user unique identifier
	OrganizationID uuid.UUID   `json:"organization_id"` // organization unique identifier
	Role           UserOrgRole `json:"role"`            // role of the user in the organization ( member, admin, owner )
	DisabledAt     *time.Time  `json:"disabled_at"`     // timestamp when the user was disabled, nil if enabled
	DisabledReason *string     `json:"disabled_reason"` // reason for disabling the user, nil if enabled
	CreatedAt      time.Time   `json:"created_at"`      // when the assignment was created
}

type AuthSnapshot struct {
	UserID             string      `json:"user_id"`             // user unique identifier
	Email              string      `json:"email"`               // user email address
	Name               string      `json:"name"`                // user full name
	OrganizationID     string      `json:"organization_id"`     // organization unique identifier
	OrganizationName   string      `json:"organization_name"`   // organization name
	OrganizationDomain string      `json:"organization_domain"` // organization domain name
	Role               UserOrgRole `json:"role"`                // role of the user in the organization
	Disabled           bool        `json:"disabled"`            // whether the user is disabled in the organization
}

type Core interface {
	// primitive database access
	GetDB() (*sql.DB, func() error, error)

	// Authentication and authrization
	SignupOrLogin(ctx context.Context, email, name string) (*User, error)
	GetAssignment(ctx context.Context, userID, orgID uuid.UUID) (UserOrgRole, bool, error)
	GetAuthSnapshot(ctx context.Context, userID, orgID uuid.UUID) (*AuthSnapshot, error)

	// Organization management
	GetOrganizationByUserID(ctx context.Context, userID uuid.UUID) (*Organization, error)
	UpdateOrganization(ctx context.Context, orgID uuid.UUID, name *string) (*Organization, error)

	// Task management
	ListTasks(ctx context.Context, userID, orgID uuid.UUID, scope TaskListScope, assignerID, assigneeID *uuid.UUID, since *time.Time, includeComplete *bool, pastDue *bool, priorityMin *TaskPriority, limit int) ([]*Task, error)
	GetTask(ctx context.Context, taskID ulid.ULID, orgID uuid.UUID) (*Task, error)
	CreateTask(ctx context.Context, orgID uuid.UUID, params *TaskCreateParams) (*Task, error)
	UpdateTaskAsManager(ctx context.Context, taskID ulid.ULID, orgID uuid.UUID, params *TaskUpdateParamsManager) (*Task, error)
	UpdateTaskAsAssignee(ctx context.Context, taskID ulid.ULID, orgID uuid.UUID, params *TaskUpdateParamsAssignee) (*Task, error)
	DeleteTaskByID(ctx context.Context, taskID ulid.ULID, orgID uuid.UUID) error

	// User management
	ListEnabledUsers(ctx context.Context, orgID uuid.UUID, search string) ([]*User, error)
	ListAllUsersWithAssignments(ctx context.Context, orgID uuid.UUID) ([]*UserWithAssignment, error)
	EnableUser(ctx context.Context, userID, orgID uuid.UUID) (*UserOrganizationAssignment, error)
	DisableUser(ctx context.Context, userID, orgID uuid.UUID, reason string) (*UserOrganizationAssignment, error)
	UpdateUserName(ctx context.Context, userID uuid.UUID, name string) (*User, error)
	UpdateUserRole(ctx context.Context, userID, orgID uuid.UUID, role UserOrgRole) (*UserOrganizationAssignment, error)
}
