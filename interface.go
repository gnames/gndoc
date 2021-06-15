package gndoc

import "io"

// GNdoc is the main interface of GNdoc library for converting a great
// variety of files into UTF8-encoded tests.
type GNdoc interface {
	// TextFromFile takes a path to a file, and returns the converted
	// UTF8-encoded text, elapsed time in seconds or an error.
	TextFromFile(path string) (string, float32, error)

	// GetText takes a io.Reader interface (for example opened file)
	// and returns back the UTF8-encoded textual content of the input.
	GetText(io.Reader) (string, error)

	// Text returns the UTF8-encoded textual content of a file, if it was
	// already received by running other methods.
	Text() string
}
