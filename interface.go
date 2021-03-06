package gndoc

import "io"

// GNdoc is the main interface of GNdoc library for converting a great
// variety of files into UTF8-encoded tests.
type GNdoc interface {
	// TextFromFile takes a path to a file, boolean indicating if the file is a
	// plain UTF8-encoded text or not, and returns the converted
	// UTF8-encoded text, elapsed time in seconds or an error.
	TextFromFile(path string, plainInput bool) (string, float32, error)

	// TextFromURL takes a URL to a page, reads its content, and converts it into
	// a plain UTF8-encoded text. If it succeeds it returns the text, the time it
	// spend on conversion, and a nil.  If it does not succeed, it returns an
	// empty string and error.
	TextFromURL(url string) (string, float32, error)

	// GetText takes a io.Reader interface (for example opened file)
	// and returns back the UTF8-encoded textual content of the input.
	GetText(io.Reader) (string, error)

	// Text returns the UTF8-encoded textual content of a file, if it was
	// already received by running other methods.
	Text() string
}
