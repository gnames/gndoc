package gnhttp_test

import (
	"io"
	"testing"
	"time"

	"github.com/gnames/gndoc/io/gnhttp"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	h := gnhttp.New()
	code, mime, body, err := h.Get("https://example.org")
	assert.Nil(t, err)
	assert.Equal(t, code, 200)
	assert.Contains(t, mime, "text/html")
	bs, err := io.ReadAll(body)
	assert.Nil(t, err)
	assert.Contains(t, string(bs), "Example")
}

func TestOpts(t *testing.T) {
	h := gnhttp.New()
	assert.Equal(t, h.ConnMax(), 200*time.Millisecond)
	assert.Equal(t, h.ReqMax(), 1000*time.Millisecond)
	opts := []gnhttp.Option{
		gnhttp.OptConnMax(3 * time.Second),
		gnhttp.OptReqMax(2 * time.Second),
	}
	h = gnhttp.New(opts...)
	assert.Equal(t, h.ConnMax(), 3000*time.Millisecond)
	assert.Equal(t, h.ReqMax(), 2000*time.Millisecond)
}
