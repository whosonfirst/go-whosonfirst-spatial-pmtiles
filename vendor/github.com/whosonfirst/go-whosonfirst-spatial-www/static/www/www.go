package www

import (
	"embed"
)

//go:embed css/* javascript/* intersects/* point-in-polygon/* *.html
var FS embed.FS
