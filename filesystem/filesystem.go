package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

func CreateDir(d string) error {
	_, err := os.Stat(d)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(d, 0755)
		}
		return err
	}

	return nil
}

func CreateFile(p string) error {
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}

	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		f, e := os.Create(p)
		if e != nil {
			return e
		}
		defer f.Close()
	}
	return err
}

func CleanDir(r string) error {
	remove := func(p string, d fs.DirEntry, err error) error {
		if p == r {
			return nil
		}
		return os.RemoveAll(p)
	}
	return filepath.WalkDir(r, remove)
}

func Exists(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}
