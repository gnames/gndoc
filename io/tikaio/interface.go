package tikaio

// Tika sends documents to an Apache Tika service for conversion.
type Tika interface {
	// Parse sends a document to Tika and returns its UTF8-encoded text.
	Parse(data []byte) (string, error)
}
