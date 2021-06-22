package gndoc_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnames/gndoc"
	"github.com/stretchr/testify/assert"
)

var tikaURL = "https://tika.globalnames.org"

func TestDoc(t *testing.T) {
	tests := []struct {
		msg, file, text, meta, lang string
	}{
		{"abbr", "abbreviations.txt", "sister species", "text/plain", "en"},
		{"ascii", "ascii.txt", "no unicode", "text/plain", "en"},
		{"big", "big.txt", "Tacsonia insignis, 38", "text/plain", "en"},
		{"binary", "binary", "", "application/x-executable", "en"},
		{"json", "dirty_names.json", "Rhus folirsternatis", "text/plain", "en"},
		// why language is detected as Thai?
		{"ms doc", "file.docx", "Global Names", "Microsoft Office Word", "th"},
		{"html", "file.html", "Aesculus", "text/html", "en"},
		// why language is detected as Thai?
		{"pdf", "file.pdf", "sabana de Bogotá", "application/pdf", "th"},
		// why language is detected as Thai?
		{"xlsx", "file.xlsx", "Uhler, 1872", "LibreOffice_project", "th"},
		{"xml", "file.xml", "fishes of New Zealand", "application/xml", "en"},
		{"jpg", "image.jpg", "Baccha occurs in Ontario", "image/jpeg", "th"},
		{"pdf img", "image.pdf", "", "application/pdf", "th"},
		{"italian", "italian.txt", "per il foro della", "text/plain", "it"},
		{"latin1", "latin1.txt", "(Ujvárosi 2005)", "charset=ISO-8859-1", "en"},
		{"utf16", "utf16.txt", "avons également reçu", "text/plain", "fr"},
	}

	for _, v := range tests {
		d := gndoc.New(tikaURL)

		f, err := os.Open(filepath.Join("testdata", v.file))
		assert.Nil(t, err)
		assert.Nil(t, err)
		txt, err := d.GetText(f)
		assert.Nil(t, err)
		assert.Contains(t, txt, v.text, v.msg)
		assert.Equal(t, d.Text(), txt)
		f.Close()
	}
}

func TestTextFromURL(t *testing.T) {
	d := gndoc.New(tikaURL)
	txt, _, err := d.TextFromURL("https://example.org")
	assert.Nil(t, err)
	assert.Contains(t, txt, "Example")
	assert.NotContains(t, txt, "<html>")
}

func Example() {
	gnd := gndoc.New(tikaURL)
	path := filepath.Join("testdata/file.pdf")
	txt, _, err := gnd.TextFromFile(path, false)
	if err != nil {
		log.Fatal(err)
	}
	hasText := strings.Contains(txt, "sabana de Bogotá")
	fmt.Printf("%v\n", hasText)

	path = filepath.Join("testdata/utf8.txt")
	txt, _, err = gnd.TextFromFile(path, true)
	if err != nil {
		log.Fatal(err)
	}
	hasText = strings.Contains(txt, "Holarctic genus")
	fmt.Printf("%v\n", hasText)

	url := "https://example.org"
	txt, _, err = gnd.TextFromURL(url)
	if err != nil {
		log.Fatal(err)
	}
	hasText = strings.Contains(txt, "Example")
	fmt.Printf("%v\n", hasText)
	// Output:
	// true
	// true
	// true
}
