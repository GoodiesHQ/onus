package handlers

import (
	"net/http"
	"strings"

	"github.com/goodieshq/onus/internal/server/session"
	"github.com/goodieshq/onus/internal/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (h *OnusHandler) GetApiMe(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "GetApiMe").Logger()

	uidStr := h.sm.GetString(r.Context(), "user_id")
	oidStr := h.sm.GetString(r.Context(), "organization_id")

	if uidStr == "" || oidStr == "" {
		_ = session.ClearSession(r.Context(), h.sm)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if session.Outdated(r.Context(), h.sm) {
		log.Info().Msg("auth snapshot outdated, refreshing from database")
		userID, err := uuid.Parse(uidStr)
		if err != nil {
			log.Error().Err(err).Msg("invalid user_id in session")
			http.Error(w, "Invalid session data", http.StatusInternalServerError)
			return
		}

		orgID, err := uuid.Parse(oidStr)
		if err != nil {
			log.Error().Err(err).Msg("invalid organization_id in session")
			http.Error(w, "Invalid session data", http.StatusInternalServerError)
			return
		}

		snapshot, err := h.core.GetAuthSnapshot(r.Context(), userID, orgID)
		if err != nil {
			log.Error().Err(err).Msg("failed to get auth snapshot")
			http.Error(w, "Failed to get user data", http.StatusInternalServerError)
			return
		}

		session.SetAuthSnapshot(r.Context(), h.sm, snapshot)
	} else {
		log.Info().Msg("using cached auth snapshot from session")
	}

	snapshot, err := session.GetAuthSnapshot(r.Context(), h.sm)
	if err != nil {
		log.Error().Err(err).Msg("failed to get auth snapshot from session")
		http.Error(w, "Failed to get user data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	util.WriteJSON(w, http.StatusOK, snapshot, session.CtxGetParamPretty(r.Context()))
}

type RequestUpdateSelf struct {
	Name string `json:"name"`
}

// PatchApiMe handles updating the authenticated user's own profile.
func (h *OnusHandler) PatchApiMe(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "PatchApiMe").Logger()

	var req RequestUpdateSelf
	if err := util.DecodeJSON(r, &req); err != nil {
		log.Error().Err(err).Msg("failed to decode request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}

	// Get the user ID from the session context
	uid := session.CtxGetUID(r.Context())

	// Update the user's name
	updatedUser, err := h.core.UpdateUserName(r.Context(), uid, name)
	if err != nil {
		log.Error().Err(err).Msg("failed to update user")
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	// Invalidate the auth snapshot in the session to force a refresh
	session.InvalidateSnapshot(r.Context(), h.sm)

	util.WriteJSON(w, http.StatusOK, updatedUser, session.CtxGetParamPretty(r.Context()))
}
