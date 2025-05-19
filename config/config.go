package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr  string
	BaseURL     string
	LogLevel    string
	StoragePath string
}

func Init() (*Config, error) {
	cfg := Config{}

	// init flags
	flag.StringVar(&cfg.ServerAddr, "a", ":8080", "server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "base url")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.StoragePath, "f", "storage.json", "storage path")

	flag.Parse()

	if serverAddrEnv := os.Getenv("SERVER_ADDRESS"); serverAddrEnv != "" {
		cfg.ServerAddr = serverAddrEnv
	}

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = ":8080"
	}

	if baseURLEnv := os.Getenv("BASE_URL"); baseURLEnv != "" {
		cfg.BaseURL = baseURLEnv
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080/"
	}

	if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
		cfg.LogLevel = logLevelEnv
	}

	if storagePathEnv := os.Getenv("FILE_STORAGE_PATH"); storagePathEnv != "" {
		cfg.StoragePath = storagePathEnv
	}

	if cfg.StoragePath == "" {
		cfg.StoragePath = "storage.json"
	}

	return &cfg, nil
}
