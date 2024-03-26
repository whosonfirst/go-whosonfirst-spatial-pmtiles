package static

import (
	"embed"
)

//go:embed tangram/* javascript/*
var FS embed.FS
