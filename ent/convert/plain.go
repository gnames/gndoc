package convert

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

// plainText converts a text in UTF-8, UTF-16 or a legacy 8-bit encoding
// into UTF-8. Every character of the input becomes one character of
// the output, including a BOM (U+FEFF), so character offsets in the
// output match the original text. Invalid bytes of UTF-8 texts become
// one U+FFFD character each.
func plainText(data []byte) (string, error) {
	enc := detectEncoding(data)
	switch {
	case enc == nil:
		return string(data), nil
	case enc == charmap.Windows1252 && isMostlyUTF8(data):
		return repairUTF8(data), nil
	}
	res, err := enc.NewDecoder().Bytes(data)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

// detectEncoding returns nil for UTF-8 texts, or the encoding of the
// text otherwise.
func detectEncoding(data []byte) encoding.Encoding {
	switch {
	case bytes.HasPrefix(data, []byte("\xff\xfe")):
		// IgnoreBOM keeps BOM in the text as U+FEFF
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	case bytes.HasPrefix(data, []byte("\xfe\xff")):
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	case utf8.Valid(data):
		return nil
	}

	// UTF-16 without BOM has many zero bytes at either odd or even positions.
	var even, odd int
	n := len(data)
	if n > 8192 {
		n = 8192
	}
	for i := 0; i < n; i++ {
		if data[i] == 0 {
			if i%2 == 0 {
				even++
			} else {
				odd++
			}
		}
	}
	switch {
	case odd > n/4:
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	case even > n/4:
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	}

	// Windows-1252 is a superset of printable ISO-8859-1.
	return charmap.Windows1252
}

// isMostlyUTF8 checks if a text that is not a valid UTF-8 is still a
// UTF-8 text with some damaged bytes, and not a text in a legacy 8-bit
// encoding. Legacy encodings rarely produce valid UTF-8 multi-byte
// sequences, so if such sequences outnumber invalid bytes, the text is
// UTF-8.
func isMostlyUTF8(data []byte) bool {
	if bytes.HasPrefix(data, []byte("\xef\xbb\xbf")) {
		return true
	}
	var multi, invalid int
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			invalid++
		case size > 1:
			multi++
		}
		i += size
	}
	return multi > invalid
}

// repairUTF8 replaces every invalid byte with U+FFFD, keeping valid
// characters as they are.
func repairUTF8(data []byte) string {
	var b strings.Builder
	b.Grow(len(data) + 16)
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			b.WriteRune(utf8.RuneError)
		} else {
			b.Write(data[i : i+size])
		}
		i += size
	}
	return b.String()
}
