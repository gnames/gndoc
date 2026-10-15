package gndoc

import "io"

// GNdoc is the main interface of GNdoc library for converting a great
// variety of files into UTF8-encoded texts. It is safe for concurrent
// use by multiple goroutines.
type GNdoc interface {
	// TextFromFile takes a path to a file, boolean indicating if the file is a
	// plain UTF8-encoded text or not, and returns the result of conversion
	// or an error. Plain UTF8-encoded texts are returned as is.
	TextFromFile(path string, plainInput bool) (Result, error)

	// TextFromURL takes a URL to an HTML page, reads its content, and converts
	// it into a plain UTF8-encoded text. It returns the result of conversion
	// or an error.
	TextFromURL(url string) (Result, error)

	// GetText takes a io.Reader interface (for example opened file)
	// and returns the result of conversion of the input or an error.
	GetText(io.Reader) (Result, error)
}
