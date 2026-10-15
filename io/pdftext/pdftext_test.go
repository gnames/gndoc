package pdftext_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnames/gndoc/io/pdftext"
	"github.com/stretchr/testify/assert"
)

func TestPDFText(t *testing.T) {
	tests := []struct {
		msg, file, text string
	}{
		{"pdf", "file.pdf", "sabana de Bogotá"},
		{"pdf words", "file.pdf", "sect. Insignes is"},
		{"pdf no widths", "no_names.pdf", "Pull requests"},
		{"pdf no text", "image.pdf", ""},
	}

	pdf := pdftext.New()
	for _, v := range tests {
		data, err := os.ReadFile(filepath.Join("..", "..", "testdata", v.file))
		assert.Nil(t, err, v.msg)
		txt, err := pdf.PDFText(data)
		assert.Nil(t, err, v.msg)
		assert.Contains(t, txt, v.text, v.msg)
	}
}

func TestPDFLines(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "file.pdf"))
	assert.Nil(t, err)
	txt, err := pdftext.New().PDFText(data)
	assert.Nil(t, err)
	// hyphenated words keep their line breaks
	assert.Contains(t, txt, "new species be-\nlonging to Passiflora")
	assert.Contains(t, txt, "Poggen-\ndorffia, Bolivia")
	// PDFium hyphen markers do not leak into text
	assert.NotContains(t, txt, "\x02")
	assert.NotContains(t, txt, "￾")
}

func TestNotPDF(t *testing.T) {
	_, err := pdftext.New().PDFText([]byte("not a PDF"))
	assert.NotNil(t, err)
}
