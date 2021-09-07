package gndoc

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gnames/gndoc/io/gnhttp"
	"github.com/gnames/gnsys"
	"github.com/google/go-tika/tika"
)

var timeout = 15 * time.Second

type gndoc struct {
	tclient *tika.Client
	text    string
}

func New(tikaURL string) GNdoc {
	tclient := tika.NewClient(nil, tikaURL)
	return &gndoc{
		tclient: tclient,
	}
}

// TextFromFile takes a path to a file, and returns the converted
// UTF8-encoded text, elapsed time in seconds or an error.
func (d *gndoc) TextFromFile(
	path string,
	plainInput bool,
) (string, float32, error) {
	var err error
	var bs []byte
	var txt string
	var dur float32

	start := time.Now()
	exists, err := gnsys.FileExists(path)
	if err != nil {
		return "", dur, err
	}
	if !exists {
		return "", dur, fmt.Errorf("file '%s' does not exist", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return "", dur, err
	}
	defer f.Close()
	if plainInput {
		bs, err = io.ReadAll(f)
		txt = string(bs)
		dur = float32(time.Since(start)) / float32(time.Second)
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
	dur = float32(time.Since(start)) / float32(time.Second)
	return txt, dur, nil
}

// TextFromURL takes a URL to a page, reads its content, and converts it into
// a plain UTF8-encoded text. If it succeeds it returns the text, the time it
// spend on conversion, and a nil.  If it does not succeed, it returns an
// empty string and error.
func (d *gndoc) TextFromURL(url string) (string, float32, error) {
	var dur float32
	var err error
	start := time.Now()
	h := gnhttp.New()

	_, mime, body, err := h.Get(url)
	if err != nil {
		return "", dur, err
	}

	if !strings.Contains(mime, "text/html") {
		err = fmt.Errorf("not an HTML text: %s", mime)
		return "", dur, err
	}

	res, err := d.GetText(body)
	if err != nil {
		return "", dur, err
	}

	dur = float32(time.Now().Sub(start)) / float32(time.Second)
	return res, dur, nil
}

// GetText takes a io.Reader interface (for example opened file)
// and returns back the UTF8-encoded textual content of the input.
func (d *gndoc) GetText(input io.Reader) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	txt, err := d.tclient.Parse(ctx, input)
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
