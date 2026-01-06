package migrate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/goodieshq/onus/sql/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

// DO NOT CHANGE - hex value of "onusonus"
const postgresLockID = int64(8029484324299371891)

// withPostgresLock acquires a PostgreSQL advisory lock for the duration of the provided function.
func withPostgresLock(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
	ctxLock, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	conn, err := pool.Acquire(ctxLock)
	if err != nil {
		return fmt.Errorf("failed to acquire a database connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctxLock, "SELECT pg_advisory_lock($1)", postgresLockID); err != nil {
		return fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		_, _ = conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", postgresLockID)
	}()

	return fn(ctxLock)
}

func newMigratorPostgres(ctx context.Context, pool *pgxpool.Pool) (*migrate.Migrate, error) {
	sub, err := fs.Sub(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to create sub filesystem for migrations: %w", err)
	}

	src, err := iofs.New(sub, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to create iofs source for migrations: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create postgres migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

func newPostgresPool(ctx context.Context, dsn string) (*pgxpool.Pool, func(), error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	return pool, pool.Close, nil
}

func runMigrateForcePostgres(ctx context.Context, dsn string, version uint) error {
	pool, close, err := newPostgresPool(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}
	defer close()

	return withPostgresLock(ctx, pool, func(ctx context.Context) error {
		m, err := newMigratorPostgres(ctx, pool)
		if err != nil {
			return fmt.Errorf("failed to create migrator: %w", err)
		}
		defer migrateCloser(m)()

		if err := m.Force(int(version)); err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				return fmt.Errorf("failed to force migrate to version %d: %w", version, err)
			}
		}
		log.Info().Uint("version", version).Msg("forced database migration version")
		return nil
	})
}

func runMigrateToPostgres(ctx context.Context, dsn string, to uint) error {
	pool, close, err := newPostgresPool(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}
	defer close()

	return withPostgresLock(ctx, pool, func(ctx context.Context) error {
		m, err := newMigratorPostgres(ctx, pool)
		if err != nil {
			return fmt.Errorf("failed to create migrator: %w", err)
		}
		defer migrateCloser(m)()

		if v, dirty, verr := currentVersion(m); verr == nil {
			log.Info().Uint("version", v).Bool("dirty", dirty).Msg("migration status (before)")
			if dirty {
				return fmt.Errorf("database is in a dirty state at version %d", v)
			}
		}

		if to == 0 {
			if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return fmt.Errorf("failed to migrate to version %d: %w", to, err)
			}
		} else {
			if err := m.Migrate(to); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return fmt.Errorf("failed to migrate to version %d: %w", to, err)
			}
		}

		if v, dirty, verr := currentVersion(m); verr == nil {
			log.Info().Uint("version", v).Bool("dirty", dirty).Msg("migration status (after)")
		}
		return nil
	})
}

func runMigrationsPostgres(ctx context.Context, dsn string, dir OnusMigrateDirection, steps uint) error {
	pool, close, err := newPostgresPool(ctx, dsn)
	if close != nil {
		defer close()
	}
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}

	return withPostgresLock(ctx, pool, func(ctx context.Context) error {
		m, err := newMigratorPostgres(ctx, pool)
		if err != nil {
			return fmt.Errorf("failed to create migrator: %w", err)
		}

		defer migrateCloser(m)()

		if v, dirty, verr := currentVersion(m); verr == nil {
			log.Info().Uint("version", v).Bool("dirty", dirty).Msg("migration status (before)")
			if dirty {
				return fmt.Errorf("database is in a dirty state at version %d", v)
			}
		}

		switch dir {
		case Up:
			if steps > 0 {
				if err := m.Steps(int(steps)); err != nil && !errors.Is(err, migrate.ErrNoChange) {
					return fmt.Errorf("failed to apply up migrations: %w", err)
				}
			} else {
				if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
					return fmt.Errorf("failed to apply up migrations: %w", err)
				}
			}
		case Down:
			if steps <= 0 {
				return fmt.Errorf("steps must be greater than >= 1 for down migrations")
			}
			if err := m.Steps(-int(steps)); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return fmt.Errorf("failed to apply down migrations: %w", err)
			}
		default:
			return fmt.Errorf("unknown migration direction: %d", dir)
		}
		if v, dirty, verr := currentVersion(m); verr == nil {
			log.Info().Uint("version", v).Bool("dirty", dirty).Msg("migration status (after)")
		}
		return nil
	})
}
