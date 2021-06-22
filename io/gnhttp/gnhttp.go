package gnhttp

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

type gnhttp struct {
	connMax time.Duration
	reqMax  time.Duration

	client http.Client
}

// New creates a new instance of GNhttp.
func New(opts ...Option) GNhttp {
	h := gnhttp{
		connMax: 1000 * time.Millisecond,
		reqMax:  1000 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(&h)
	}

	h.client = http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: h.connMax,
			}).DialContext,
		},
	}

	return &h
}

func (h *gnhttp) Get(url string) (int, string, io.Reader, error) {
	var code int
	var mime string
	var body bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), h.reqMax)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return code, mime, nil, err
	}

	res, err := h.client.Do(req)
	if res != nil {
		defer res.Body.Close()
		io.Copy(&body, res.Body)
		code = res.StatusCode
		mime = "mime"
	}
	return code, mime, &body, err
}

func (h *gnhttp) ConnMax(d time.Duration) {
	h.connMax = d
}

func (h *gnhttp) ReqMax(d time.Duration) {
	h.reqMax = d
}

// Option allows to modify GNhttp instance.
type Option func(GNhttp)

// OptOptConnMax sets the wait time for establishing a connection.
func OptConnMax(d time.Duration) Option {
	return func(h GNhttp) {
		h.ConnMax(d)
	}
}

// OptReqMax sets the wait time for processing a request.
func OptReqMax(d time.Duration) Option {
	return func(h GNhttp) {
		h.ReqMax(d)
	}
}
