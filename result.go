package gndoc

// Converter tells which converter created the text.
type Converter int

const (
	// ConverterNone means the input was returned as is, without
	// conversion (plainInput option of TextFromFile).
	ConverterNone Converter = iota
	// ConverterLocal means the input was converted locally.
	ConverterLocal
	// ConverterTika means the input was converted by Tika service.
	ConverterTika
)

// String returns the name of a converter.
func (c Converter) String() string {
	switch c {
	case ConverterNone:
		return "none"
	case ConverterLocal:
		return "local"
	case ConverterTika:
		return "tika"
	default:
		return "unknown"
	}
}

// Result contains the text extracted from an input, and information
// about how it was received.
type Result struct {
	// Text is the UTF8-encoded textual content of the input.
	Text string

	// MIME is the format of the input detected from its content
	// (for example "application/pdf", "text/plain; charset=utf-8").
	MIME string

	// Converter tells which converter created the text.
	Converter Converter

	// OrigOffsets is true when character (not byte) offsets in Text
	// match the original input. It happens when the input is a plain
	// text that was returned as is, or only re-encoded to UTF-8.
	// If it is false, the text was modified during conversion, and
	// offsets of strings found in Text are only valid for Text, so
	// clients should provide Text together with such offsets.
	OrigOffsets bool

	// OrigByteOffsets is true when byte offsets in Text match the
	// original input. It happens only when the input is a UTF-8 plain
	// text returned without changes. Texts re-encoded from UTF-16 or
	// Latin-1 keep character offsets (OrigOffsets is true), but their
	// byte offsets differ from the original.
	OrigByteOffsets bool

	// Duration is the time spent on conversion in seconds.
	Duration float32
}
