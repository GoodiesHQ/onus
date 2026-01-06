package server

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/goodieshq/onus/internal/config"
	"github.com/goodieshq/onus/internal/server/auth"
	"github.com/goodieshq/onus/internal/server/core"
	"github.com/goodieshq/onus/internal/server/handlers"
	"github.com/rs/zerolog/log"
)

type OnusServer struct {
	host      string
	port      uint16
	url       string
	core      core.Core
	sm        *scs.SessionManager
	staticDir *string
}

// Url constructs a full URL by appending the given parts to the base URL
func (s *OnusServer) Url(parts ...string) string {
	fullUrl := s.url
	for _, p := range parts {
		fullUrl += "/" + p
	}
	return fullUrl
}

func NewOnusServer(cfg *config.ServerConfig, core core.Core, store scs.Store) *OnusServer {
	// Initialize the session manager
	sm := scs.New()
	if store != nil {
		sm.Store = store
	}
	sm.Lifetime = 7 * 24 * time.Hour
	sm.Cookie.SameSite = http.SameSiteLaxMode
	sm.Cookie.HttpOnly = true
	if strings.HasPrefix(cfg.URL, "https://") {
		sm.Cookie.Secure = true
	}

	var staticDir *string
	info, err := os.Stat(cfg.StaticDir)
	if err != nil || !info.IsDir() {
		log.Warn().Msgf("Static directory %q is invalid. Serving API only.", cfg.StaticDir)
	} else {
		staticDir = &cfg.StaticDir
	}

	// Create and return the OnusServer instance
	return &OnusServer{
		host:      cfg.Host,
		port:      cfg.Port,
		url:       cfg.URL,
		core:      core,
		sm:        sm,
		staticDir: staticDir,
	}
}

// RegisterProviderOIDC registers a new OIDC authentication provider with the server
func (s *OnusServer) RegisterProviderOIDC(ctx context.Context, name string, cfg *auth.ConfigOIDC) error {
	p, err := auth.NewProviderOIDC(ctx, cfg, s.sm)
	if err != nil {
		return err
	}

	auth.RegisterProvider(name, p)
	return nil
}

// Run starts the Onus server and listens for incoming HTTP requests
func (s *OnusServer) Run(ctx context.Context, cfg *config.OnusConfig) error {
	for name, providerCfg := range cfg.Auth.OIDC.Providers {
		provider, err := auth.NewProviderOIDC(ctx, &auth.ConfigOIDC{
			IssuerURL:    providerCfg.IssuerURL,
			ClientID:     providerCfg.ClientID,
			ClientSecret: providerCfg.ClientSecret,
			RedirectURL:  cfg.Server.URL + "/auth/" + name + "/callback",
			Scopes:       providerCfg.Scopes,
		}, s.sm)
		if err != nil {
			return fmt.Errorf("failed to create OIDC provider %s: %w", name, err)
		}
		auth.RegisterProvider(name, provider)
	}

	r := s.routes(ctx)

	if s.staticDir != nil {
		r.NotFound(s.static(*s.staticDir))
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		httpServer.Shutdown(ctx)
	}()

	log.Info().Msgf("Starting Onus server on %s:%d", s.host, s.port)
	if err := httpServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}
		log.Info().Msg("Onus server has stopped")
	}
	return nil
}

func init() {
	// Register types for session storage
	gob.Register(time.Time{})
}

func (s *OnusServer) static(staticDir string) http.HandlerFunc {
	fs := http.Dir(staticDir)
	fileServer := http.FileServer(fs)
	indexPath := filepath.Join(staticDir, "index.html")

	if _, err := os.Stat(indexPath); err != nil {
		log.Error().Err(err).Msgf("index.html not found at %q", indexPath)
	} else {
		log.Info().Msgf("Serving SPA index at %q", indexPath)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// serve the static folder here...
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/auth") {
			http.NotFound(w, r)
			return
		}

		// Normalize path
		clean := path.Clean(r.URL.Path)
		if clean == "." {
			clean = "/"
		}

		if clean == "/" {
			w.Header().Set("Cache-Control", "no-store")
			http.ServeFile(w, r, indexPath)
			return
		}

		// Check if the file has an extension, assume it's a static asset
		if ext := path.Ext(clean); ext != "" {
			if strings.HasPrefix(clean, "/_app/") {
				// TODO: adjust caching as needed
				// w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, indexPath)
	})
}

func (s *OnusServer) routes(ctx context.Context) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(s.sm.LoadAndSave)

	// Health check endpoint, including database connectivity check, useful for health probes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		db, close, err := s.core.GetDB()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer close()
		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
		}
		w.Write([]byte("OK"))
	})

	handler := handlers.NewOnusHandler(s.core, s.sm)

	// Authentication related routes
	r.Route("/auth", func(r chi.Router) {
		// List available authentication providers
		r.Get("/", handler.GetAuthList)

		// Routes for each authentication provider (login and callback)
		r.Get("/{provider}/login", handler.GetAuthProviderLogin)
		r.Get("/{provider}/callback", handler.GetAuthProviderCallback)

		// Logout route
		r.Route("/logout", func(r chi.Router) {
			r.Use(MiddlewareCheckSession(s.core, s.sm))
			r.Get("/", handler.GetAuthLogout)
		})
	})

	// Version endpoint
	r.Route("/api/version", func(r chi.Router) {
		r.Use(MiddlewarePrettyPrint)
		r.Get("/", handler.GetApiVersion)
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// All API routes require a valid session
		r.Use(MiddlewareCheckSession(s.core, s.sm))
		r.Use(MiddlewarePrettyPrint)

		// Get and set current user info
		r.Get("/me", handler.GetApiMe)
		r.Patch("/me", handler.PatchApiMe)

		// task-related routes
		r.Route("/tasks", func(r chi.Router) {
			// List tasks
			r.Get("/", handler.GetApiTasks)

			// Create new task
			r.Post("/new", handler.PostApiTasksNew)

			// Routes for specific task identified by task_id
			r.Route("/{task_id}", func(r chi.Router) {
				r.Use(MiddlewarePathTaskID(s.core, s.sm))

				// Get, update, and delete specific task
				r.Get("/", handler.GetApiTaskByID)
				r.Patch("/", handler.PatchApiTaskByID)
				r.Delete("/", handler.DeleteApiTaskByID)
			})
		})

		// user-related routes
		r.Route("/users", func(r chi.Router) {
			// List enabled users
			r.Get("/", handler.GetApiUsers)
		})

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			// All admin routes require at least an admin role
			r.Use(MiddlewareMinRole(s.sm, s.core, core.RoleAdmin))

			// Organization management
			r.Patch("/organization", handler.PatchApiAdminOrganization)

			// User management
			r.Route("/users", func(r chi.Router) {
				r.Get("/", handler.GetApiAdminUsers)

				// Routes for specific user identified by user_id
				r.Route("/{user_id}", func(r chi.Router) {
					r.Use(MiddlewarePathUserID(s.core, s.sm))

					// Enable or disable specific user
					r.Post("/disable", handler.PostApiAdminUserDisable)
					r.Post("/enable", handler.PostApiAdminUserEnable)

					// Update user role
					r.Patch("/role", handler.PatchApiAdminUserRole)
				})
			})
		})
	})

	return r
}
