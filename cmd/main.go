package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/goodieshq/onus/internal/commands/migrate"
	"github.com/goodieshq/onus/internal/commands/serve"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

var Version string = "dev"

var app *cli.Command

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Stack().Caller().Logger()

	godotenv.Load()

	app = &cli.Command{
		Name:  "onus",
		Usage: "Onus ToDo application server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c", "cfg"},
				Usage:   "Onus configuration file path",
				Value:   "/app/onus.yml", // default path inside Docker container
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Run the Onus server",
				Action: func(ctx context.Context, c *cli.Command) error {
					return serve.Run(ctx, c)
				},
			},
			{
				Name:  "migrate",
				Usage: "Run all necessary database migrations",
				Commands: []*cli.Command{
					{
						Name:  "status",
						Usage: "Show the current migration version and whether the database is in a dirty state",
						Action: func(ctx context.Context, c *cli.Command) error {
							return migrate.Status(ctx, c)
						},
					},
					{
						Name:  "force",
						Usage: "Force set the migration version to clear dirty status (use with caution)",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:     "version",
								Aliases:  []string{"v"},
								Usage:    "Version to migrate to",
								Required: true,
								Validator: func(i int) error {
									if i < 0 {
										return fmt.Errorf("target version must be >= 0")
									}
									return nil
								},
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return migrate.Force(ctx, c, uint(c.Int("version")))
						},
					},
					{
						Name:  "to",
						Usage: "Migrate to a specific version",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:     "version",
								Aliases:  []string{"v"},
								Usage:    "Version to migrate to",
								Required: true,
								Validator: func(i int) error {
									if i < 0 {
										return fmt.Errorf("target version must be >= 0")
									}
									return nil
								},
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return migrate.To(ctx, c, uint(c.Int("version")))
						},
					},
					{
						Name:  "up",
						Usage: "Apply all up migrations",
						Flags: []cli.Flag{
							&cli.UintFlag{
								Name:    "steps",
								Aliases: []string{"s"},
								Value:   0,
								Usage:   "Number of up migration steps to apply",
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return migrate.Run(ctx, c, migrate.Up, c.Uint("steps"))
						},
					},
					{
						Name:  "down",
						Usage: "Apply down migrations",
						Flags: []cli.Flag{
							&cli.UintFlag{
								Name:    "steps",
								Aliases: []string{"s"},
								Value:   1,
								Usage:   "Number of down migration steps to apply",
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return migrate.Run(ctx, c, migrate.Down, c.Uint("steps"))
						},
					},
				},
			},
		},
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, os.Args); err != nil {
		if err == context.Canceled {
			return
		}
		log.Fatal().Err(err).Msg("application error")
	}
}
