package doc

import "io"

type Doc interface {
	GetContent(io.Reader) (string, error)
	Content() string

	GetLanguage(io.Reader) (string, error)
	Language() string

	GetMeta(io.Reader) (string, error)
	Meta() string
}
