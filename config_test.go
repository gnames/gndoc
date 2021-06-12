package gndoc_test

import (
	"testing"

	"github.com/gnames/gndoc"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	t.Run("default opts", func(t *testing.T) {
		cfg := gndoc.NewConfig()
		assert.Equal(t, cfg.TikaURL, "https://tika.globalnames.org")
	})

	t.Run("change opts", func(t *testing.T) {
		opts := []gndoc.Option{
			gndoc.OptTikaURL("https://example.org"),
		}
		cfg := gndoc.NewConfig(opts...)
		assert.Equal(t, cfg.TikaURL, "https://example.org")
	})
}
