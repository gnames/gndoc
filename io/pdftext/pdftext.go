// Package pdftext extracts the text layer of PDF documents using PDFium
// compiled to WebAssembly. PDFium workers are shared by the whole
// program, they start on the first conversion.
package pdftext

import (
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

// hyphenBreaks restores line breaks after hyphens. PDFium replaces a
// hyphen at the end of a line, together with the line break, by a marker
// character (0x02, or U+FFFE in some APIs). Keeping lines as they are in
// the document lets clients decide how to join hyphenated words.
var hyphenBreaks = strings.NewReplacer("\x02", "-\n", "\ufffe", "-\n")

// pdfTimeout limits waiting for a free PDFium worker.
const pdfTimeout = 30 * time.Second

var (
	pdfMu       sync.Mutex
	pdfPool     pdfium.Pool
	pdfPoolErr  error
	pdfStarted  bool
	pdfCacheDir string
)

// PDFText implements convert.PDF interface.
type PDFText struct{}

// New creates a PDF text extractor.
func New() PDFText {
	return PDFText{}
}

// PDFText extracts text layer from a PDF document. Scanned documents
// without text layer return an empty string.
func (PDFText) PDFText(data []byte) (string, error) {
	return pdfText(data)
}

// SetCacheDir sets a directory for caching the compiled PDFium
// WebAssembly module, so following runs of a program do not need to
// compile it again. It has an effect only before the first PDF
// conversion, because PDFium workers are shared by the whole program.
func SetCacheDir(dir string) {
	pdfMu.Lock()
	defer pdfMu.Unlock()
	if !pdfStarted {
		pdfCacheDir = dir
	}
}

// getPool starts PDFium WebAssembly workers on the first use, so
// programs that do not convert PDFs do not pay for it.
func getPool() (pdfium.Pool, error) {
	pdfMu.Lock()
	defer pdfMu.Unlock()
	if pdfStarted {
		return pdfPool, pdfPoolErr
	}
	pdfStarted = true
	pdfPool, pdfPoolErr = webassembly.Init(webassembly.Config{
		MinIdle:       0,
		MaxIdle:       1,
		MaxTotal:      runtime.NumCPU(),
		RuntimeConfig: runtimeConfig(pdfCacheDir),
		// documents are loaded from memory, PDFium does not need
		// access to the file system.
		FSConfig: wazero.NewFSConfig(),
	})
	return pdfPool, pdfPoolErr
}

// runtimeConfig creates WebAssembly runtime settings with an optional
// compilation cache. If the cache cannot be created, it is not used.
func runtimeConfig(cacheDir string) wazero.RuntimeConfig {
	// go-pdfium requires these features for the bundled PDFium module.
	cfg := wazero.NewRuntimeConfig().WithCoreFeatures(
		api.CoreFeaturesV2 | experimental.CoreFeaturesExceptionHandling,
	)
	if cacheDir == "" {
		return cfg
	}
	cache, err := wazero.NewCompilationCacheWithDir(cacheDir)
	if err != nil {
		return cfg
	}
	return cfg.WithCompilationCache(cache)
}

func pdfText(data []byte) (string, error) {
	pool, err := getPool()
	if err != nil {
		return "", err
	}
	inst, err := pool.GetInstance(pdfTimeout)
	if err != nil {
		return "", err
	}
	defer inst.Close()

	doc, err := inst.OpenDocument(&requests.OpenDocument{File: &data})
	if err != nil {
		return "", err
	}
	defer inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{
		Document: doc.Document,
	})

	count, err := inst.FPDF_GetPageCount(&requests.FPDF_GetPageCount{
		Document: doc.Document,
	})
	if err != nil {
		return "", err
	}

	var b strings.Builder
	for i := 0; i < count.PageCount; i++ {
		res, err := inst.GetPageText(&requests.GetPageText{
			Page: requests.Page{
				ByIndex: &requests.PageByIndex{
					Document: doc.Document,
					Index:    i,
				},
			},
		})
		if err != nil {
			return "", err
		}
		b.WriteString(hyphenBreaks.Replace(res.Text))
		b.WriteString("\n\n")
	}
	return b.String(), nil
}
