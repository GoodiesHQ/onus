package database

import (
	"context"
	"fmt"

	"github.com/goodieshq/onus/internal/config"
	"github.com/goodieshq/onus/internal/server/core"
	"github.com/goodieshq/onus/internal/server/core_pgx"
)

func NewDatabaseCore(ctx context.Context, cfg *config.DatabaseConfig) (core.Core, error) {
	switch cfg.Type {
	case "postgres":
		return core_pgx.NewCorePGX(ctx, cfg.DSN())
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}
