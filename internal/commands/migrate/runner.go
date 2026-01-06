package migrate

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/goodieshq/onus/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

type OnusMigrateDirection int

const (
	Up OnusMigrateDirection = iota
	Down
)

func currentVersion(m *migrate.Migrate) (uint, bool, error) {
	v, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return v, dirty, nil
}

func Status(ctx context.Context, c *cli.Command) error {
	cfg, err := config.Load(c.String("config"), config.ModeMigrate)
	if err != nil {
		return fmt.Errorf("failed to load the configuration: %w", err)
	}

	switch cfg.Database.Type {
	case "postgres":
		pool, close, err := newPostgresPool(ctx, cfg.Database.DSN())
		if err != nil {
			return fmt.Errorf("failed to connect to the database: %w", err)
		}
		defer close()
		m, err := newMigratorPostgres(ctx, pool)
		if err != nil {
			return fmt.Errorf("failed to create migrator: %w", err)
		}
		defer migrateCloser(m)()

		if v, dirty, verr := currentVersion(m); verr == nil {
			output := fmt.Sprintf("Version %d", v)
			if dirty {
				output += " (dirty)"
			} else {
				output += " (clean)"
			}
			fmt.Println(output)
		} else {
			return fmt.Errorf("failed to get migration status: %w", verr)
		}
	}
	return nil
}

func Force(ctx context.Context, c *cli.Command, version uint) error {
	cfg, err := config.Load(c.String("config"), config.ModeMigrate)
	if err != nil {
		return fmt.Errorf("failed to load the configuration: %w", err)
	}

	switch cfg.Database.Type {
	case "postgres":
		return runMigrateForcePostgres(ctx, cfg.Database.DSN(), version)
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}
}

func To(ctx context.Context, c *cli.Command, to uint) error {
	cfg, err := config.Load(c.String("config"), config.ModeMigrate)
	if err != nil {
		return fmt.Errorf("failed to load the configuration: %w", err)
	}

	switch cfg.Database.Type {
	case "postgres":
		return runMigrateToPostgres(ctx, cfg.Database.DSN(), to)
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}
}

func migrateCloser(m *migrate.Migrate) func() {
	return func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Error().Err(srcErr).Msg("failed to close migration source cleanly")
		}
		if dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to close database connection cleanly")
		}
	}
}

func Run(ctx context.Context, c *cli.Command, dir OnusMigrateDirection, steps uint) error {
	cfg, err := config.Load(c.String("config"), config.ModeMigrate)
	if err != nil {
		return fmt.Errorf("failed to load the configuration: %w", err)
	}

	switch cfg.Database.Type {
	case "postgres":
		return runMigrationsPostgres(ctx, cfg.Database.DSN(), dir, steps)
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}
}
