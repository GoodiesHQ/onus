package server

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/goodieshq/onus/internal/server/core"
	"github.com/goodieshq/onus/internal/server/session"
	"github.com/goodieshq/onus/internal/util"
	"github.com/google/uuid"
	"github.com/oklog/ulid"
	"github.com/rs/zerolog/log"
)

// MiddlewareMinRole checks if the user has at least the minimum role required
func MiddlewareMinRole(sm *scs.SessionManager, c core.Core, roleMin core.UserOrgRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := session.CtxGetRole(r.Context())
			if role < roleMin {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// MiddlewareCheckSession checks if a valid authenticated session exists
func MiddlewareCheckSession(c core.Core, sm *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// check if the session exists
			sub := sm.GetString(ctx, "principal_subject")
			uidStr := sm.GetString(ctx, "user_id")
			oidStr := sm.GetString(ctx, "organization_id")

			// if any of the required session values are missing, return unauthorized
			if uidStr == "" || oidStr == "" || sub == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Ensure the user_id is a valid UUID
			uid, err := uuid.Parse(uidStr)
			if err != nil {
				log.Error().Err(err).Msg("invalid user_id in session")
				http.Error(w, "Invalid session data", http.StatusUnauthorized)
				return
			}

			// Ensure the organization_id is a valid UUID
			oid, err := uuid.Parse(oidStr)
			if err != nil {
				log.Error().Err(err).Msg("invalid organization_id in session")
				http.Error(w, "Invalid session data", http.StatusUnauthorized)
				return
			}

			// Get the user's role and status within the organization from the session
			role, disabled, err := c.GetAssignment(ctx, uid, oid)
			if err != nil {
				if err == core.ErrNotFound {
					log.Error().Str("user_id", uidStr).Str("organization_id", oidStr).Msg("role assignment not found")
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				log.Error().Err(err).Msg("failed to get role assignment")
				http.Error(w, "Failed to verify permissions", http.StatusInternalServerError)
				return
			}

			// check if the user is disabled, deauthenticate if so
			if disabled {
				if err := session.ClearSession(r.Context(), sm); err != nil {
					log.Error().Err(err).Msg("failed to clear session for disabled user")
				}
				http.Error(w, "User account is disabled", http.StatusUnauthorized)
				return
			}

			if !role.IsValid() {
				log.Error().Str("role", role.String()).Msg("invalid role in session")
				http.Error(w, "Invalid session data", http.StatusUnauthorized)
				return
			}

			// save the uid in the request context
			ctx = session.CtxSetUID(ctx, uid)
			ctx = session.CtxSetOID(ctx, oid)
			ctx = session.CtxSetRole(ctx, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MiddlewarePathTaskID extracts the task_id from the URL path and adds it to the request context
func MiddlewarePathTaskID(c core.Core, sm *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ensure task ID is valid
			taskIDstring := chi.URLParam(r, "task_id")
			taskID, err := ulid.ParseStrict(taskIDstring)
			if err != nil {
				http.Error(w, "Invalid task ID", http.StatusBadRequest)
				return
			}
			ctx := session.CtxSetPathTaskID(r.Context(), taskID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MiddlewarePathUserID extracts the user_id from the URL path and adds it to the request context
func MiddlewarePathUserID(c core.Core, sm *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ensure user ID is valid
			userIDstring := chi.URLParam(r, "user_id")
			userID, err := uuid.Parse(userIDstring)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}

			ctx := session.CtxSetPathUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Middleware to see if pretty print is requested via query param
func MiddlewarePrettyPrint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pretty, _ := util.ParseBool(r.URL.Query().Get("pretty"))
		ctx := session.CtxSetParamPretty(r.Context(), pretty)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
