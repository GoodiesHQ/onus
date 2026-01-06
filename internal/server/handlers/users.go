package handlers

import (
	"net/http"

	"github.com/goodieshq/onus/internal/server/session"
	"github.com/goodieshq/onus/internal/util"
	"github.com/rs/zerolog/log"
)

// GetApiUsers handles requests to list enabled users, optionally filtered by a search query.
func (h *OnusHandler) GetApiUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	ctx := r.Context()

	// List all enabled users
	users, err := h.core.ListEnabledUsers(
		ctx, session.CtxGetOID(ctx), search,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to list users")
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, users, session.CtxGetParamPretty(ctx))
}

// GetApiAdminUsers handles requests to list all users with their assignments for admin purposes.
func (h *OnusHandler) GetApiAdminUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	oid := session.CtxGetOID(ctx)

	// List all users with their assignments.
	users, err := h.core.ListAllUsersWithAssignments(ctx, oid)
	if err != nil {
		log.Error().Err(err).Msg("failed to list all users with assignments")
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, users, session.CtxGetParamPretty(ctx))
}
