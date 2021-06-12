package gndoc

import (
	"fmt"
	"os"

	"github.com/gnames/gndoc/ent/doc"
	"github.com/gnames/gnfinder"
	gnfc "github.com/gnames/gnfinder/config"
	"github.com/gnames/gnfinder/ent/nlp"
	"github.com/gnames/gnfinder/ent/output"
	"github.com/gnames/gnfinder/io/dict"
	"github.com/gnames/gnsys"
)

type gndoc struct {
	Config
	gnf gnfinder.GNfinder
}

func New(cfg Config) GNdoc {
	dict := dict.LoadDictionary()
	weights := nlp.BayesWeights()
	gnfConfig := gnfc.New()
	gnd := gndoc{
		Config: cfg,
		gnf:    gnfinder.New(gnfConfig, dict, weights),
	}
	return gnd
}

func (gnd gndoc) FileToText(path string) (doc.Doc, error) {
	exists, err := gnsys.FileExists(path)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("File '%s' does not exist", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	d := doc.NewDoc(gnd.GetConfig().TikaURL)
	_, err = d.GetContent(f)
	return d, err
}

func (gnd gndoc) Find(d doc.Doc) output.Output {
	return gnd.gnf.Find(d.Content())
}

func (gnd gndoc) GetConfig() Config {
	return gnd.Config
}

func (gnd gndoc) ChangeConfig(opts ...Option) GNdoc {
	for _, opt := range opts {
		opt(&gnd.Config)
	}
	return gnd
}

func (gnd gndoc) GetVersion() FullVersion {
	gnfver := gnd.gnf.GetVersion()
	res := FullVersion{
		Version:         Version,
		Build:           Build,
		GNfinderVersion: gnfver.Version,
	}
	return res
}
