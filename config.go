package gndoc

import (
	"log"

	"github.com/gnames/gnfmt"
)

type Config struct {
	Format  gnfmt.Format
	TikaURL string
	Port    int
}

func NewConfig(opts ...Option) Config {
	cfg := Config{
		Format:  gnfmt.CSV,
		TikaURL: "https://tika.globalnames.org",
		Port:    8080,
	}

	for i := range opts {
		opts[i](&cfg)
	}
	return cfg
}

type Option func(*Config)

func OptFormat(s string) Option {
	return func(cfg *Config) {
		f, err := gnfmt.NewFormat(s)
		if err != nil {
			f = gnfmt.CSV
			log.Printf("Set default CSV format due to error: %s", err)
		}
		cfg.Format = f
	}
}

func OptTikaURL(s string) Option {
	return func(cfg *Config) {
		cfg.TikaURL = s
	}
}

func OptPort(i int) Option {
	return func(cfg *Config) {
		cfg.Port = i
	}
}
