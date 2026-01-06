package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/goodieshq/onus/internal/util"
	"gopkg.in/yaml.v3"
)

const (
	defaultServerPort = uint16(8080)
)

type OnusMode int

const (
	ModeServe OnusMode = iota
	ModeMigrate
)

type ServerConfig struct {
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
	URL  string `yaml:"url"`
}

type DatabaseConfig struct {
	Type string `yaml:"type"`
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
	Name string `yaml:"name"`
	SSL  bool   `yaml:"ssl"`
}

func (dbc *DatabaseConfig) DSN() string {
	switch dbc.Type {
	case "postgres":
		if dbc.Port == 0 {
			dbc.Port = 5432
		}
		ssl := "disable"
		if dbc.SSL {
			ssl = "enable"
		}
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dbc.Host, dbc.Port, dbc.User, dbc.Pass, dbc.Name, ssl,
		)
	}

	return "<unsupported>"
}

type AuthOIDProviderConfig struct {
	IssuerURL    string   `yaml:"issuer_url"`
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	Scopes       []string `yaml:"scopes"` // Optional, defaults to [openid, profile, email]
}

type AuthOIDCConfig struct {
	// How do I allow custom-named providers here?
	Enabled   bool                             `yaml:"enabled"`
	Providers map[string]AuthOIDProviderConfig `yaml:"providers"`
}

type AuthConfig struct {
	OIDC AuthOIDCConfig `yaml:"oidc"`
}

type OnusConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
}

var supportedDBTypes = map[string]bool{
	"postgres": true,
}

var supportedDBTypesList []string

func init() {
	supportedDBTypesList = make([]string, 0, len(supportedDBTypes))
	for dbType := range supportedDBTypes {
		supportedDBTypesList = append(supportedDBTypesList, dbType)
	}
}

func (oc *OnusConfig) Validate(mode OnusMode) error {
	if oc.Database.Type == "" {
		return fmt.Errorf("database type must be set")
	}

	if _, ok := supportedDBTypes[oc.Database.Type]; !ok {
		return fmt.Errorf("unsupported database type: %s. must be one of %v", oc.Database.Type, supportedDBTypesList)
	}

	/*
		if oc.Database.Port == 0 {
			return fmt.Errorf("database port must be set and non-zero")
		}
	*/

	if oc.Database.User == "" {
		return fmt.Errorf("database user must be set")
	}

	if oc.Database.Pass == "" {
		return fmt.Errorf("database password must be set")
	}

	if oc.Database.Name == "" {
		return fmt.Errorf("database name must be set")
	}

	if mode == ModeServe {
		if oc.Server.Port == 0 {
			return fmt.Errorf("server port must be set and non-zero")
		}

		if oc.Server.URL == "" {
			return fmt.Errorf("server url must be set")
		}

		if oc.Auth.OIDC.Enabled {
			if len(oc.Auth.OIDC.Providers) == 0 {
				return fmt.Errorf("at least one OIDC provider must be configured")
			}
			for name, provider := range oc.Auth.OIDC.Providers {
				if name != strings.ToLower(name) {
					return fmt.Errorf("oidc provider name %q must be lowercase", name)
				}

				if provider.IssuerURL == "" {
					return fmt.Errorf("oidc provider %s: issuer_url must be set", name)
				}
				if provider.ClientID == "" {
					return fmt.Errorf("oidc provider %s: client_id must be set", name)
				}
				if provider.ClientSecret == "" {
					return fmt.Errorf("oidc provider %s: client_secret must be set", name)
				}
			}
		}

		// TODO: add more auth types here once implemented, right now only OIDC exists
		if !oc.Auth.OIDC.Enabled {
			return fmt.Errorf("at least one authentication provider must be enabled")
		}
	}
	return nil
}

