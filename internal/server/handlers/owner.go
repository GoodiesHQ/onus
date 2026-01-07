package handlers

import (
	"net/http"

	"github.com/goodieshq/onus/internal/server/session"
	"github.com/goodieshq/onus/internal/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type RequestTransferOwnership struct {
	NewOwnerUserID uuid.UUID `json:"new_owner_user_id"`
}

func (h *OnusHandler) PostApiOwnerTransfer(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "PostApiOwnerTransfer").Logger()

	var req RequestTransferOwnership
	if err := util.DecodeJSON(r, &req); err != nil {
		log.Error().Err(err).Msg("failed to decode request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	uid := session.CtxGetUID(ctx)
	err := h.core.TransferOwnership(ctx, uid, req.NewOwnerUserID, session.CtxGetOID(ctx))
	if err != nil {
		log.Error().Err(err).Msg("failed to transfer ownership")
		http.Error(w, "Failed to transfer ownership", http.StatusInternalServerError)
		return
	}

	// Invalidate the session's cached auth snapshot to ensure the new role is reflected
	session.InvalidateSnapshot(r.Context(), h.sm)

	w.WriteHeader(http.StatusNoContent)
}
