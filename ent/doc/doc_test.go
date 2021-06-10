package doc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnames/gndoc"
	"github.com/gnames/gndoc/ent/doc"
	"github.com/stretchr/testify/assert"
)

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

	cnf := gndoc.NewConfig()

	for _, v := range tests {
		d := doc.NewDoc(cnf.TikaURL)

		f, err := os.Open(filepath.Join("..", "..", "testdata", v.file))
		assert.Nil(t, err)
		lang, err := d.GetLanguage(f)
		assert.Nil(t, err)
		assert.Equal(t, lang, v.lang, v.msg)
		assert.Equal(t, d.Language(), lang)
		f.Close()

		f, err = os.Open(filepath.Join("..", "..", "testdata", v.file))
		assert.Nil(t, err)
		assert.Nil(t, err)
		meta, err := d.GetMeta(f)
		assert.Nil(t, err)
		assert.Contains(t, meta, v.meta, v.msg)
		assert.Equal(t, d.Meta(), meta)
		f.Close()

		f, err = os.Open(filepath.Join("..", "..", "testdata", v.file))
		assert.Nil(t, err)
		assert.Nil(t, err)
		txt, err := d.GetContent(f)
		assert.Nil(t, err)
		// if v.msg == "pdf img" {
		// 	fmt.Printf("txt: %s\n", txt)
		// }
		assert.Contains(t, txt, v.text, v.msg)
		assert.Equal(t, d.Content(), txt)
		f.Close()

	}
}
