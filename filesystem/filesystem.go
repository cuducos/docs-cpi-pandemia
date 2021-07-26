package filesystem

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cuducos/docs-cpi-pandemia/text"
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

func unzipFile(d string, z *zip.File, normalize bool) error {
	r, err := z.Open()
	if err != nil {
		log.Output(2, "Erro ao abrir arquivo dentro de um .zip")
		return err
	}
	defer r.Close()

	p := filepath.Join(d, z.Name)
	if normalize {
		p, err = text.Normalize(p)
		if err != nil {
			return err
		}
	}

	if !strings.HasPrefix(p, filepath.Clean(d)+string(os.PathSeparator)) {
		return fmt.Errorf("Erro ao extrair arquivo em %s: %s", d, p)
	}

	if z.FileInfo().IsDir() {
		os.MkdirAll(p, 0755)
		return nil
	}

	os.MkdirAll(filepath.Dir(p), 0755)
	w, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Output(2, "Erro ao criar arquivo descompactado")
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	return err
}

func Unzip(p string, normalize bool) error {
	r, err := zip.OpenReader(p)
	if err != nil {
		return err
	}
	defer r.Close()

	d := p[0 : len(p)-4]
	os.MkdirAll(d, 0755)

	for _, f := range r.File {
		err := unzipFile(d, f, normalize)
		if err != nil {
			return err
		}
	}

	return os.Remove(p)
}
