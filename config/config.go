package config

import (
	"flag"
)

type Config struct {
	ServerAddr string
	BaseURL    string
}

func Init() (*Config, error) {
	cfg := Config{}

	// init flags
	flag.StringVar(&cfg.ServerAddr, "a", ":8080", "server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "base url")

	flag.Parse()

	return &cfg, nil
}
