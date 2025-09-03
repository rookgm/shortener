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
	EnableHTTPS bool
}

// config default values
const (
	// base server address
	defaultServerAddr = ":8080"
	// base https server address
	defaultHTTPSServerAddr = ":8443"
	// base address URL of shortened URLs
	defaultBaseURL = "http://localhost:8080/"
	// base address URL of shortened URLs
	defaultHTTPSBaseURL = "https://localhost:8443/"
	// default logging level
	defaultLogLevel = "debug"
	// file storage path name
	defaultStoragePath = "/tmp/short-url-db.json"
	// default debug mode
	defaultDebugMode = false
	// set https
	defaultHTTPS = false
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
		flag.BoolVar(&cfg.EnableHTTPS, "s", defaultHTTPS, "enable https")
		flag.StringVar(&cfg.ServerAddr, "a", "", "server address")
		flag.StringVar(&cfg.BaseURL, "b", "", "base url")
		flag.StringVar(&cfg.LogLevel, "l", "", "log level")
		flag.StringVar(&cfg.StoragePath, "f", "", "storage path")
		flag.StringVar(&cfg.DataBaseDSN, "d", "", "database address")
		flag.BoolVar(&cfg.DebugMode, "debug", defaultDebugMode, "enable debug mode")

		flag.Parse()

		// sets https support
		if httpsEnv := os.Getenv("ENABLE_HTTPS"); httpsEnv != "" {
			cfg.EnableHTTPS = httpsEnv == "true"
		}

		// sets base server address
		if serverAddrEnv := os.Getenv("SERVER_ADDRESS"); serverAddrEnv != "" {
			cfg.ServerAddr = serverAddrEnv
		}
		if cfg.ServerAddr == "" {
			// if https is enabled
			if cfg.EnableHTTPS {
				cfg.ServerAddr = defaultHTTPSServerAddr
			} else {
				cfg.ServerAddr = defaultServerAddr
			}
		}

		// sets base address URL of shortened URLs
		if baseURLEnv := os.Getenv("BASE_URL"); baseURLEnv != "" {
			cfg.BaseURL = baseURLEnv
		}
		if cfg.BaseURL == "" {
			if cfg.EnableHTTPS {
				cfg.BaseURL = defaultHTTPSBaseURL
			} else {
				cfg.BaseURL = defaultBaseURL
			}
		}

		// sets logging level
		if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
			cfg.LogLevel = logLevelEnv
		}
		if cfg.LogLevel == "" {
			cfg.LogLevel = defaultLogLevel
		}

		// sets file storage path
		if storagePathEnv := os.Getenv("FILE_STORAGE_PATH"); storagePathEnv != "" {
			cfg.StoragePath = storagePathEnv
		}
		if cfg.StoragePath == "" {
			cfg.StoragePath = defaultStoragePath
		}

		// sets database source namse
		if dataBaseDSNEnv := os.Getenv("DATABASE_DSN"); dataBaseDSNEnv != "" {
			cfg.DataBaseDSN = dataBaseDSNEnv
		}

		// sets debug mode
		if debugModeEnv := os.Getenv("DEBUG_MODE"); debugModeEnv != "" {
			cfg.DebugMode = debugModeEnv == "true"
		}

		singleton = &cfg
	})

	return singleton, nil
}
