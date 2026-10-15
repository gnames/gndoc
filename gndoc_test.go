package gndoc_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gnames/gndoc"
	"github.com/stretchr/testify/assert"
)

var tikaURL = "https://tika.globalnames.org"

func TestDoc(t *testing.T) {
	tests := []struct {
		msg, file, text, mime string
		origOffsets           bool
	}{
		{"abbr", "abbreviations.txt", "sister species", "text/plain", true},
		{"ascii", "ascii.txt", "no unicode", "text/plain", true},
		{"big", "big.txt", "Tacsonia insignis, 38", "text/plain", true},
		{"json", "dirty_names.json", "Rhus folirsternatis", "application/json", true},
		{"ms doc", "file.docx", "Global Names", "wordprocessingml", false},
		{"html", "file.html", "Aesculus", "text/html", false},
		{"pdf", "file.pdf", "sabana de Bogotá", "application/pdf", false},
		{"xlsx", "file.xlsx", "Uhler, 1872", "spreadsheetml", false},
		{"xml", "file.xml", "fishes of New Zealand", "xml", false},
		{"italian", "italian.txt", "per il foro della", "text/plain", true},
		{"latin1", "latin1.txt", "(Ujvárosi 2005)", "iso-8859-1", true},
		{"utf16", "utf16.txt", "avons également reçu", "utf-16le", true},
	}

	// Tika is not needed for these formats.
	d := gndoc.New("")
	for _, v := range tests {
		f, err := os.Open(filepath.Join("testdata", v.file))
		assert.Nil(t, err)
		res, err := d.GetText(f)
		assert.Nil(t, err, v.msg)
		assert.Contains(t, res.Text, v.text, v.msg)
		assert.Contains(t, res.MIME, v.mime, v.msg)
		assert.Equal(t, gndoc.ConverterLocal, res.Converter, v.msg)
		assert.Equal(t, v.origOffsets, res.OrigOffsets, v.msg)
		f.Close()
	}
}

// TestOrigOffsets checks that plain texts keep the same characters as
// the original input, so offsets of found strings match the input.
func TestOrigOffsets(t *testing.T) {
	d := gndoc.New("")
	tests := []struct {
		file      string
		origBytes bool
	}{
		{"utf8.txt", true},
		{"italian.txt", true},
		{"utf16.txt", false},
		{"latin1.txt", false},
	}
	for _, v := range tests {
		path := filepath.Join("testdata", v.file)
		res, err := d.TextFromFile(path, false)
		assert.Nil(t, err, v.file)
		assert.True(t, res.OrigOffsets, v.file)
		assert.Equal(t, v.origBytes, res.OrigByteOffsets, v.file)

		// UTF-8 texts are returned byte-to-byte, including CRLF
		if v.origBytes {
			data, err := os.ReadFile(path)
			assert.Nil(t, err)
			assert.Equal(t, string(data), res.Text, v.file)
		}
	}

	// converted formats do not keep any offsets
	res, err := d.TextFromFile(filepath.Join("testdata", "file.pdf"), false)
	assert.Nil(t, err)
	assert.False(t, res.OrigOffsets)
	assert.False(t, res.OrigByteOffsets)

	// damaged UTF-8 keeps character offsets, but not byte offsets
	res, err = d.GetText(strings.NewReader("Ujvárosi é\xff é"))
	assert.Nil(t, err)
	assert.Equal(t, "Ujvárosi é\ufffd é", res.Text)
	assert.True(t, res.OrigOffsets)
	assert.False(t, res.OrigByteOffsets)

	// UTF-8 BOM stays as U+FEFF
	res, err = d.GetText(strings.NewReader("\ufeffPardosa moesta"))
	assert.Nil(t, err)
	assert.Equal(t, "\ufeffPardosa moesta", res.Text)

	// UTF-16 BOM stays as U+FEFF
	res, err = d.TextFromFile(filepath.Join("testdata", "utf16.txt"), false)
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(res.Text, "\ufeff"))
}

// TestConcurrent checks that one GNdoc instance can be used by several
// goroutines. Run with -race to detect data races.
func TestConcurrent(t *testing.T) {
	d := gndoc.New("")
	files := []string{"file.pdf", "no_names.pdf", "file.docx", "utf8.txt"}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		file := files[i%len(files)]
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := d.TextFromFile(filepath.Join("testdata", file), false)
			assert.Nil(t, err, file)
			assert.NotEmpty(t, res.Text, file)
		}()
	}
	wg.Wait()
}

func TestPlainInput(t *testing.T) {
	d := gndoc.New(tikaURL, gndoc.OptPreferTika(true))
	path := filepath.Join("testdata", "utf8.txt")
	res, err := d.TextFromFile(path, true)
	assert.Nil(t, err)
	data, err := os.ReadFile(path)
	assert.Nil(t, err)
	assert.Equal(t, string(data), res.Text)
	assert.Equal(t, gndoc.ConverterNone, res.Converter)
	assert.True(t, res.OrigOffsets)
	assert.True(t, res.OrigByteOffsets)
}

func TestDocNoTika(t *testing.T) {
	d := gndoc.New("")
	for _, file := range []string{"binary", "image.jpg"} {
		_, err := d.TextFromFile(filepath.Join("testdata", file), false)
		assert.ErrorIs(t, err, gndoc.ErrUnsupported, file)
	}
}

