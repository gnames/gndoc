package doc

import (
	"context"
	"io"
	"time"

	"github.com/google/go-tika/tika"
)

var timeout = 5 * time.Second

type doc struct {
	client   *tika.Client
	meta     string
	language string
	text     string
}

func NewDoc(tikaURL string) Doc {
	client := tika.NewClient(nil, tikaURL)
	return &doc{
		client: client,
	}
}

func (d *doc) GetMeta(input io.Reader) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	meta, err := d.client.Meta(ctx, input)
	if err == nil {
		d.meta = meta
	}
	return meta, err
}

func (d *doc) Meta() string {
	return d.meta
}

func (d *doc) GetLanguage(input io.Reader) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	l, err := d.client.Language(ctx, input)
	if err == nil {
		d.language = l
	}
	return l, err
}

func (d *doc) Language() string {
	return d.language
}

func (d *doc) GetContent(input io.Reader) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	txt, err := d.client.Parse(ctx, input)
	if err == nil {
		d.text = txt
	}
	return txt, err
}

func (d *doc) Content() string {
	return d.text
}
