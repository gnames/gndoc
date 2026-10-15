package convert

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var htmlSpace = regexp.MustCompile(`\s+`)

// htmlBlocks are HTML elements that start a new line of text.
var htmlBlocks = map[string]struct{}{
	"address": {}, "article": {}, "aside": {}, "blockquote": {}, "br": {},
	"dd": {}, "div": {}, "dl": {}, "dt": {}, "figcaption": {}, "figure": {},
	"footer": {}, "form": {}, "h1": {}, "h2": {}, "h3": {}, "h4": {},
	"h5": {}, "h6": {}, "header": {}, "hr": {}, "li": {}, "main": {},
	"nav": {}, "ol": {}, "p": {}, "pre": {}, "section": {}, "table": {},
	"td": {}, "th": {}, "tr": {}, "ul": {}, "title": {},
}

// htmlSkip are HTML elements that do not contain readable text.
var htmlSkip = map[string]struct{}{
	"script": {}, "style": {}, "noscript": {}, "template": {}, "svg": {},
	"head": {},
}

// htmlText extracts readable text from an HTML document.
func htmlText(data []byte) (string, error) {
	r, err := charset.NewReader(bytes.NewReader(data), "text/html")
	if err != nil {
		return "", err
	}
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	// title is inside of head, so we take it separately
	if t := findNode(doc, "title"); t != nil {
		walkHTML(t, &b)
		b.WriteString("\n\n")
	}
	walkHTML(doc, &b)
	return normalizeSpace(b.String()), nil
}

func findNode(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if res := findNode(c, tag); res != nil {
			return res
		}
	}
	return nil
}

func inPre(n *html.Node) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && p.Data == "pre" {
			return true
		}
	}
	return false
}

func walkHTML(n *html.Node, b *strings.Builder) {
	switch n.Type {
	case html.TextNode:
		if inPre(n) {
			b.WriteString(n.Data)
			return
		}
		b.WriteString(htmlSpace.ReplaceAllString(n.Data, " "))
		return
	case html.ElementNode:
		if _, ok := htmlSkip[n.Data]; ok {
			return
		}
	}

	_, block := htmlBlocks[n.Data]
	if block && n.Type == html.ElementNode {
		b.WriteByte('\n')
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkHTML(c, b)
	}
	if block && n.Type == html.ElementNode {
		b.WriteByte('\n')
	}
}

// xmlText extracts character data from a generic XML document.
func xmlText(data []byte) (string, error) {
	return xmlChars(bytes.NewReader(data), xmlConf{})
}

// xmlConf determines how text is collected from an XML document.
type xmlConf struct {
	// text contains elements that keep text. If it is empty, text is
	// collected from all elements.
	text map[string]struct{}
	// breaks contains elements that end a line of text. If it is empty,
	// all elements end a line.
	breaks map[string]struct{}
	// tabs contains elements that are converted to a tab character.
	tabs map[string]struct{}
	// spaces contains elements that are converted to a space character.
	spaces map[string]struct{}
}

// xmlChars collects character data from XML according to the
// configuration.
func xmlChars(r io.Reader, conf xmlConf) (string, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	dec.CharsetReader = charset.NewReaderLabel

	var b strings.Builder
	var inText int
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.CharData:
			if conf.text == nil || inText > 0 {
				b.Write(t)
			}
		case xml.StartElement:
			name := t.Name.Local
			if _, ok := conf.text[name]; ok {
				inText++
			}
			if _, ok := conf.tabs[name]; ok {
				b.WriteByte('\t')
			}
			if _, ok := conf.spaces[name]; ok {
				b.WriteByte(' ')
			}
			if name == "br" || name == "line-break" {
				b.WriteByte('\n')
			}
		case xml.EndElement:
			name := t.Name.Local
			if _, ok := conf.text[name]; ok {
				inText--
			}
			if conf.breaks == nil {
				b.WriteByte('\n')
			} else if _, ok := conf.breaks[name]; ok {
				b.WriteByte('\n')
			}
		}
	}
	return normalizeSpace(b.String()), nil
}

// set creates a set out of strings.
func set(ss ...string) map[string]struct{} {
	res := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		res[s] = struct{}{}
	}
	return res
}
