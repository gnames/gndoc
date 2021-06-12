package gndoc_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/gnames/gndoc"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		cfg := gndoc.NewConfig()
		gnd := gndoc.New(cfg)
		assert.Equal(t, gnd.GetConfig().TikaURL, "https://tika.globalnames.org")
	})

	t.Run("options", func(t *testing.T) {
		opts := []gndoc.Option{
			gndoc.OptTikaURL("https://example.org"),
		}
		cfg := gndoc.NewConfig(opts...)
		gnd := gndoc.New(cfg)
		assert.Equal(t, gnd.GetConfig().TikaURL, "https://example.org")
	})
}

func TestGetVersion(t *testing.T) {
	cfg := gndoc.NewConfig()
	gnd := gndoc.New(cfg)
	ver := gnd.GetVersion()
	assert.Regexp(t, regexp.MustCompile(`^v`), ver.Version)
	assert.Regexp(t, regexp.MustCompile(`^v`), ver.GNfinderVersion)
	assert.True(t, len(ver.Build) > 0)
}

func TestChangeConfig(t *testing.T) {
	cfg := gndoc.NewConfig()
	gnd := gndoc.New(cfg)
	assert.Equal(t, gnd.GetConfig().TikaURL, "https://tika.globalnames.org")
	opts := []gndoc.Option{
		gndoc.OptTikaURL("https://example.org"),
	}
	gnd2 := gnd.ChangeConfig(opts...)
	assert.Equal(t, gnd.GetConfig().TikaURL, "https://tika.globalnames.org")
	assert.Equal(t, gnd2.GetConfig().TikaURL, "https://example.org")
}

func TestFileToText(t *testing.T) {
	tests := []struct {
		msg, file, text string
		hasError        bool
	}{
		{"bad", "nofile.txt", "", true},
		{"txt", "utf8.txt", "Holarctic genus", false},
	}

	cfg := gndoc.NewConfig()
	gnd := gndoc.New(cfg)
	for _, v := range tests {
		path := filepath.Join("testdata", v.file)
		doc, err := gnd.FileToText(path)
		assert.Equal(t, err != nil, v.hasError)
		if !v.hasError {
			assert.Contains(t, doc.Content(), v.text)
		}
	}
}

func TestFind(t *testing.T) {
	cfg := gndoc.NewConfig()
	gnd := gndoc.New(cfg)
	path := filepath.Join("testdata", "utf8.txt")
	doc, err := gnd.FileToText(path)
	assert.Nil(t, err)

	t.Run("default find", func(t *testing.T) {
		o := gnd.Find(doc)
		assert.Greater(t, len(o.Names), 0)
	})
}
