package convert

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// maxPartSize limits the size of an uncompressed part of an archive to
// protect from zip bombs.
const maxPartSize = 512 << 20

// docxText extracts text from MS Word documents (Office Open XML).
func docxText(data []byte) (string, error) {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	conf := xmlConf{
		text:   set("t"),
		breaks: set("p", "tr"),
		tabs:   set("tab", "tc"),
	}
	parts := []string{
		"word/document.xml",
		"word/footnotes.xml",
		"word/endnotes.xml",
	}
	return zipTexts(z, parts, conf)
}

// pptxText extracts text from MS PowerPoint presentations.
func pptxText(data []byte) (string, error) {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	conf := xmlConf{
		text:   set("t"),
		breaks: set("p"),
	}
	slides := numberedParts(z, "ppt/slides/slide")
	return zipTexts(z, slides, conf)
}

// odfText extracts text from OpenDocument text, spreadsheet and
// presentation files.
func odfText(data []byte) (string, error) {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	conf := xmlConf{
		text:   set("body"),
		breaks: set("p", "h", "table-row", "list-item"),
		tabs:   set("tab", "table-cell"),
		spaces: set("s"),
	}
	return zipTexts(z, []string{"content.xml"}, conf)
}

// zipTexts concatenates texts from the given parts of a zip archive.
// Missing parts are ignored.
func zipTexts(z *zip.Reader, parts []string, conf xmlConf) (string, error) {
	var res []string
	for _, p := range parts {
		f := zipFile(z, p)
		if f == nil {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return "", err
		}
		txt, err := xmlChars(io.LimitReader(r, maxPartSize), conf)
		r.Close()
		if err != nil {
			return "", fmt.Errorf("cannot parse %s: %w", p, err)
		}
		if txt != "" {
			res = append(res, txt)
		}
	}
	return strings.Join(res, "\n\n"), nil
}

func zipFile(z *zip.Reader, name string) *zip.File {
	for _, f := range z.File {
		if f.Name == name {
			return f
		}
	}
	return nil
}

var partNum = regexp.MustCompile(`(\d+)\.xml$`)

// numberedParts returns names of parts like 'prefix1.xml', 'prefix2.xml'
// sorted by their number.
func numberedParts(z *zip.Reader, prefix string) []string {
	type part struct {
		name string
		num  int
	}
	var parts []part
	for _, f := range z.File {
		if !strings.HasPrefix(f.Name, prefix) {
			continue
		}
		m := partNum.FindStringSubmatch(f.Name[len(prefix)-1:])
		if m == nil || path.Dir(f.Name) != path.Dir(prefix+"1.xml") {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		parts = append(parts, part{f.Name, n})
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].num < parts[j].num })
	res := make([]string, len(parts))
	for i := range parts {
		res[i] = parts[i].name
	}
	return res
}

// xlsxText extracts text from MS Excel spreadsheets. Cells are separated
// by tabs, rows by new lines.
func xlsxText(data []byte) (string, error) {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	shared, err := sharedStrings(z)
	if err != nil {
		return "", err
	}

	var res []string
	for _, name := range numberedParts(z, "xl/worksheets/sheet") {
		txt, err := sheetText(zipFile(z, name), shared)
		if err != nil {
			return "", fmt.Errorf("cannot parse %s: %w", name, err)
		}
		if txt != "" {
			res = append(res, txt)
		}
	}
	return strings.Join(res, "\n\n"), nil
}

// sharedStrings reads the table of strings that are referred to by cells.
func sharedStrings(z *zip.Reader) ([]string, error) {
	f := zipFile(z, "xl/sharedStrings.xml")
	if f == nil {
		return nil, nil
	}
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	dec := xml.NewDecoder(io.LimitReader(r, maxPartSize))
	var res []string
	var b strings.Builder
	var inT, inPhonetic bool
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "si":
				b.Reset()
			case "t":
				inT = true
			case "rPh":
				inPhonetic = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "si":
				res = append(res, b.String())
			case "t":
				inT = false
			case "rPh":
				inPhonetic = false
			}
		case xml.CharData:
			if inT && !inPhonetic {
				b.Write(t)
			}
		}
	}
	return res, nil
}

// sheetText converts a worksheet to tab-separated text.
func sheetText(f *zip.File, shared []string) (string, error) {
	r, err := f.Open()
	if err != nil {
		return "", err
	}
	defer r.Close()

	dec := xml.NewDecoder(io.LimitReader(r, maxPartSize))
	var rows []string
	var row []string
	var val strings.Builder
	var cellType string
	var inVal bool
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "row":
				row = row[:0]
			case "c":
				cellType = ""
				val.Reset()
				for _, a := range t.Attr {
					if a.Name.Local == "t" {
						cellType = a.Value
					}
				}
			case "v", "t":
				inVal = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "v", "t":
				inVal = false
			case "c":
				row = append(row, cellValue(cellType, val.String(), shared))
			case "row":
				line := strings.TrimRight(strings.Join(row, "\t"), "\t")
				if line != "" {
					rows = append(rows, line)
				}
			}
		case xml.CharData:
			if inVal {
				val.Write(t)
			}
		}
	}
	return strings.Join(rows, "\n"), nil
}

func cellValue(cellType, val string, shared []string) string {
	if cellType == "s" {
		idx, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil || idx < 0 || idx >= len(shared) {
			return ""
		}
		val = shared[idx]
	}
	// cells with new lines or tabs would break the table structure
	return strings.Join(strings.Fields(val), " ")
}
