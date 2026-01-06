package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/goodieshq/onus/internal/server/auth"
	"github.com/goodieshq/onus/internal/server/session"
	"github.com/goodieshq/onus/internal/util"
	"github.com/rs/zerolog/log"
)

// GetAuthList handles requests to list available authentication providers
func (h *OnusHandler) GetAuthList(w http.ResponseWriter, r *http.Request) {
	providers := auth.ListProviders()
	ctx := r.Context()

	util.WriteJSON(w, http.StatusOK, providers, session.CtxGetParamPretty(ctx))
}

// GetAuthProviderLogin initiates the login process for a specific authentication provider
func (h *OnusHandler) GetAuthProviderLogin(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "GetAuthProviderLogin").Logger()

	providerName := chi.URLParam(r, "provider")
	provider, exists := auth.GetProvider(providerName)
	if !exists {
		log.Error().Str("provider", providerName).Msg("unknown auth provider")
		http.Error(w, "Unknown auth provider", http.StatusBadRequest)
		return
	}
	provider.StartAuth(w, r)
}

// GetAuthLogout handles user logout by clearing the session
func (h *OnusHandler) GetAuthLogout(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "GetAuthLogout").Logger()

	err := session.ClearSession(r.Context(), h.sm)
	if err != nil {
		log.Error().Err(err).Msg("failed to clear session on logout")
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// GetAuthProviderCallback handles the callback from the authentication provider
func (h *OnusHandler) GetAuthProviderCallback(w http.ResponseWriter, r *http.Request) {
	log := log.With().Str("handler", "GetAuthProviderCallback").Logger()

	providerName := chi.URLParam(r, "provider")
	provider, exists := auth.GetProvider(providerName)
	if !exists {
		log.Error().Str("provider", providerName).Msg("unknown auth provider")
		http.Error(w, "Unknown auth provider", http.StatusBadRequest)
		return
	}

	principal, err := provider.HandleCallback(w, r)
	if err != nil {
		http.Error(w, "Failed to handle auth callback: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debug().Any("principal", principal).Msg("user authenticated via sso")

	// Register or login the user
	user, err := h.core.SignupOrLogin(r.Context(), principal.Email, principal.Name)
	if err != nil {
		log.Error().Err(err).Msg("failed to signup or login user")
		http.Error(w, "Failed to authenticate user.", http.StatusInternalServerError)
		return
	}
	log.Debug().Msg("user logged in: " + user.Email)

	// Get the associated organization for the user
	org, err := h.core.GetOrganizationByUserID(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get organization for user")
		http.Error(w, "Failed to authenticate user.", http.StatusInternalServerError)
		return
	}

	// Get a snapshot for the user in the organization
	snapshot, err := h.core.GetAuthSnapshot(r.Context(), user.ID, org.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get auth snapshot for user")
		http.Error(w, "Failed to authenticate user.", http.StatusInternalServerError)
		return
	}

	// Check if the user is disabled
	if snapshot.Disabled {
		log.Debug().Str("user", user.Email).Msg("disabled user attempted to login")
		http.Error(w, "User account is disabled", http.StatusUnauthorized)
		return
	}

	// We have successfully authenticated the user
	h.sm.Put(r.Context(), "principal_subject", principal.Subject)
	h.sm.Put(r.Context(), "principal_issuer", principal.Issuer)

	// Store auth snapshot in session
	session.SetAuthSnapshot(r.Context(), h.sm, snapshot)

	// Renew the session
	if err := h.sm.RenewToken(r.Context()); err != nil {
		log.Error().Err(err).Msg("failed to renew session after login")
		http.Error(w, "Failed to renew session", http.StatusInternalServerError)
		return
	}

	// Redirect to the app home page
	http.Redirect(w, r, "/app", http.StatusFound)
}
