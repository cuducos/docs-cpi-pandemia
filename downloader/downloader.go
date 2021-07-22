package downloader

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cuducos/docs-cpi-pandemia/cache"
	"github.com/cuducos/docs-cpi-pandemia/filesystem"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/semaphore"
)

const url = "https://legis.senado.leg.br/comissoes/docsRecCPI?codcol=2441"
const prefix = "https://legis.senado.leg.br/sdleg-getter/documento/download/"

type settings struct {
	client    *retryablehttp.Client
	semaphore *semaphore.Weighted
	bar       *progressbar.ProgressBar
	cache     cache.Cache
	directory string
}

func getUrls() ([]string, error) {
	s := make(map[string]struct{})
	d, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	d.Find("a").Each(func(_ int, a *goquery.Selection) {
		h, exist := a.Attr("href")
		if !exist {
			return
		}

		if strings.HasPrefix(h, prefix) {
			s[h] = struct{}{}
		}
	})

	u := []string{}
	for k := range s {
		u = append(u, k)
	}
	return u, nil
}

func getFileInfo(s *settings, u string) (string, error) {
	r, err := s.client.Head(u)
	if err != nil {
		return "", err
	}

	e := fmt.Errorf("Erro ao identificar nome de arquivo para %s", u)
	v := r.Header.Get("Content-Disposition")
	if v == "" {
		return "", e
	}

	p := strings.Split(v, "=")
	if len(p) != 2 {
		return "", e
	}

	return p[1], nil
}

func unarchive(p string) error {
	if !strings.HasSuffix(strings.ToLower(p), ".zip") {
		return nil
	}

	return filesystem.Unzip(p)
}

func downloadFile(s *settings, u string) {
	defer s.semaphore.Release(1)
	defer s.bar.Add(1)

	if s.cache.Exists(u) {
		return
	}

	n, err := getFileInfo(s, u)
	if err != nil {
		log.Output(2, err.Error())
		return
	}

	f := filepath.Join(s.directory, n)
	t, err := os.Create(f)
	if err != nil {
		log.Output(2, fmt.Sprintf("Erro ao criar arquivo %s: %s", f, err.Error()))
		return
	}
	defer t.Close()

	r, err := s.client.Get(u)
	if err != nil {
		log.Output(2, fmt.Sprintf("Erro ao baixar arquivo %s: %s", u, err.Error()))
		return
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		log.Output(2, fmt.Sprintf("Erro ao baixar arquivo %s: HTTP status %d", u, r.StatusCode))
		return
	}

	_, err = io.Copy(t, r.Body)
	if err != nil {
		log.Output(2, fmt.Sprintf("Erro ao baixar arquivo %s: %s", u, err.Error()))
		return
	}

	if err := unarchive(f); err != nil {
		log.Output(2, fmt.Sprintf("Erro ao descompactar arquivo %s: %s", f, err.Error()))
		return
	}

	s.cache.Set(u)
}

func Download(d string, w, r uint) {
	log.Output(2, fmt.Sprintf("Colentando URLs para baixar…"))
	us, err := getUrls()
	if err != nil {
		log.Output(2, fmt.Sprintf("Erro ao coletar as URLS: %s", err.Error()))
		os.Exit(1)
	}

	s := settings{
		retryablehttp.NewClient(),
		semaphore.NewWeighted(int64(w)),
		progressbar.NewOptions(
			len(us),
			progressbar.OptionSetItsString("arquivos"),
			progressbar.OptionSetDescription("Baixando arquivos"),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionThrottle(3*time.Second),
			// default values
			progressbar.OptionSetWidth(10),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
		),
		cache.Cache{Directory: d},
		d,
	}
	s.client.RetryMax = int(r)
	s.client.Logger = nil

	if err := filesystem.CreateDir(s.directory); err != nil {
		log.Output(2, fmt.Sprintf("Erro ao criar diretório %s: %s", s.directory, err.Error()))
		os.Exit(1)
	}

	log.Output(2, fmt.Sprintf("Começando a baixar os arquivos…"))
	for _, u := range us {
		s.semaphore.Acquire(context.Background(), 1)
		go downloadFile(&s, u)
	}
}
