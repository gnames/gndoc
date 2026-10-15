// Package tikaio converts documents to text using a remote Apache Tika
// service.
package tikaio

import (
	"bytes"
	"context"
	"time"

	"github.com/google/go-tika/tika"
)

type tikaio struct {
	client  *tika.Client
	timeout time.Duration
}

// New creates a Tika client for the service at url. Every request is
// limited by timeout. No connection is made until the first request.
func New(url string, timeout time.Duration) Tika {
	return &tikaio{
		client:  tika.NewClient(nil, url),
		timeout: timeout,
	}
}

// Parse sends a document to Tika and returns its UTF8-encoded text.
func (t *tikaio) Parse(data []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()
	return t.client.Parse(ctx, bytes.NewReader(data))
}
