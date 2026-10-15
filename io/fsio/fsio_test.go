package fsio_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnames/gndoc/io/fsio"
	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	assert.Nil(t, os.WriteFile(path, []byte("Pardosa moesta"), 0644))
	data, err := fsio.ReadFile(path)
	assert.Nil(t, err)
	assert.Equal(t, "Pardosa moesta", string(data))

	_, err = fsio.ReadFile(filepath.Join(t.TempDir(), "missing.txt"))
	assert.ErrorContains(t, err, "does not exist")
}

func TestHomeDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir, err := fsio.HomeDir()
	assert.Nil(t, err)
	assert.Equal(t, home, dir)
}
