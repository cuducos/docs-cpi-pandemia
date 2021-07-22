package cache

import (
	"path/filepath"
	"strings"

	"github.com/cuducos/docs-cpi-pandemia/filesystem"
)

type Cache struct {
	Directory string
}

func (c Cache) pathFromUrl(u string) string {
	p := strings.Split(u, "/")
	d := strings.Split(p[len(p)-1], "-")
	return filepath.Join(c.Directory, ".cache", filepath.Join(d...))
}

func (c Cache) Set(u string) {
	filesystem.CreateFile(c.pathFromUrl(u))
}

func (c Cache) Exists(u string) bool {
	return filesystem.Exists(c.pathFromUrl(u))
}
