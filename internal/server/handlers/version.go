package handlers

import (
	"net/http"

	"github.com/goodieshq/onus/internal/buildinfo"
	"github.com/goodieshq/onus/internal/server/session"
	"github.com/goodieshq/onus/internal/util"
)

func (h *OnusHandler) GetApiVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	version := buildinfo.GetBuildInfo()
	util.WriteJSON(w, http.StatusOK, version, session.CtxGetParamPretty(ctx))
}
