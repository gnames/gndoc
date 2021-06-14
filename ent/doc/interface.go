package doc

import "io"

type Doc interface {
	ContentFromFile(path string) (string, float32, error)

	GetContent(io.Reader) (string, error)
	Content() string

	GetLanguage(io.Reader) (string, error)
	Language() string

	GetMeta(io.Reader) (string, error)
	Meta() string
}