func TestDocTika(t *testing.T) {
	tests := []struct {
		msg, file, text, mime string
	}{
		{"binary", "binary", "", "application/x-executable"},
		{"jpg", "image.jpg", "Baccha occurs in Ontario", "image/jpeg"},
		{"pdf img", "image.pdf", "", "application/pdf"},
	}

	d := gndoc.New(tikaURL)
	for _, v := range tests {
		res, err := d.TextFromFile(filepath.Join("testdata", v.file), false)
		assert.Nil(t, err, v.msg)
		assert.Contains(t, res.Text, v.text, v.msg)
		assert.Equal(t, v.mime, res.MIME, v.msg)
		assert.Equal(t, gndoc.ConverterTika, res.Converter, v.msg)
		assert.False(t, res.OrigOffsets, v.msg)
	}
}

func TestDocTikaUnreachable(t *testing.T) {
	d := gndoc.New("http://127.0.0.1:1")

	// local conversion works, Tika is not called
	res, err := d.TextFromFile(filepath.Join("testdata", "file.pdf"), false)
	assert.Nil(t, err)
	assert.Contains(t, res.Text, "sabana de Bogotá")
	assert.Equal(t, gndoc.ConverterLocal, res.Converter)

	// format needs Tika, both errors are kept
	_, err = d.TextFromFile(filepath.Join("testdata", "image.jpg"), false)
	assert.ErrorIs(t, err, gndoc.ErrUnsupported)
	assert.Contains(t, err.Error(), "local: ")
	assert.Contains(t, err.Error(), "tika: ")

	// scanned PDF needs Tika
	_, err = d.TextFromFile(filepath.Join("testdata", "image.pdf"), false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no text layer")
}

func TestDocPreferTika(t *testing.T) {
	path := filepath.Join("testdata", "file.pdf")
	local := gndoc.New(tikaURL)
	resLocal, err := local.TextFromFile(path, false)
	assert.Nil(t, err)
	assert.Equal(t, gndoc.ConverterLocal, resLocal.Converter)

	tk := gndoc.New(tikaURL, gndoc.OptPreferTika(true))
	resTika, err := tk.TextFromFile(path, false)
	assert.Nil(t, err)
	assert.Contains(t, resTika.Text, "sabana de Bogotá")
	assert.Equal(t, gndoc.ConverterTika, resTika.Converter)
	assert.Equal(t, "application/pdf", resTika.MIME)
	// Tika keeps its own layout, so texts differ.
	assert.NotEqual(t, resLocal.Text, resTika.Text)

	// Tika might modify plain texts as well.
	res, err := tk.TextFromFile(filepath.Join("testdata", "utf8.txt"), false)
	assert.Nil(t, err)
	assert.Equal(t, gndoc.ConverterTika, res.Converter)
	assert.False(t, res.OrigOffsets)
	assert.False(t, res.OrigByteOffsets)

	// Tika is not reachable, local conversion is used instead.
	bad := gndoc.New("http://127.0.0.1:1", gndoc.OptPreferTika(true))
	res, err = bad.TextFromFile(path, false)
	assert.Nil(t, err)
	assert.Equal(t, resLocal.Text, res.Text)
	assert.Equal(t, gndoc.ConverterLocal, res.Converter)

	// Tika is not reachable, and format is not supported locally.
	_, err = bad.TextFromFile(filepath.Join("testdata", "image.jpg"), false)
	assert.ErrorIs(t, err, gndoc.ErrUnsupported)
}

func TestTextFromURL(t *testing.T) {
	d := gndoc.New(tikaURL)
	res, err := d.TextFromURL("https://example.org")
	assert.Nil(t, err)
	assert.Contains(t, res.Text, "Example")
	assert.NotContains(t, res.Text, "<html>")
	assert.Contains(t, res.MIME, "text/html")
	assert.False(t, res.OrigOffsets)
	assert.Greater(t, res.Duration, float32(0))
}

func TestTextFromURLWithUserAgent(t *testing.T) {
	ua := "test-agent/1.0 (https://example.org)"
	d := gndoc.New(tikaURL, gndoc.OptUserAgent(ua))
	res, err := d.TextFromURL("https://example.org")
	assert.Nil(t, err)
	assert.Contains(t, res.Text, "Example")
}

func TestDefaultPDFCacheDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir, err := gndoc.DefaultPDFCacheDir()
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(home, ".cache", "gndoc"), dir)
}

func TestConverterString(t *testing.T) {
	assert.Equal(t, "none", gndoc.ConverterNone.String())
	assert.Equal(t, "local", gndoc.ConverterLocal.String())
	assert.Equal(t, "tika", gndoc.ConverterTika.String())
}

func Example() {
	// Empty Tika URL means all conversions are done locally.
	gnd := gndoc.New("")
	res, err := gnd.TextFromFile("testdata/file.pdf", false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Contains(res.Text, "sabana de Bogotá"))
	fmt.Println(res.Converter, res.OrigOffsets)

	// The file is known to be UTF-8, conversion is not needed.
	res, err = gnd.TextFromFile("testdata/utf8.txt", true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Contains(res.Text, "Holarctic genus"))
	fmt.Println(res.Converter, res.OrigOffsets)

	// Only HTML pages are supported by TextFromURL.
	res, err = gnd.TextFromURL("https://example.org")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Contains(res.Text, "Example"))
	// Output:
	// true
	// local false
	// true
	// none true
	// true
}
