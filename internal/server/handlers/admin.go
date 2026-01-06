package handlers

import (
	"net/http"

	"github.com/goodieshq/onus/internal/server/core"
	"github.com/goodieshq/onus/internal/server/session"
	"github.com/goodieshq/onus/internal/util"
	"github.com/rs/zerolog/log"
)

type RequestUpdateOrganization struct {
	Name string `json:"name"`
}

// PatchApiAdminOrganization handles updating the organization's details
func (h *OnusHandler) PatchApiAdminOrganization(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "PatchApiAdminOrganization").Logger()

	// Parse the rename request
	var req RequestUpdateSelf
	if err := util.DecodeJSON(r, &req); err != nil {
		log.Error().Err(err).Msg("failed to decode request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get organization ID from session
	oid := session.CtxGetOID(r.Context())

	// Update the organization name
	updatedOrg, err := h.core.UpdateOrganization(r.Context(), oid, util.Ptr(req.Name))
	if err != nil {
		log.Error().Err(err).Msg("failed to update organization")
		http.Error(w, "Failed to update organization", http.StatusInternalServerError)
		return
	}

	// Invalidate the session's cached auth snapshot to ensure the new org name is reflected
	session.InvalidateSnapshot(r.Context(), h.sm)

	util.WriteJSON(w, http.StatusOK, updatedOrg, session.CtxGetParamPretty(r.Context()))
}

type RequestDisableUser struct {
	Reason string `json:"reason"`
}

// PostApiAdminUserDisable handles disabling a user account
func (h *OnusHandler) PostApiAdminUserDisable(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "PostApiAdminUserDisable").Logger()

	// Parse the disable user request
	var req RequestDisableUser
	if err := util.DecodeJSON(r, &req); err != nil {
		log.Error().Err(err).Msg("failed to decode request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	uid := session.CtxGetUID(ctx)
	oid := session.CtxGetOID(ctx)
	role := session.CtxGetRole(ctx)
	targetUserID := session.CtxGetPathUserID(ctx)

	// Prevent admins from disabling their own account
	if uid == targetUserID {
		http.Error(w, "Cannot disable your own user account", http.StatusBadRequest)
		return
	}

	// Get the target user's role and disabled status
	targetUserRole, targetUserDisabled, err := h.core.GetAssignment(ctx, targetUserID, oid)
	if err != nil {
		if err == core.ErrNotFound {
			http.Error(w, "Target user not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("failed to get target user assignment")
		http.Error(w, "Failed to get target user assignment", http.StatusInternalServerError)
		return
	}

	// Check if the target user is already disabled
	if targetUserDisabled {
		http.Error(w, "User is already disabled", http.StatusBadRequest)
		return
	}

	// Ensure the admin has sufficient permissions to disable the target user
	if role < targetUserRole {
		http.Error(w, "Insufficient permissions to disable target user", http.StatusForbidden)
		return
	}

	// Disable the user
	assignment, err := h.core.DisableUser(ctx, targetUserID, oid, req.Reason)
	if err != nil {
		log.Error().Err(err).Msg("failed to disable user")
		http.Error(w, "Failed to disable user", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, assignment, session.CtxGetParamPretty(ctx))
}

func (h *OnusHandler) PostApiAdminUserEnable(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "PostApiAdminUserEnable").Logger()
	ctx := r.Context()

	uid := session.CtxGetUID(ctx)
	oid := session.CtxGetOID(ctx)
	role := session.CtxGetRole(ctx)
	targetUserID := session.CtxGetPathUserID(ctx)

	// Prevent admins from enabling their own account
	if uid == targetUserID {
		http.Error(w, "Cannot enable your own user account", http.StatusBadRequest)
		return
	}

	// Get the target user's role and disabled status
	targetUserRole, targetUserDisabled, err := h.core.GetAssignment(ctx, targetUserID, oid)
	if err != nil {
		if err == core.ErrNotFound {
			http.Error(w, "Target user not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("failed to get target user assignment")
		http.Error(w, "Failed to get target user assignment", http.StatusInternalServerError)
		return
	}

	// Check if the target user is already enabled
	if !targetUserDisabled {
		http.Error(w, "User is already enabled", http.StatusBadRequest)
		return
	}

	// Ensure the admin has sufficient permissions to enable the target user
	if role < targetUserRole {
		http.Error(w, "Insufficient permissions to enable target user", http.StatusForbidden)
		return
	}

	// Enable the user
	assignment, err := h.core.EnableUser(ctx, targetUserID, oid)
	if err != nil {
		log.Error().Err(err).Msg("failed to enable user")
		http.Error(w, "Failed to enable user", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, assignment, session.CtxGetParamPretty(ctx))
}

type RequestUpdateUserRole struct {
	Role core.UserOrgRole `json:"role"`
}

func (h *OnusHandler) PatchApiAdminUserRole(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "PatchApiAdminUserRole").Logger()
	ctx := r.Context()

	var req RequestUpdateUserRole
	if err := util.DecodeJSON(r, &req); err != nil {
		log.Error().Err(err).Msg("failed to decode request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate the requested role
	if !req.Role.IsValid() {
		http.Error(w, "Invalid role specified", http.StatusUnprocessableEntity)
		return
	}

	uid := session.CtxGetUID(ctx)
	oid := session.CtxGetOID(ctx)
	role := session.CtxGetRole(ctx)
	targetUserID := session.CtxGetPathUserID(ctx)

	// Prevent admins from changing their own role
	if uid == targetUserID {
		http.Error(w, "Cannot modif/y your own user role", http.StatusBadRequest)
		return
	}

	if req.Role > role {
		http.Error(w, "Cannot assign a role higher than your own", http.StatusForbidden)
		return
	}

	if req.Role == core.RoleOwner {
		http.Error(w, "Cannot assign owner role, must transfer ownership", http.StatusForbidden)
		return
	}

	// Update the user's role
	assignment, err := h.core.UpdateUserRole(ctx, targetUserID, oid, req.Role)
	if err != nil {
		log.Error().Err(err).Msg("failed to update user role")
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, assignment, session.CtxGetParamPretty(ctx))
}
