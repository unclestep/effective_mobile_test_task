package em

import (
	"embed"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

//go:embed config/*.yml
var ConfigFS embed.FS
