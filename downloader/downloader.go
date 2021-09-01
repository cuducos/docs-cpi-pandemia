package downloader

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cuducos/docs-cpi-pandemia/bar"
	"github.com/cuducos/docs-cpi-pandemia/cache"
	"github.com/cuducos/docs-cpi-pandemia/filesystem"
	"github.com/cuducos/docs-cpi-pandemia/text"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/schollz/progressbar/v3"
)

const url = "https://legis.senado.leg.br/comissoes/docsRecCPI?codcol=2441"
const prefix = "https://legis.senado.leg.br/sdleg-getter/documento/download/"

type settings struct {
	client    *retryablehttp.Client
	bar       *progressbar.ProgressBar
	cache     cache.Cache
	directory string
}

type result struct {
	url string
	err error
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

func getFileName(s *settings, u string) (string, error) {
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

	return text.Normalize(p[1])
}

func downloadFile(s *settings, u string) result {
	if s.cache.Exists(u) {
		return result{u, nil}
	}

	n, err := getFileName(s, u)
	if err != nil {
		return result{u, err}
	}

	f := filepath.Join(s.directory, n)
	t, err := os.Create(f)
	if err != nil {
		return result{u, err}
	}
	defer t.Close()

	r, err := s.client.Get(u)
	if err != nil {
		return result{u, err}
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return result{u, err}
	}

	_, err = io.Copy(t, r.Body)
	if err != nil {
		return result{u, err}
	}

	s.cache.Set(u)
	return result{u, nil}
}

func taskConsumer(s *settings, q chan string, errs chan error) {
	for url := range q {
		result := downloadFile(s, url)
		if result.err != nil {
			if result.err == io.EOF {
				q <- result.url
			} else {
				errs <- result.err
				s.bar.Add(1)
			}
		} else {
			s.bar.Add(1)
		}
	}
}

func queueConsumer(s *settings, q chan string, workers int) error {
	errs := make(chan error)
	for i := 0; i < workers; i++ {
		go taskConsumer(s, q, errs)
	}

	f, err := os.Create(filepath.Join(s.directory, "erros.txt"))
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)
	defer w.Flush()

	for err := range errs {
		_, err := w.WriteString(err.Error() + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func Download(d string, w, r uint, t time.Duration) error {
	log.Output(2, fmt.Sprintf("Colentando URLs para baixar…"))
	us, err := getUrls()
	if err != nil {
		return fmt.Errorf("Erro ao coletar as URLS: %s", err)
	}

	s := settings{
		retryablehttp.NewClient(),
		bar.New(len(us), "arquivos", "Baixando", 3),
		cache.Cache{Directory: d},
		d,
	}
	s.client.RetryMax = int(r)
	s.client.Logger = nil
	s.client.HTTPClient.Timeout = t

	if err := filesystem.CreateDir(s.directory); err != nil {
		return fmt.Errorf("Erro ao criar diretório %s: %s", s.directory, err)
	}

	q := make(chan string)
	for _, u := range us {
		go func(u string) { q <- u }(u)
	}

	log.Output(2, fmt.Sprintf("Começando a baixar os arquivos…"))
	return queueConsumer(&s, q, int(w))
}
