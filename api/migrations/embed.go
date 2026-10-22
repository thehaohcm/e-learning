// Package migrations embeds the SQL migration files into the server binary.
// They are applied in filename order by db.Migrate on startup.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