// envOverrides overrides config values with environment variables if they are set
func envOverrides(cfg *OnusConfig) error {
	// Server overrides
	if host := os.Getenv("ONUS_SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}

	if port := os.Getenv("ONUS_SERVER_PORT"); port != "" {
		p, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return fmt.Errorf("invalid ONUS_DB_PORT: %w", err)
		}
		cfg.Server.Port = uint16(p)
	}

	if url := os.Getenv("ONUS_SERVER_URL"); url != "" {
		cfg.Server.URL = url
	}

	// Database overrides
	if dbType := os.Getenv("ONUS_DB_TYPE"); dbType != "" {
		cfg.Database.Type = dbType
	}

	if dbHost := os.Getenv("ONUS_DB_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}

	if dbPort := os.Getenv("ONUS_DB_PORT"); dbPort != "" {
		p, err := strconv.ParseUint(dbPort, 10, 16)
		if err != nil {
			return fmt.Errorf("invalid ONUS_DB_PORT: %w", err)
		}
		cfg.Database.Port = uint16(p)
	}

	if dbUser := os.Getenv("ONUS_DB_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}

	if dbPass := os.Getenv("ONUS_DB_PASS"); dbPass != "" {
		cfg.Database.Pass = dbPass
	}

	if dbName := os.Getenv("ONUS_DB_NAME"); dbName != "" {
		cfg.Database.Name = dbName
	}

	if dbSSL := os.Getenv("ONUS_DB_SSL"); dbSSL != "" {
		b, err := util.ParseBool(dbSSL)
		if err != nil {
			return fmt.Errorf("invalid ONUS_DB_SSL: %w", err)
		}
		cfg.Database.SSL = b
	}

	if oidcEnabled := os.Getenv("ONUS_AUTH_OIDC_ENABLED"); oidcEnabled != "" {
		b, err := util.ParseBool(oidcEnabled)
		if err != nil {
			return fmt.Errorf("invalid ONUS_AUTH_OIDC_ENABLED: %w", err)
		}
		cfg.Auth.OIDC.Enabled = b
	}

	if cfg.Auth.OIDC.Providers == nil {
		cfg.Auth.OIDC.Providers = make(map[string]AuthOIDProviderConfig)
	}

	providers := envProvidersOIDC()
	for name, provider := range providers {
		cfg.Auth.OIDC.Providers[name] = provider
	}

	return nil
}

func envProvidersOIDC() map[string]AuthOIDProviderConfig {
	providers := make(map[string]AuthOIDProviderConfig)
	const prefix = "ONUS_AUTH_OIDC_PROVIDER_"

	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		trimmed := strings.TrimPrefix(k, prefix)
		providerKey, field, ok := splitProviderFieldOIDC(trimmed)
		if !ok {
			continue
		}

		provider := providers[providerKey]
		switch field {
		case "ISSUER_URL":
			provider.IssuerURL = v
		case "CLIENT_ID":
			provider.ClientID = v
		case "CLIENT_SECRET":
			provider.ClientSecret = v
		case "SCOPES":
			scopes := strings.Split(v, ",")
			for i := range scopes {
				scopes[i] = strings.TrimSpace(scopes[i])
			}
			provider.Scopes = scopes
		}
		providers[providerKey] = provider
	}

	return providers
}

// / splitProviderFieldOIDC splits a string of the form "PROVIDERNAME_FIELD" into "PROVIDERNAME" and "FIELD"
func splitProviderFieldOIDC(s string) (string, string, bool) {
	s = strings.ToUpper(strings.TrimSpace(s))

	// oidc provider fields
	fields := []string{
		"ISSUER_URL", "CLIENT_ID", "CLIENT_SECRET", "SCOPES",
	}

	// look for known fields as suffixes
	for _, field := range fields {
		suffix := "_" + field
		if strings.HasSuffix(s, suffix) {
			providerName := strings.TrimSuffix(s, suffix)
			if providerName == "" {
				return "", "", false
			}
			return strings.ToLower(providerName), field, true
		}
	}
	return "", "", false
}

func Load(path string, mode OnusMode) (*OnusConfig, error) {
	var config OnusConfig

	cfg, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		// file does not exist, try to load from env only
	} else {
		err = yaml.Unmarshal(cfg, &config)
		if err != nil {
			return nil, err
		}
	}

	if err := envOverrides(&config); err != nil {
		return nil, err
	}

	if config.Server.Port == 0 {
		config.Server.Port = defaultServerPort
	}

	config.Server.URL = strings.TrimRight(config.Server.URL, "/")

	if err := config.Validate(mode); err != nil {
		return nil, err
	}

	return &config, nil
}
