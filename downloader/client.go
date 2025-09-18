package downloader

import (
	"log/slog"
	"math"
	"net/http"
	"time"
)

type client struct {
	client  *http.Client
	retries uint
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	var t uint
	for {
		r, err := c.client.Do(req)
		if err != nil || r.StatusCode == http.StatusTooManyRequests {
			t++
			if t >= c.retries {
				if err != nil {
					slog.Error("could not download", "url", req.URL, "error", err)
				} else {
					slog.Error("could not download", "url", req.URL, "status", r.Status)
				}
			}
			w := time.Duration(int(math.Exp2(float64(t)))) * time.Second
			slog.Debug("failed request waiting for retry", "url", req.URL, "wait", w, "error", err)
			time.Sleep(w)
			continue
		}
		return r, nil
	}
}

func newClient(conns, retries uint, timeout time.Duration) *client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxConnsPerHost = int(conns)
	t.MaxIdleConnsPerHost = int(conns)
	c := http.Client{Transport: t}
	c.Timeout = timeout
	return &client{&c, retries}
}
