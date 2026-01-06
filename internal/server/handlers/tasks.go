package handlers

import (
	"net/http"
	"time"

	"github.com/goodieshq/onus/internal/server/core"
	"github.com/goodieshq/onus/internal/server/session"
	"github.com/goodieshq/onus/internal/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// RequestTaskList represents the query parameters for listing tasks
type RequestTaskList struct {
	Scope           core.TaskListScope `schema:"scope"`
	Since           *time.Time         `schema:"since"`
	PriorityMin     *core.TaskPriority `schema:"priority_min"`
	IncludeComplete *bool              `schema:"include_complete"`
	PastDue         *bool              `schema:"past_due"`
	Limit           *int               `schema:"limit"`
	AssigneeID      *uuid.UUID         `schema:"assignee_id"`
	AssignerID      *uuid.UUID         `schema:"assigner_id"`
}

const limitMax = 1000

func (h *OnusHandler) GetApiTasks(w http.ResponseWriter, r *http.Request) {
	var req RequestTaskList
	if err := util.DecodeQueryParams(r, &req); err != nil {
		log.Error().Err(err).Msg("failed to decode query params")
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	if !req.Scope.IsValid() {
		log.Error().Msg("invalid scope value, '" + string(req.Scope) + "'")
		http.Error(w, "Invalid scope value", http.StatusBadRequest)
		return
	}

	if req.PriorityMin != nil {
		if !req.PriorityMin.IsValid() {
			http.Error(w, "Invalid priority_min value", http.StatusBadRequest)
			return
		}
	}

	var limit int
	if req.Limit == nil || *req.Limit <= 0 {
		limit = limitMax
	}
	if limit > limitMax {
		limit = limitMax
	}

	ctx := r.Context()

	// get organization and user IDs from context
	oid := session.CtxGetOID(ctx)
	uid := session.CtxGetUID(ctx)

	var assignerID, assigneeID *uuid.UUID

	switch req.Scope {
	case core.TaskListScopeAssigned: // all "assigned" tasks are assigned to the user, filter by assigner
		assignerID = req.AssignerID
	case core.TaskListScopeRequested: // all "requested" tasks are assigned by the user, filter by assignee
		assigneeID = req.AssigneeID
	}

	tasks, err := h.core.ListTasks(
		ctx, uid, oid, req.Scope, assignerID, assigneeID, req.Since, req.IncludeComplete, req.PastDue, req.PriorityMin, limit,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to list tasks")
		http.Error(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, tasks, session.CtxGetParamPretty(ctx))
}

// RequestTaskCreate represents the payload for creating a new task
type RequestTaskCreate struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Priority    *core.TaskPriority `json:"priority,omitempty"`
	DueBy       *time.Time         `json:"due_by"`
	AssigneeID  *uuid.UUID         `json:"assignee_id"`
}

// PostApiTaskNew handles the creation of a new task via API
func (h *OnusHandler) PostApiTasksNew(w http.ResponseWriter, r *http.Request) {
	var req RequestTaskCreate
	if err := util.DecodeJSON(r, &req); err != nil {
		log.Error().Err(err).Msg("failed to decode request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// validate priority if provided
	if req.Priority != nil {
		if !req.Priority.IsValid() {
			http.Error(w, "Invalid priority value", http.StatusBadRequest)
			return
		}
	}

	// title is required
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// get user and organization IDs from context
	uid := session.CtxGetUID(r.Context())
	oid := session.CtxGetOID(r.Context())

	assigneeID := req.AssigneeID
	if assigneeID == nil {
		assigneeID = &uid
	}

	// create the task in the database
	task, err := h.core.CreateTask(r.Context(), oid, &core.TaskCreateParams{
		Title:          req.Title,
		Description:    req.Description,
		DueBy:          req.DueBy,
		Priority:       req.Priority,
		AssigneeUserID: assigneeID,
		AssignerUserID: uid,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create task")
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// successfully created the task
	util.WriteJSON(w, http.StatusCreated, task, session.CtxGetParamPretty(r.Context()))
}

// GetApiTaskByID handles fetching a specific task by its ID via API
func (h *OnusHandler) GetApiTaskByID(w http.ResponseWriter, r *http.Request) {
	taskID := session.CtxGetPathTaskID(r.Context())

	role := session.CtxGetRole(r.Context())
	uid := session.CtxGetOID(r.Context())
	oid := session.CtxGetOID(r.Context())

	task, err := h.core.GetTask(r.Context(), taskID, oid)
	if err != nil {
		if err == core.ErrNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("failed to get task")
		http.Error(w, "Failed to get the task", http.StatusInternalServerError)
		return
	}

	// check if the user has permission to view the task
	isAdming := role >= core.RoleAdmin
	isAssigner := task.Assigner == uid
	isAssignee := task.Assignee == uid

	if !isAdming && !isAssigner && !isAssignee {
		log.Error().Str("user_id", uid.String()).Msg("user is neither admin, assigner, nor assignee, deny access")
		http.Error(w, "Cannot view this task", http.StatusForbidden)
		return
	}

	util.WriteJSON(w, http.StatusOK, task, session.CtxGetParamPretty(r.Context()))
}

// RequestTaskUpdate represents the payload for updating a task
type RequestTaskUpdate struct {
	Title          *string            `json:"title,omitempty"`
	Description    *string            `json:"description,omitempty"`
	Notes          *string            `json:"notes,omitempty"`
	Progress       *int               `json:"progress,omitempty"`
	Priority       *core.TaskPriority `json:"priority,omitempty"`
	ClearDueBy     *bool              `json:"clear_due_by,omitempty"`
	DueBy          *time.Time         `json:"due_by,omitempty"`
	AssigneeUserID *uuid.UUID         `json:"assignee_user_id,omitempty"`
}

// PatchApiTaskByID handles updating a specific task by its ID via API
func (h *OnusHandler) PatchApiTaskByID(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "PatchApiTaskByID").Logger()

	taskID := session.CtxGetPathTaskID(r.Context())
	uid := session.CtxGetUID(r.Context())
	oid := session.CtxGetOID(r.Context())

	var req RequestTaskUpdate
	if err := util.DecodeJSON(r, &req); err != nil {
		log.Error().Err(err).Msg("failed to decode request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Priority != nil {
		if !req.Priority.IsValid() {
			http.Error(w, "Invalid priority value", http.StatusBadRequest)
			return
		}
	}

	if req.Progress != nil {
		if *req.Progress < 0 || *req.Progress > 100 {
			http.Error(w, "Progress must be between 0 and 100", http.StatusBadRequest)
			return
		}
	}

	task, err := h.core.GetTask(r.Context(), taskID, oid)
	if err != nil {
		if err == core.ErrNotFound {
			log.Error().Str("task_id", taskID.String()).Msg("task not found")
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("failed to get task")
		http.Error(w, "Failed to get task: ", http.StatusInternalServerError)
		return
	}

	isAdmin := session.CtxGetRole(r.Context()) >= core.RoleAdmin
	isAssigner := task.Assigner == uid
	isAssignee := task.Assignee == uid

	if isAdmin || isAssigner {
		// user is a manager of the task, allow full update
		task, err = h.core.UpdateTaskAsManager(r.Context(), taskID, oid, &core.TaskUpdateParamsManager{
			Title:          req.Title,
			Description:    req.Description,
			Notes:          req.Notes,
			ClearDueBy:     req.ClearDueBy,
			DueBy:          req.DueBy,
			Progress:       req.Progress,
			Priority:       req.Priority,
			AssigneeUserID: req.AssigneeUserID,
		})
	} else if isAssignee {
		// user is not admin or the assigner of the task, restrict fields they can update
		task, err = h.core.UpdateTaskAsAssignee(r.Context(), taskID, oid, &core.TaskUpdateParamsAssignee{
			Notes:    req.Notes,
			Progress: req.Progress,
		})
	} else {
		if !isAssignee {
			log.Error().Msg("user is neither admin, assigner, nor assignee, deny update")
			// user is neither admin, assigner, nor assignee, deny update
			http.Error(w, "Cannot update this task", http.StatusForbidden)
			return
		}
	}

	if err != nil {
		if err == core.ErrNotFound {
			log.Error().Str("task_id", taskID.String()).Msg("task not found during update")
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("failed to update task")
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, task, session.CtxGetParamPretty(r.Context()))
}

// DeleteApiTaskByID handles deleting a specific task by its ID via API
func (h *OnusHandler) DeleteApiTaskByID(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "DeleteApiTaskByID").Logger()

	ctx := r.Context()
	oid := session.CtxGetOID(ctx)
	uid := session.CtxGetUID(ctx)

	taskID := session.CtxGetPathTaskID(ctx)
	task, err := h.core.GetTask(ctx, taskID, oid)
	if err != nil {
		if err == core.ErrNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("failed to get task")
		http.Error(w, "Failed to get task: ", http.StatusInternalServerError)
		return
	}

	isAdmin := session.CtxGetRole(ctx) >= core.RoleAdmin
	isAssigner := task.Assigner == uid
	isAssignee := task.Assignee == uid

	if !isAdmin && !isAssigner {
		if !isAssignee {
			// user is neither admin, assigner, nor assignee, deny deletion
			log.Error().Str("user_id", uid.String()).Msg("user is neither admin, assigner, nor assignee, deny deletion")
			http.Error(w, "Insufficient permissions to delete this task", http.StatusForbidden)
			return
		}
		// user is not admin or the assigner of the task, deny deletion
		http.Error(w, "Insufficient permissions to delete this task", http.StatusForbidden)
		return
	}
	log.Debug().Str("task_id", taskID.String()).Msg("task deletion authorized")

	err = h.core.DeleteTaskByID(ctx, taskID, oid)
	if err != nil {
		if err == core.ErrNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("failed to delete task")
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}
	log.Debug().Str("task_id", taskID.String()).Msg("task deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}
