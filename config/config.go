package config

import (
	"flag"
	"os"
	"sync"
)

// Config contains configuration information.
type Config struct {
	ServerAddr  string
	BaseURL     string
	LogLevel    string
	StoragePath string
	DataBaseDSN string
	DebugMode   bool
}

// config default values
const (
	// base server address
	defaultServerAddr = ":8080"
	// base address URL of shortened URLs
	defaultBaseURL = "http://localhost:8080/"
	// default logging level
	defaultLogLevel = "debug"
	// file storage path name
	defaultStoragePath = "/tmp/short-url-db.json"
	// default debug mode
	defaultDebugMode = false
)

// singleton
var (
	once      sync.Once
	singleton *Config
)

// New creates a single instance of config
func New() (*Config, error) {
	once.Do(func() {
		cfg := Config{}

		// init flags
		flag.StringVar(&cfg.ServerAddr, "a", "", "server address")
		flag.StringVar(&cfg.BaseURL, "b", "", "base url")
		flag.StringVar(&cfg.LogLevel, "l", "", "log level")
		flag.StringVar(&cfg.StoragePath, "f", "", "storage path")
		flag.StringVar(&cfg.DataBaseDSN, "d", "", "database address")
		flag.BoolVar(&cfg.DebugMode, "debug", defaultDebugMode, "enable debug mode")

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

		if debugModeEnv := os.Getenv("DEBUG_MODE"); debugModeEnv != "" {
			cfg.DebugMode = debugModeEnv == "true"
		}
		singleton = &cfg
	})

	return singleton, nil
}
