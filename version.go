package gndoc

var (
	// Version of the gnverifier
	Version = "v0.1+"
	// Build timestamp
	Build = "n/a"
)

type FullVersion struct {
	Version         string
	Build           string
	GNfinderVersion string
}
