package auth

import (
	"net/http"
	"strings"
)

// AuthProvider defines the interface that OAuth providers must implement
type AuthProvider interface {
	StartAuth(w http.ResponseWriter, r *http.Request)
	HandleCallback(w http.ResponseWriter, r *http.Request) (*Principal, error)
}

// list of all registered providers
var providers = make(map[string]AuthProvider)

// ListProviders returns the names of all registered OAuth providers
func ListProviders() []string {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	return names
}

// RegisterProvider registers a new OAuth provider with the given name
func RegisterProvider(name string, provider AuthProvider) {
	name = strings.ToLower(name)
	providers[name] = provider
}

// GetProvider retrieves a registered OAuth provider by name
func GetProvider(name string) (AuthProvider, bool) {
	provider, exists := providers[strings.ToLower(name)]
	return provider, exists
}
