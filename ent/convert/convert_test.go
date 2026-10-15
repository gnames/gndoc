package convert_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gnames/gndoc/ent/convert"
	"github.com/stretchr/testify/assert"
)

func TestText(t *testing.T) {
	tests := []struct {
		msg, file, text, mime string
	}{
		{"abbr", "abbreviations.txt", "sister species", "text/plain"},
		{"ascii", "ascii.txt", "no unicode", "text/plain"},
		{"big", "big.txt", "Tacsonia insignis, 38", "text/plain"},
		{"json", "dirty_names.json", "Rhus folirsternatis", "application/json"},
		{"docx", "file.docx", "Global Names", "wordprocessingml"},
		{"html", "file.html", "Aesculus", "text/html"},
		{"xlsx", "file.xlsx", "Uhler, 1872", "spreadsheetml"},
		{"xlsx cells", "file.xlsx", "Criocoris\t\tapicalis", "spreadsheetml"},
		{"xml", "file.xml", "fishes of New Zealand", "xml"},
		{"italian", "italian.txt", "per il foro della", "text/plain"},
		{"latin1", "latin1.txt", "(Ujvárosi 2005)", "iso-8859-1"},
		{"utf16", "utf16.txt", "avons également reçu", "utf-16le"},
	}

	for _, v := range tests {
		data, err := os.ReadFile(filepath.Join("..", "..", "testdata", v.file))
		assert.Nil(t, err, v.msg)
		out, err := convert.Text(data, nil)
		assert.Nil(t, err, v.msg)
		assert.Contains(t, out.Text, v.text, v.msg)
		assert.Contains(t, out.MIME, v.mime, v.msg)
	}
}

// fakePDF is a PDF engine for tests.
type fakePDF struct {
	text string
	err  error
}

func (f fakePDF) PDFText([]byte) (string, error) {
	return f.text, f.err
}

func TestPDF(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "file.pdf"))
	assert.Nil(t, err)

	// PDF text is normalized
	pdf := fakePDF{text: "  Pardosa   moesta \n\n\n\nBanks, 1892  "}
	out, err := convert.Text(data, pdf)
	assert.Nil(t, err)
	assert.Equal(t, "Pardosa moesta\n\nBanks, 1892", out.Text)
	assert.Equal(t, "application/pdf", out.MIME)
	assert.False(t, out.Plain)

	// errors of PDF engine are returned
	_, err = convert.Text(data, fakePDF{err: errors.New("broken")})
	assert.EqualError(t, err, "broken")

	// without PDF engine PDFs are not supported
	out, err = convert.Text(data, nil)
	assert.ErrorIs(t, err, convert.ErrUnsupported)
	assert.Equal(t, "application/pdf", out.MIME)
}

func TestPlain(t *testing.T) {
	tests := []struct {
		msg   string
		input []byte
		text  string
	}{
		{"utf8", []byte("Pardosa  moesta\r\n"), "Pardosa  moesta\r\n"},
		{"utf8 bom", []byte("\xef\xbb\xbfPardosa"), "\ufeffPardosa"},
		{"utf16le bom", []byte("\xff\xfeP\x00a\x00"), "\ufeffPa"},
		{"utf16be bom", []byte("\xfe\xff\x00P\x00a"), "\ufeffPa"},
		{"latin1", []byte("Ujv\xe1rosi"), "Ujvárosi"},
		{"latin1 more", []byte("Ujv\xe1rosi \xe9t\xe9"), "Ujvárosi été"},
		// damaged UTF-8: each invalid byte becomes U+FFFD
		{"utf8 damaged", []byte("Ujvárosi é\xff é"), "Ujvárosi é\ufffd é"},
		{"utf8 damaged run", []byte("été\xff\xfe é"), "été\ufffd\ufffd é"},
		{"utf8 truncated", []byte("Ujvárosi é\xc3"), "Ujvárosi é\ufffd"},
		{"utf8 bom damaged", []byte("\xef\xbb\xbfab\xffc"), "\ufeffab\ufffdc"},
	}
	for _, v := range tests {
		out, err := convert.Text(v.input, nil)
		assert.Nil(t, err, v.msg)
		assert.Equal(t, v.text, out.Text, v.msg)
		assert.True(t, out.Plain, v.msg)
		assert.Equal(t, string(v.input) == out.Text, out.Verbatim, v.msg)
	}
}

func TestDamagedUTF8(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "utf8.txt")
	data, err := os.ReadFile(path)
	assert.Nil(t, err)
	orig := string(data)

	// insert an invalid byte in the middle of the text
	idx := strings.Index(orig, "Ujvárosi")
	assert.Greater(t, idx, 0)
	damaged := append([]byte(orig[:idx]), 0xff)
	damaged = append(damaged, orig[idx:]...)

	out, err := convert.Text(damaged, nil)
	assert.Nil(t, err)
	assert.True(t, out.Plain)
	assert.False(t, out.Verbatim)
	assert.Equal(t, orig[:idx]+"\ufffd"+orig[idx:], out.Text)
	// one character per invalid byte keeps character offsets
	assert.Equal(t, utf8.RuneCountInString(orig)+1,
		utf8.RuneCountInString(out.Text))
}

func TestHTML(t *testing.T) {
	html := `<html><head><title>Title</title><style>p {}</style></head>
<body><script>var a = 1;</script><p>Pardosa
   moesta</p><p>Banks,&nbsp;1892</p></body></html>`
	out, err := convert.Text([]byte(html), nil)
	assert.Nil(t, err)
	assert.Equal(t, "Title\n\nPardosa moesta\n\nBanks, 1892", out.Text)
	assert.False(t, out.Plain)
}

func TestUnsupported(t *testing.T) {
	for _, f := range []string{"binary", "image.jpg"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "testdata", f))
		assert.Nil(t, err)
		out, err := convert.Text(data, nil)
		assert.ErrorIs(t, err, convert.ErrUnsupported, f)
		assert.Empty(t, out.Text)
		assert.NotEmpty(t, out.MIME)
	}

	rtf := `{\rtf1\ansi{\fonttbl\f0\fswiss Helvetica;}\f0\pard Pardosa moesta\par}`
	_, err := convert.Text([]byte(rtf), nil)
	assert.ErrorIs(t, err, convert.ErrUnsupported, "rtf")
}

func TestEmpty(t *testing.T) {
	out, err := convert.Text(nil, nil)
	assert.Nil(t, err)
	assert.Empty(t, out.Text)
}
