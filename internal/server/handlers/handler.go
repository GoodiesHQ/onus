package handlers

import (
	"github.com/alexedwards/scs/v2"
	"github.com/goodieshq/onus/internal/server/core"
)

// OnusHandler is the main handler struct for Onus server, almost
// everything will use the core and session manager
type OnusHandler struct {
	core core.Core
	sm   *scs.SessionManager
}

func NewOnusHandler(core core.Core, sm *scs.SessionManager) *OnusHandler {
	return &OnusHandler{
		core: core,
		sm:   sm,
	}
}
