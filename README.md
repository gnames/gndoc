# GNdoc

GNdoc is a library for extracting the content of a large variety of files
into UTF8-encoded text format.

## Install

```bash
go get github.com/gnames/gndoc
```

## Usage

```go
import (
  "fmt"
  "path/filepath"
  "strings"

  "github.com/gnames/gndoc"
)

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
  // Output:
  // true
  // true
}
```
