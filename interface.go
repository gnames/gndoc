package gndoc

import (
	"github.com/gnames/gndoc/ent/doc"
	"github.com/gnames/gnfinder/ent/output"
)

type GNdoc interface {
	FileToText(string) (doc.Doc, error)
	Find(doc.Doc) output.Output
	GetConfig() Config
	ChangeConfig(opts ...Option) GNdoc
	GetVersion() FullVersion
}
