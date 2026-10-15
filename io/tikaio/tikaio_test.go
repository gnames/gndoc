package tikaio_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnames/gndoc/io/tikaio"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "image.jpg"))
	assert.Nil(t, err)
	tk := tikaio.New("https://tika.globalnames.org", 15*time.Second)
	txt, err := tk.Parse(data)
	assert.Nil(t, err)
	assert.Contains(t, txt, "Baccha occurs in Ontario")
}

func TestUnreachable(t *testing.T) {
	tk := tikaio.New("http://127.0.0.1:1", time.Second)
	_, err := tk.Parse([]byte("text"))
	assert.NotNil(t, err)
}
