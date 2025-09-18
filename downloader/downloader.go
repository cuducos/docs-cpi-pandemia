package downloader

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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
	"golang.org/x/sync/errgroup"
)

const (
	url      = "https://legis.senado.leg.br/atividade/comissoes/comissao/2441/documentos-recebidos"
	prefix   = "https://legis.senado.leg.br/sdleg-getter/documento/download/"
	lastPage = 287
)

func getUrls(c *client) ([]string, error) {
	s := make(map[string]struct{})
	g, ctx := errgroup.WithContext(context.Background())
	ch := make(chan struct{})
	for p := 1; p <= lastPage; p++ {
		slog.Debug("requesting", "page", p)
		u := fmt.Sprintf("%s?p=%d", url, p)
		g.Go(func() error {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if err != nil {
				return fmt.Errorf("error creating request for %s: %w", u, err)
			}
			res, err := c.Do(req)
			if err != nil {
				return fmt.Errorf("error requesting %s: %w", u, err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected http response for %s: %s", u, res.Status)
			}
			d, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				return fmt.Errorf("error parsing html for %s: %w", u, err)
			}
			var c int
			d.Find("a").Each(func(_ int, a *goquery.Selection) {
				c++
				h, exist := a.Attr("href")
				if !exist {
					return
				}
				if strings.HasPrefix(h, prefix) {
					s[h] = struct{}{}
				}
			})
			ch <- struct{}{}
			return nil
		})
	}
	go func() {
		b := bar.New(lastPage, "páginas", "Buscando")
		for range lastPage {
			select {
			case <-ctx.Done():
				b.Close()
				return
			default:
				<-ch
				b.Add(1)
			}
		}
	}()
	if err := g.Wait(); err != nil {
		return nil, err
	}
	u := []string{}
	for k := range s {
		u = append(u, k)
	}
	return u, nil
}

type downloader struct {
	client    *client
	cache     cache.Cache
	directory string
}

func (d *downloader) getFileName(ctx context.Context, u string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
	if err != nil {
		return "", fmt.Errorf("error creating head request for %s: %w", u, err)
	}
	res, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error requesting file name for %s: %w", u, err)
	}
	v := res.Header.Get("Content-Disposition")
	if v == "" {
		return "", fmt.Errorf("error getting content disposition for %s", u)
	}
	p := strings.Split(v, "=")
	if len(p) != 2 {
		return "", fmt.Errorf("error parsing content disposition for %s", u)
	}
	return text.Normalize(strings.Trim(p[1], `"`))
}

func (d *downloader) downloadFile(ctx context.Context, u string) error {
	if d.cache.Exists(u) {
		return nil
	}
	n, err := d.getFileName(ctx, u)
	if err != nil {
		return fmt.Errorf("error getting file name for %s: %w", u, err)
	}
	f := filepath.Join(d.directory, n)
	t, err := os.Create(f)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", f, err)
	}
	defer t.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("error creating get request for %s: %w", u, err)
	}
	r, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("error requesting %s: %w", u, err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status for %s: %s", u, r.Status)
	}
	_, err = io.Copy(t, r.Body)
	if err != nil {
		return fmt.Errorf("error writing to %s: %w", f, err)
	}
	d.cache.Set(u)
	return nil
}

func (d *downloader) download(urls []string, crash bool) error {
	b := bar.New(len(urls), "arquivos", "Baixando")
	ok := make(chan struct{})
	errs := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, u := range urls {
		go func(u string) {
			err := d.downloadFile(ctx, u)
			if err != nil {
				errs <- err
			} else {
				ok <- struct{}{}
			}
		}(u)
	}
	for range urls {
		select {
		case err := <-errs:
			if crash {
				return err
			}
			slog.Error("download failed", err)
		case <-ok:
			b.Add(1)
		}
	}
	return nil
}

func newDownloader(c *client, dir string) (*downloader, error) {
	d := downloader{c, cache.Cache{Directory: dir}, dir}
	if err := filesystem.CreateDir(d.directory); err != nil {
		return nil, fmt.Errorf("error creating %s: %v", d.directory, err)
	}
	return &d, nil
}

func Download(dir string, conns, retries uint, timeout time.Duration, tolerant bool) error {
	c := newClient(conns, retries, timeout)
	urls, err := getUrls(c)
	if err != nil {
		return fmt.Errorf("error collecting the urls: %w", err)
	}
	d, err := newDownloader(c, dir)
	if err != nil {
		return err
	}
	return d.download(urls, !tolerant)
}
