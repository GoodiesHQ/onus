package onus

import "embed"

//go:embed sql/migrations/*.sql
var MigrationsFS embed.FS
