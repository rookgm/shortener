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
	DataBaseDSN string
}

const (
	defaultServerAddr  = ":8080"
	defaultBaseURL     = "http://localhost:8080/"
	defaultLogLevel    = "info"
	defaultStoragePath = "/tmp/short-url-db.json"
)

func Init() (*Config, error) {
	cfg := Config{}

	// init flags
	flag.StringVar(&cfg.ServerAddr, "a", "", "server address")
	flag.StringVar(&cfg.BaseURL, "b", "", "base url")
	flag.StringVar(&cfg.LogLevel, "l", "", "log level")
	flag.StringVar(&cfg.StoragePath, "f", "", "storage path")
	flag.StringVar(&cfg.DataBaseDSN, "d", "", "database address")

	flag.Parse()

	if serverAddrEnv := os.Getenv("SERVER_ADDRESS"); serverAddrEnv != "" {
		cfg.ServerAddr = serverAddrEnv
	}

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = defaultServerAddr
	}

	if baseURLEnv := os.Getenv("BASE_URL"); baseURLEnv != "" {
		cfg.BaseURL = baseURLEnv
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
		cfg.LogLevel = logLevelEnv
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = defaultLogLevel
	}

	if storagePathEnv := os.Getenv("FILE_STORAGE_PATH"); storagePathEnv != "" {
		cfg.StoragePath = storagePathEnv
	}

	if cfg.StoragePath == "" {
		cfg.StoragePath = defaultStoragePath
	}

	if dataBaseDSNEnv := os.Getenv("DATABASE_DSN"); dataBaseDSNEnv != "" {
		cfg.DataBaseDSN = dataBaseDSNEnv
	}

	return &cfg, nil
}
