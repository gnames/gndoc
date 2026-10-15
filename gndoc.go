package gndoc

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gnames/gndoc/ent/convert"
	"github.com/gnames/gndoc/io/fsio"
	"github.com/gnames/gndoc/io/gnhttp"
	"github.com/gnames/gndoc/io/pdftext"
	"github.com/gnames/gndoc/io/tikaio"
)

// tikaTimeout limits the time of a Tika request.
var tikaTimeout = 15 * time.Second

// ErrUnsupported is returned when a format cannot be converted locally,
// and Tika service is not set.
var ErrUnsupported = convert.ErrUnsupported

type gndoc struct {
	// tika is nil if Tika URL is not set.
	tika       tikaio.Tika
	pdf        convert.PDF
	preferTika bool
	userAgent  string
}

// Option modifies a GNdoc instance.
type Option func(*gndoc)

// OptUserAgent sets the User-Agent header used when fetching URLs.
func OptUserAgent(ua string) Option {
	return func(d *gndoc) {
		d.userAgent = ua
	}
}

// OptPreferTika makes Tika the main converter for all formats, local
// conversion is used only if Tika fails (for example if the service is
// not reachable). It has no effect if Tika URL is empty.
func OptPreferTika(b bool) Option {
	return func(d *gndoc) {
		d.preferTika = b
	}
}

// OptPDFCacheDir sets a directory for caching compiled PDFium
// WebAssembly module. With the cache, only the first run of a program
// compiles the module, the following runs start PDF conversion faster.
// PDFium workers are shared by all GNdoc instances of a program, so the
// setting is process-wide, and it has an effect only if it is given
// before the first PDF conversion. If the cache cannot be used, PDFs
// are converted without it.
func OptPDFCacheDir(dir string) Option {
	return func(d *gndoc) {
		pdftext.SetCacheDir(dir)
	}
}

// New creates an instance of GNdoc. Common formats (plain texts in
// various encodings, HTML, XML, PDF, MS Office and OpenDocument files)
// are converted locally. If tikaURL is not empty, the Tika service is
// used as a fallback for formats that cannot be converted locally (for
// example images or scanned PDFs that require OCR). If tikaURL is empty,
// such formats return an error. To use Tika for all formats, add
// OptPreferTika option.
func New(tikaURL string, opts ...Option) GNdoc {
	d := &gndoc{pdf: pdftext.New()}
	if tikaURL != "" {
		d.tika = tikaio.New(tikaURL, tikaTimeout)
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// TextFromFile takes a path to a file, and returns the result of
// conversion or an error. If plainInput is true, the file is returned
// as is, without conversion.
func (d *gndoc) TextFromFile(path string, plainInput bool) (Result, error) {
	start := time.Now()
	data, err := fsio.ReadFile(path)
	if err != nil {
		return Result{}, err
	}

	if !plainInput {
		return d.toText(data, start)
	}

	res := Result{
		Text:            string(data),
		MIME:            "text/plain; charset=utf-8",
		Converter:       ConverterNone,
		OrigOffsets:     true,
		OrigByteOffsets: true,
		Duration:        since(start),
	}
	return res, nil
}

// TextFromURL takes a URL to an HTML page, reads its content, and
// converts it into a plain UTF8-encoded text. It returns the result of
// conversion or an error.
func (d *gndoc) TextFromURL(url string) (Result, error) {
	start := time.Now()
	h := gnhttp.New(
		gnhttp.OptUserAgent(d.userAgent),
		gnhttp.OptConnMax(10*time.Second),
		gnhttp.OptReqMax(60*time.Second),
	)

	_, mime, body, err := h.Get(url)
	if err != nil {
		return Result{}, err
	}

	if !strings.Contains(mime, "text/html") {
		return Result{}, fmt.Errorf("not an HTML text: %s", mime)
	}

	res, err := d.GetText(body)
	res.Duration = since(start)
	return res, err
}

// GetText takes a io.Reader interface (for example opened file)
// and returns the result of conversion of the input or an error.
func (d *gndoc) GetText(input io.Reader) (Result, error) {
	start := time.Now()
	data, err := io.ReadAll(input)
	if err != nil {
		return Result{}, err
	}
	return d.toText(data, start)
}

// toText converts data to text according to the settings.
func (d *gndoc) toText(data []byte, start time.Time) (Result, error) {
	var res Result
	var err error
	if d.tika != nil && d.preferTika {
		res, err = d.tikaFirst(data)
	} else {
		res, err = d.localFirst(data)
	}
	if err != nil {
		if d.tika == nil && errors.Is(err, convert.ErrUnsupported) {
			err = fmt.Errorf("%w (Tika URL is not set)", err)
		}
		return Result{}, err
	}
	res.Duration = since(start)
	return res, nil
}

// localFirst converts data locally, and uses Tika, if it is set, for
// formats that cannot be converted locally.
func (d *gndoc) localFirst(data []byte) (Result, error) {
	out, err := convert.Text(data, d.pdf)
	// scanned PDFs do not have a text layer and need OCR
	noText := err == nil && out.Text == "" &&
		strings.HasPrefix(out.MIME, "application/pdf")
	if d.tika == nil || (err == nil && !noText) {
		return localResult(out), err
	}

	ttxt, terr := d.tika.Parse(data)
	switch {
	case terr == nil:
		return tikaResult(ttxt, out.MIME), nil
	case err != nil:
		return Result{}, fmt.Errorf("local: %w; tika: %w", err, terr)
	default:
		return Result{}, fmt.Errorf("local: PDF has no text layer; tika: %w", terr)
	}
}

// tikaFirst converts data with Tika, and converts it locally if Tika
// fails.
func (d *gndoc) tikaFirst(data []byte) (Result, error) {
	txt, err := d.tika.Parse(data)
	if err == nil {
		return tikaResult(txt, convert.MIME(data)), nil
	}
	out, lerr := convert.Text(data, d.pdf)
	if lerr != nil {
		return Result{}, fmt.Errorf("tika: %w; local: %w", err, lerr)
	}
	return localResult(out), nil
}

func localResult(out convert.Output) Result {
	return Result{
		Text:            out.Text,
		MIME:            out.MIME,
		Converter:       ConverterLocal,
		OrigOffsets:     out.Plain,
		OrigByteOffsets: out.Verbatim,
	}
}

// tikaResult creates result of Tika conversion. Tika might modify any
// text, including plain texts, so offsets are not preserved.
func tikaResult(txt, mime string) Result {
	return Result{
		Text:      txt,
		MIME:      mime,
		Converter: ConverterTika,
	}
}

// since returns time passed from start in seconds.
func since(start time.Time) float32 {
	return float32(time.Since(start)) / float32(time.Second)
}
