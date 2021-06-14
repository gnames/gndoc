package gndoc

import (
	"github.com/gnames/gndoc/ent/doc"
	"github.com/gnames/gnfinder/ent/output"
)

type GNdoc interface {
	Find(doc.Doc) output.Output
	GetConfig() Config
	ChangeConfig(opts ...Option) GNdoc
	GetVersion() FullVersion
}
