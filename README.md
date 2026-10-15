# GNdoc

GNdoc is a library for extracting the content of a large variety of files
into UTF8-encoded text format.

## Supported formats

The following formats are converted locally, without external services:

- plain texts (including JSON, CSV etc.) in UTF-8, UTF-16 (with or without
  BOM) and legacy 8-bit encodings (Latin-1/Windows-1252)
- HTML and XML
- PDF documents with a text layer, using [PDFium] compiled to WebAssembly
  via [go-pdfium], so no CGO or system libraries are needed. See
  [PDF performance](#pdf-performance) about the start-up time.
- MS Office Open XML (DOCX, XLSX, PPTX)
- OpenDocument (ODT, ODS, ODP)

Format is detected from the content of a file, not from its extension.

Other formats (images, scanned PDFs that require OCR, legacy MS Office
files etc.) need an [Apache Tika] service.

### PDF text layout

Text keeps the lines of a PDF document. Words hyphenated at the end of a
line are not joined, they stay as `hyphen-\nated`, so clients can decide
how to join them.

## Tika service

The first argument of `gndoc.New` is a URL of a Tika service. Tika is
optional, there are three modes of using it:

| Setup                                           | Behavior                          |
|-------------------------------------------------|-----------------------------------|
| `gndoc.New(tikaURL)`                            | local first, Tika as a fallback   |
| `gndoc.New(tikaURL, gndoc.OptPreferTika(true))` | Tika first, local as a fallback   |
| `gndoc.New("")`                                 | local conversion only, no network |

The choice is made once, when GNdoc is created. Clients do not need to
decide which files go to Tika, GNdoc decides it for every file.

### Tika as a fallback (recommended)

Keep a Tika URL in the client's settings and always give it to
`gndoc.New`. Files are converted locally, and Tika is called only when
local conversion cannot deal with a file. Having the URL costs nothing:
GNdoc does not connect to Tika until a file needs it, so most files never
leave the machine.

```go
gnd := gndoc.New("https://tika.globalnames.org")
```

A file is sent to Tika if local conversion:

- does not support the format (images, legacy `.doc`, `.xls`, RTF etc.);
- fails, for example on a corrupted or a password-protected file;
- returns no text for a PDF file (a scanned PDF without a text layer).

Note that a PDF where only some pages are scanned has some text, so it is
not sent to Tika.

If Tika fails as well (for example, the service is not reachable), the
error contains both the local and the Tika reasons, and
`errors.Is(err, gndoc.ErrUnsupported)` still reports formats that need
Tika.

### Tika first

For users who prefer Tika's output for all files, there is the
`OptPreferTika(true)` option. It sends every file to Tika. Note that it
is "Tika first", not "Tika only": if Tika fails (for example, the service
is not reachable), the file is converted locally. If both fail, the error
contains both reasons. The option has no effect if Tika URL is empty.

```go
gnd := gndoc.New("https://tika.globalnames.org", gndoc.OptPreferTika(true))
```

### Local only

An empty Tika URL disables Tika. Use it when files must not be sent to a
remote service, or there is no network. Formats that cannot be converted
locally return an error that wraps `gndoc.ErrUnsupported`:

```go
gnd := gndoc.New("")
_, err := gnd.TextFromFile("image.jpg", false)
if errors.Is(err, gndoc.ErrUnsupported) {
  // needs Tika
}
```

### Plain UTF-8 input

`TextFromFile(path, true)` treats the file as a UTF-8 text and returns it
as is, without format detection, local conversion or Tika.

## Result and offsets

`TextFromFile`, `TextFromURL` and `GetText` return a `gndoc.Result`:

```go
type Result struct {
  Text            string    // UTF8-encoded text
  MIME            string    // detected format, e.g. "application/pdf"
  Converter       Converter // ConverterNone, ConverterLocal or ConverterTika
  OrigOffsets     bool      // character offsets in Text match the input
  OrigByteOffsets bool      // byte offsets in Text match the input
  Duration        float32   // conversion time in seconds
}
```

`Text` is always UTF-8, inputs in other encodings are converted.

Clients like [GNfinder] find strings in `Text` and report their offsets,
counted either in characters or in bytes. `OrigOffsets` and
`OrigByteOffsets` tell if such offsets can be used with the original file.
Check the flag that corresponds to the kind of offsets a client reports.

- **Plain UTF-8 text**: returned byte-to-byte as in the file, including a
  BOM, white spaces and line endings. GNdoc does not apply Unicode
  normalization (NFC/NFD). Both character and byte offsets match the
  original file.
- **Plain text in UTF-16 or Latin-1/Windows-1252**: re-encoded into UTF-8.
  Every character of the input stays one character in `Text` (a BOM is
  kept as U+FEFF), so **character** offsets match the original file, but
  byte offsets do not, because characters take a different number of
  bytes in UTF-8.
- **Damaged UTF-8 text** (a UTF-8 file with some invalid bytes, for
  example after a broken download): every invalid byte is replaced with
  one U+FFFD character, the rest stays as is. Character offsets match the
  original file, byte offsets do not. A text is treated as damaged UTF-8
  if it has a UTF-8 BOM, or if its valid UTF-8 multi-byte characters
  outnumber invalid bytes. Otherwise it is decoded as Windows-1252.
- **Converted formats** (PDF, HTML, XML, MS Office, OpenDocument, or
  anything converted by Tika, including plain texts): neither kind of
  offsets matches the original file. White spaces of locally converted
  formats are normalized: lines are trimmed, runs of spaces and empty lines
  are collapsed. Offsets are valid only for `Text`, so clients should give
  `Text` to users together with the offsets, or warn that offsets refer to
  the converted text.

| Input                            | Converter | OrigOffsets | OrigByteOffsets |
|----------------------------------|-----------|-------------|-----------------|
| `TextFromFile(path, true)`       | none      | true        | true            |
| UTF-8 plain text                 | local     | true        | true            |
| UTF-16, Latin-1 plain text       | local     | true        | false           |
| damaged UTF-8 plain text         | local     | true        | false           |
| other formats, converted locally | local     | false       | false           |
| any format, converted by Tika    | tika      | false       | false           |

Note that a plain text file that looks like HTML or XML is detected as
markup and converted. If a file is known to be a UTF-8 text, use
`TextFromFile(path, true)`.

## PDF performance

PDFium WebAssembly module is compiled on the first PDF conversion in a
program, it takes about 1.5 seconds. Following conversions are fast
(about 50 ms for a 20-page article). The PDFium workers are shared by all
GNdoc instances of a program, there are up to one worker per CPU.

### Compilation cache

Programs that run once per file (for example, a command line tool) pay
the start-up time on every run. `OptPDFCacheDir` keeps the compiled
module in a directory, so only the first run compiles it:

```go
cacheDir, err := gndoc.DefaultPDFCacheDir() // ~/.cache/gndoc
if err != nil {
  log.Fatal(err)
}
gnd := gndoc.New(tikaURL, gndoc.OptPDFCacheDir(cacheDir))
```

`DefaultPDFCacheDir` returns `~/.cache/gndoc` on all operating systems,
following the same convention as git. Any other directory can be used as
well. The cache is not used unless the option is given. Missing
directories are created (with `0700` permissions).

On our test PDF, a run with a filled cache took 0.09 seconds instead of
1.65. The cache takes about 20 MB of disk space. The setting is
process-wide, and it has an effect only if it is given before the first
PDF conversion. If the cache cannot be used, PDFs are converted without
it.

The cache directory can be deleted at any time, it will be created again
on the next run. A new version of PDFium or of the WebAssembly runtime
adds new files to the cache without removing old ones, so after upgrades
deleting the directory frees disk space.

### Many files

To convert many files faster, convert them in one program run, and use
several goroutines. GNdoc is safe for concurrent use, one instance can
be shared by all goroutines:

```go
gnd := gndoc.New(tikaURL)
var wg sync.WaitGroup
for _, path := range paths {
  wg.Add(1)
  go func() {
    defer wg.Done()
    res, err := gnd.TextFromFile(path, false)
    // ...
  }()
}
wg.Wait()
```

In real code limit the number of goroutines (for example, to the number
of CPUs), because every PDF conversion holds a PDFium worker. Other PDF
conversions wait for a free worker, and fail if they wait longer than 30
seconds.

## Install

Requires Go 1.26 or later.

```bash
go get github.com/gnames/gndoc
```

## Usage

```go
import (
  "fmt"
  "log"
  "strings"

  "github.com/gnames/gndoc"
)

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
```

`GetText(io.Reader)` converts any input, for example an opened file or an
HTTP response body. The whole input is read into memory.

`OptUserAgent` sets the User-Agent header for `TextFromURL` requests:

```go
gnd := gndoc.New("", gndoc.OptUserAgent("my-app/1.0 (https://example.org)"))
```

## Development

Code that has side effects (file system, network, PDFium workers) lives
in `io`, the rest of the code is pure:

| Package       | Purpose                                                   |
|---------------|-----------------------------------------------------------|
| `gndoc`       | API, decides which converter to use                       |
| `ent/convert` | format detection, conversion of texts, markup and Office  |
| `io/pdftext`  | PDF text extraction with PDFium WebAssembly workers       |
| `io/tikaio`   | client for Apache Tika service                            |
| `io/gnhttp`   | download of web pages                                     |
| `io/fsio`     | reading files, user's home directory                      |

`ent/convert` does not depend on PDFium, it receives a PDF converter
through the `convert.PDF` interface.

Run tests with [just]:

```bash
just test
```

Some tests need network access to `https://tika.globalnames.org` and
`https://example.org`.

[Apache Tika]: https://tika.apache.org/
[GNfinder]: https://github.com/gnames/gnfinder
[PDFium]: https://pdfium.googlesource.com/pdfium/
[go-pdfium]: https://github.com/klippa-app/go-pdfium
[just]: https://github.com/casey/just
