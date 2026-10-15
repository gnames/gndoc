# Changelog

## Unreleased

- Add: convert plain texts, HTML, XML, PDF, DOCX, XLSX, PPTX and
  OpenDocument files locally using Go modules. PDFs are converted by
  PDFium (WebAssembly build via go-pdfium).
- Add: Tika service is optional, it is used only as a fallback for formats
  that cannot be converted locally (images, scanned PDFs etc.). Use an empty
  Tika URL to disable it.
- Add: `OptPreferTika` option to use Tika for all formats, with local
  conversion as a fallback.
- Add: `ErrUnsupported` error for formats that need Tika when it is disabled.
- Add: BREAKING `TextFromFile`, `TextFromURL` and `GetText` return
  `Result` with the text, detected MIME type, converter, conversion time,
  and `OrigOffsets`, `OrigByteOffsets` flags that tell if character or byte
  offsets in the text match the original input.
- Add: plain texts keep BOM as U+FEFF, so character offsets match the input.
- Add: UTF-8 texts with some invalid bytes are not decoded as Windows-1252,
  invalid bytes are replaced with U+FFFD instead.
- Add: `OptPDFCacheDir` option to cache compiled PDFium module on disk
  for faster start of PDF conversion, and `DefaultPDFCacheDir` that returns
  `~/.cache/gndoc`.
- Add: BREAKING GNdoc is safe for concurrent use, `Text()` method is removed,
  use `Result.Text` instead.
- Add: minimal Go version is 1.26.

## [v0.3.0]

- Add [#9]: add a method to read texts from URLs.

## [v0.2.1]

- Add: Increase wait time for tika to 15 sec.

## [v0.2.0]

- Add [#8]: add a flag for UTF8 texts to treat them locally without Tika.

## [v0.1.0]

- Add [#7]: convert GNdoc into library only for extracting text from files.
- Add [#3]: find names in files.
- Add [#2]: add config and CLI app.
- Add [#1]: extract text from files.

## Footnotes

This document follows [changelog guidelines]

[v0.3.0]: https://github.com/gnames/gndoc/compare/v0.2.1...v0.3.0
[v0.2.1]: https://github.com/gnames/gndoc/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/gnames/gndoc/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/gnames/gndoc/tree/v0.1.0

[#10]: https://github.com/gnames/gndoc/issues/10
[#9]: https://github.com/gnames/gndoc/issues/9
[#8]: https://github.com/gnames/gndoc/issues/8
[#7]: https://github.com/gnames/gndoc/issues/7
[#6]: https://github.com/gnames/gndoc/issues/6
[#5]: https://github.com/gnames/gndoc/issues/5
[#4]: https://github.com/gnames/gndoc/issues/4
[#3]: https://github.com/gnames/gndoc/issues/3
[#2]: https://github.com/gnames/gndoc/issues/2
[#1]: https://github.com/gnames/gndoc/issues/1

[changelog guidelines]: https://github.com/olivierlacan/keep-a-changelog
