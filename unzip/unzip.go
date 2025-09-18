package unzip

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cuducos/docs-cpi-pandemia/bar"
	"github.com/cuducos/docs-cpi-pandemia/text"
	"golang.org/x/sync/errgroup"
)

type archive struct {
	zips   []string
	dir    string
	target string
}

func (a archive) createTarget() error {
	p := filepath.Join(a.dir, a.target)
	if err := os.MkdirAll(p, 0755); err != nil {
		return fmt.Errorf("could not create target directory %s: %w", p, err)
	}
	return nil
}

func (a archive) remove() error {
	g := new(errgroup.Group)
	for _, p := range a.zips {
		g.Go(func() error { return os.Remove(filepath.Join(a.dir, p)) })
	}
	return g.Wait()
}

func (a archive) unzipFile(z *zip.File) error {
	r, err := z.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	pth := filepath.Join(a.dir, a.target, z.Name)
	pth, err = text.Normalize(pth)
	if err != nil {
		return err
	}
	if z.FileInfo().IsDir() {
		if err := os.MkdirAll(pth, 0755); err != nil {
			return fmt.Errorf("could not create subdirectory %s: %w", pth, err)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(pth), 0755); err != nil {
		return fmt.Errorf("could not create parent directory %s: %w", pth, err)
	}
	w, err := os.OpenFile(pth, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		slog.Error("error unarchiving", "file", pth)
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, r)
	return err
}

func (a archive) unzip() error {
	if len(a.zips) == 0 {
		slog.Debug("no source files", "archive", a)
		return nil
	}
	if err := a.createTarget(); err != nil {
		return err
	}
	g := new(errgroup.Group)
	for _, z := range a.zips {
		g.Go(func() error {
			pth := filepath.Join(a.dir, z)
			r, err := zip.OpenReader(pth)
			if err != nil {
				return err
			}
			defer r.Close()
			sg := new(errgroup.Group)
			for _, f := range r.File {
				sg.Go(func() error { return a.unzipFile(f) })
			}
			return sg.Wait()
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return a.remove()
}

func getTargetDir(zip string) (string, error) {
	seq := regexp.MustCompile(`(?i)(.+\D)(\d+)(\.zip)$`)
	if seq.MatchString(zip) {
		return seq.FindStringSubmatch(zip)[1], nil
	}
	one := regexp.MustCompile(`(?i)\.zip$`)
	if one.MatchString(zip) {
		return one.ReplaceAllString(zip, ""), nil
	}
	return "", fmt.Errorf("cannot detect target directory name for %s", zip)
}

func getArchives(dir string) ([]archive, error) {
	c := filepath.Join(dir, ".cache")
	var zs []string
	filepath.WalkDir(
		dir,
		func(pth string, entry fs.DirEntry, err error) error {
			if entry.IsDir() || strings.HasPrefix(pth, c) || !strings.HasSuffix(strings.ToLower(pth), ".zip") {
				return nil
			}
			zs = append(zs, filepath.Base(pth))
			return nil
		},
	)
	m := make(map[string]archive)
	for _, z := range zs {
		target, err := getTargetDir(z)
		if err != nil {
			return nil, err
		}
		a := m[target]
		if len(a.zips) == 0 {
			a.dir = dir
			a.target = target
		}
		a.zips = append(a.zips, z)
		m[target] = a
	}
	var as []archive
	for _, a := range m {
		as = append(as, a)
	}
	return as, nil
}

func UnzipAll(dir string) error {
	as, err := getArchives(dir)
	if err != nil {
		return fmt.Errorf("error collecting archived files: %w", err)
	}
	ch := make(chan error)
	bar := bar.New(len(as), "arquivos", "Descompactando")
	for _, a := range as {
		go func(a archive) {
			ch <- a.unzip()
		}(a)
	}
	for range as {
		if err = <-ch; err != nil {
			slog.Error("could not unzip", "error", err)
		}
		bar.Add(1)
	}
	return nil
}
