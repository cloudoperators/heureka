package mariadb

import (
	"embed"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS
var Migration *embed.FS = &migrationFiles
