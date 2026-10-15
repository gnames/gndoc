// Package convert converts documents of common formats into UTF8-encoded
// text. Conversion is pure: it takes bytes and returns text without
// side effects. PDF documents need a PDF engine, it is provided by the
// caller via the PDF interface.
package convert

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

// ErrUnsupported is returned when a format cannot be converted locally.
var ErrUnsupported = errors.New("format is not supported locally")

var spaces = regexp.MustCompile(` {2,}`)

// Output is the result of a local conversion.
type Output struct {
	// Text is the UTF8-encoded textual content of the input.
	Text string
	// MIME is the detected format of the input.
	MIME string
	// Plain is true if the input is a plain text. Such text is only
	// converted to UTF-8, characters stay the same as in the input.
	// Other formats get their white spaces normalized.
	Plain bool
	// Verbatim is true if the input is a UTF-8 plain text, and Text is
	// byte-to-byte identical to the input.
	Verbatim bool
}

// PDF extracts the text layer of PDF documents.
type PDF interface {
	// PDFText returns the text of a PDF document. Lines of the document
	// are kept, hyphens at the end of lines are kept as well.
	PDFText(data []byte) (string, error)
}

// MIME detects the format of the input data.
func MIME(data []byte) string {
	return mimetype.Detect(data).String()
}

// Text detects the format of the input data and converts it to text.
// The MIME type is set in the output even if conversion fails. If the
// format cannot be processed locally, the error wraps ErrUnsupported.
// PDF documents are converted by pdf, if it is nil, they are not
// supported.
func Text(data []byte, pdf PDF) (Output, error) {
	if len(data) == 0 {
		return Output{MIME: "text/plain", Plain: true, Verbatim: true}, nil
	}

	mt := mimetype.Detect(data)
	mime := mt.String()

	var txt string
	var plain bool
	var err error
	switch {
	case is(mt, "application/pdf"):
		txt, err = pdfText(data, pdf, mime)
	case is(mt, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"):
		txt, err = docxText(data)
	case is(mt, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"):
		txt, err = xlsxText(data)
	case is(mt, "application/vnd.openxmlformats-officedocument.presentationml.presentation"):
		txt, err = pptxText(data)
	case is(mt, "application/vnd.oasis.opendocument.text"),
		is(mt, "application/vnd.oasis.opendocument.spreadsheet"),
		is(mt, "application/vnd.oasis.opendocument.presentation"):
		txt, err = odfText(data)
	case is(mt, "text/html"):
		txt, err = htmlText(data)
	case is(mt, "text/xml"), is(mt, "application/xml"):
		txt, err = xmlText(data)
	// RTF is detected as a text, but it needs a real parser.
	case is(mt, "text/rtf"):
		err = fmt.Errorf("%w: %s", ErrUnsupported, mime)
	case is(mt, "text/plain"):
		txt, err = plainText(data)
		plain = true
	default:
		// mimetype might not recognize texts in legacy encodings
		if looksLikeText(data) {
			txt, err = plainText(data)
			plain = true
		} else {
			err = fmt.Errorf("%w: %s", ErrUnsupported, mime)
		}
	}
	if err != nil {
		return Output{MIME: mime}, err
	}
	out := Output{
		Text:     txt,
		MIME:     mime,
		Plain:    plain,
		Verbatim: plain && txt == string(data),
	}
	return out, nil
}

// pdfText converts a PDF document with the given PDF engine.
func pdfText(data []byte, pdf PDF, mime string) (string, error) {
	if pdf == nil {
		return "", fmt.Errorf("%w: %s", ErrUnsupported, mime)
	}
	txt, err := pdf.PDFText(data)
	if err != nil {
		return "", err
	}
	return normalizeSpace(txt), nil
}

// is checks if MIME type is the given one, or inherits from it.
func is(mt *mimetype.MIME, mime string) bool {
	for m := mt; m != nil; m = m.Parent() {
		if m.Is(mime) {
			return true
		}
	}
	return false
}

// normalizeSpace trims lines, collapses runs of spaces inside of lines,
// and collapses runs of empty lines.
func normalizeSpace(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	lines := strings.Split(s, "\n")
	var b strings.Builder
	empty := 0
	for _, l := range lines {
		l = strings.Trim(l, " \t\r")
		l = spaces.ReplaceAllString(l, " ")
		if l == "" {
			empty++
			if empty > 1 {
				continue
			}
		} else {
			empty = 0
		}
		b.WriteString(l)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

// looksLikeText checks that data has no control characters typical for
// binary formats.
func looksLikeText(data []byte) bool {
	sample := data
	if len(sample) > 8192 {
		sample = sample[:8192]
	}
	return !bytes.ContainsAny(sample, "\x00\x01\x02\x03\x04\x05\x06\x07\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f")
}
