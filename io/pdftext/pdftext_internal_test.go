package pdftext

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuntimeConfig(t *testing.T) {
	assert.NotNil(t, runtimeConfig(""))
	assert.NotNil(t, runtimeConfig(t.TempDir()))

	// a file instead of a directory, the cache cannot be used
	file := filepath.Join(t.TempDir(), "file")
	assert.Nil(t, os.WriteFile(file, nil, 0644))
	assert.NotNil(t, runtimeConfig(filepath.Join(file, "cache")))
}

func TestSetPDFCacheDirAfterStart(t *testing.T) {
	_, err := getPool()
	assert.Nil(t, err)
	before := pdfCacheDir
	SetCacheDir(t.TempDir())
	assert.Equal(t, before, pdfCacheDir)
}
