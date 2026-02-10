package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

func GetDistFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}

func HasAssets() bool {
	entries, err := distFS.ReadDir("dist")
	if err != nil {
		return false
	}
	return len(entries) > 0
}
