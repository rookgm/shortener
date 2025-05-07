package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr string
	BaseURL    string
}

func Init() (*Config, error) {
	cfg := Config{}

	serverAddrEnv := os.Getenv("SERVER_ADDRESS")
	baseURLEnv := os.Getenv("BASE_URL")

	// init flags
	flag.StringVar(&cfg.ServerAddr, "a", ":8080", "server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "base url")

	flag.Parse()

	if serverAddrEnv != "" {
		cfg.ServerAddr = serverAddrEnv
	}

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = ":8080"
	}

	if baseURLEnv != "" {
		cfg.BaseURL = baseURLEnv
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080/"
	}

	return &cfg, nil
}
