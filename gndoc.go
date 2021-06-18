package gndoc

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gnames/gnsys"
	"github.com/google/go-tika/tika"
)

var timeout = 5 * time.Second

type gndoc struct {
	client *tika.Client
	text   string
}

func New(tikaURL string) GNdoc {
	client := tika.NewClient(nil, tikaURL)
	return &gndoc{
		client: client,
	}
}

// TextFromFile takes a path to a file, and returns the converted
// UTF8-encoded text, elapsed time in seconds or an error.
func (d *gndoc) TextFromFile(
	path string,
	plainInput bool,
) (string, float32, error) {
	var err error
	var txt string
	var dur float32

	start := time.Now()
	exists, err := gnsys.FileExists(path)
	if err != nil {
		return "", dur, err
	}
	if !exists {
		return "", dur, fmt.Errorf("File '%s' does not exist", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return "", dur, err
	}
	defer f.Close()
	if plainInput {
		bs, err := io.ReadAll(f)
		txt = string(bs)
		dur = float32(time.Now().Sub(start)) / float32(time.Second)
		if err != nil {
			return "", dur, err
		}
		return txt, dur, nil
	} else {
		txt, err = d.GetText(f)
		if err != nil {
			return "", dur, err
		}
	}
	dur = float32(time.Now().Sub(start)) / float32(time.Second)
	return txt, dur, nil
}

// GetText takes a io.Reader interface (for example opened file)
// and returns back the UTF8-encoded textual content of the input.
func (d *gndoc) GetText(input io.Reader) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	txt, err := d.client.Parse(ctx, input)
	if err == nil {
		d.text = txt
	}
	return txt, err
}

// Text returns the UTF8-encoded textual content of a file, if it was
// already received by running other methods.
func (d *gndoc) Text() string {
	return d.text
}
