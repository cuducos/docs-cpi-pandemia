package unzip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/cuducos/docs-cpi-pandemia/bar"
	"github.com/cuducos/docs-cpi-pandemia/text"
)

type archive struct {
	zipFileNames []string
	dir          string
	target       string
	norm         bool
}

func (a archive) entryFile() (string, error) {
	if len(a.zipFileNames) == 0 {
		return "", errors.New("No ZIP file path for this archive")
	}

	return filepath.Join(a.dir, a.zipFileNames[0]), nil
}

func (a archive) createTarget() error {
	return os.MkdirAll(filepath.Join(a.dir, a.target), 0755)
}

func (a archive) remove() error {
	for _, p := range a.zipFileNames {
		err := os.Remove(filepath.Join(a.dir, p))
		if err != nil {
			return err
		}
	}

	return nil
}

func (a archive) unzipFile(z *zip.File) error {
	r, err := z.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	pth := filepath.Join(a.target, z.Name)
	if a.norm {
		pth, err = text.Normalize(pth)
		if err != nil {
			return err
		}
	}

	if !strings.HasPrefix(pth, filepath.Clean(a.target)+string(os.PathSeparator)) {
		return fmt.Errorf("Erro ao extrair arquivo em %s: %s", a.target, pth)
	}

	if z.FileInfo().IsDir() {
		os.MkdirAll(pth, 0755)
		return nil
	}

	os.MkdirAll(filepath.Dir(pth), 0755)
	w, err := os.OpenFile(pth, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Output(2, "Erro ao criar arquivo descompactado")
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	return err
}

func (a archive) unzip() error {
	pth, err := a.entryFile()
	if err != nil {
		return err
	}

	if err = a.createTarget(); err != nil {
		return err
	}

	r, err := zip.OpenReader(pth)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		err := a.unzipFile(f)
		if err != nil {
			return err
		}
	}

	return a.remove()
}

func getTargetDir(zip string) (string, error) {
	sequence := regexp.MustCompile("(?i)(.+\\D)(\\d+)(\\.zip)$")
	if sequence.MatchString(zip) {
		return sequence.FindStringSubmatch(zip)[1], nil
	}

	single := regexp.MustCompile("(?i)\\.zip$")
	if single.MatchString(zip) {
		return single.ReplaceAllString(zip, ""), nil
	}

	return "", fmt.Errorf("Cannot detect target directory name for %s", zip)
}

func getArchives(dir string) ([]archive, error) {
	var zips []string
	filepath.WalkDir(
		dir,
		func(pth string, entry fs.DirEntry, err error) error {
			if entry.IsDir() {
				return nil
			}
			if !strings.HasSuffix(strings.ToLower(pth), ".zip") {
				return nil
			}

			zips = append(zips, filepath.Base(pth))
			return nil
		},
	)
	sort.Strings(zips)

	var lastTarget string
	var batch []string
	var as []archive
	for i, zip := range zips {
		target, err := getTargetDir(zip)
		if err != nil {
			return nil, err
		}

		if target != lastTarget {
			as = append(
				as,
				archive{
					zipFileNames: batch,
					dir:          dir,
					target:       lastTarget,
					norm:         true,
				},
			)
			batch = []string{}
		}

		batch = append(batch, zip)
		lastTarget = target

		if i+1 == len(zips) {
			as = append(
				as,
				archive{
					zipFileNames: batch,
					dir:          dir,
					target:       target,
					norm:         true,
				},
			)
		}
	}

	return as, nil
}

func UnzipAll(dir string) error {
	as, err := getArchives(dir)
	if err != nil {
		return fmt.Errorf("Erro ao coletar os arquivos compactados: %s", err.Error())
	}

	queue := make(chan error)
	bar := bar.New(len(as), "arquivos", "Descompactando", 3)
	for _, a := range as {
		go func(a archive) {
			queue <- a.unzip()
		}(a)
	}

	for range as {
		if err = <-queue; err != nil {
			return err
		}
		bar.Add(1)
	}

	return nil
}
