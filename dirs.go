package gndoc

import (
	"path/filepath"

	"github.com/gnames/gndoc/io/fsio"
)

// DefaultPDFCacheDir returns ~/.cache/gndoc, a directory for
// OptPDFCacheDir. Like git, GNdoc uses ~/.cache on all operating
// systems.
func DefaultPDFCacheDir() (string, error) {
	home, err := fsio.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "gndoc"), nil
}
