package gnhttp_test

import (
	"testing"

	"github.com/gnames/gndoc/io/gnhttp"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	h := gnhttp.New()
	code, mime, body, err := h.Get("https://example.org")
	assert.Nil(t, err)
	_, _, _ = code, mime, body
}
