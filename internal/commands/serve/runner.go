package serve

import (
	"context"
	"fmt"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/goodieshq/onus/internal/config"
	"github.com/goodieshq/onus/internal/database"
	"github.com/goodieshq/onus/internal/server"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, c *cli.Command) error {
	cfg, err := config.Load(c.String("config"), config.ModeServe)
	if err != nil {
		return fmt.Errorf("failed to load the configuration: %w", err)
	}

	core, err := database.NewDatabaseCore(ctx, &cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize the database: %w", err)
	}

	db, cleanup, err := core.GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database handle: %w", err)
	}
	defer cleanup()

	pgstore := postgresstore.New(db)

	srv := server.NewOnusServer(&cfg.Server, core, pgstore)
	if err := srv.Run(ctx, cfg); err != nil {
		return fmt.Errorf("failed to run the Onus server: %w", err)
	}

	return nil
}
