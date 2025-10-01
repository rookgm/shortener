package config

import (
	"encoding/json"
	"flag"
	"os"
)

// Config contains configuration information.
type Config struct {
	ServerAddr    string
	BaseURL       string
	LogLevel      string
	StoragePath   string
	DataBaseDSN   string
	DebugMode     bool
	EnableHTTPS   bool
	ConfigPath    string
	TrustedSubNet string
	EnableGRPC    bool
	GRPCSeverAddr string
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
	// set https
	defaultHTTPS = false
	// grpc servers is disabled
	defaultGRPS = false
	// grpc server address
	defaultGRPCServerAddr = ":3200"
)

// Option is config func option
type Option func(*Config)

// WithServerAddr sets server address in Config
func WithServerAddr(addr string) Option {
	return func(c *Config) {
		if addr != "" {
			c.ServerAddr = addr
		}
	}
}

// WithBaseURL sets base url in Config
func WithBaseURL(url string) Option {
	return func(c *Config) {
		if url != "" {
			c.BaseURL = url
		}
	}
}

// WithLogLevel sets logging level
func WithLogLevel(level string) Option {
	return func(c *Config) {
		if level != "" {
			c.LogLevel = level
		}
	}
}

// WithStoragePath sets storage path
func WithStoragePath(path string) Option {
	return func(c *Config) {
		if path != "" {
			c.StoragePath = path
		}
	}
}

// WithDatabaseDSN sets data source name
func WithDatabaseDSN(dsn string) Option {
	return func(c *Config) {
		if dsn != "" {
			c.DataBaseDSN = dsn
		}
	}
}

// WithDebugMode sets debug mode
func WithDebugMode(mode bool) Option {
	return func(c *Config) {
		c.DebugMode = mode
	}
}

// WithEnableHTTPS sets enabling https
func WithEnableHTTPS(enable bool) Option {
	return func(c *Config) {
		c.EnableHTTPS = enable
	}
}

// WithTrustedSubNet sets trusted subnet
func WithTrustedSubNet(s string) Option {
	return func(c *Config) {
		if s != "" {
			c.TrustedSubNet = s
		}
	}
}

// WithEnableGRPC sets enabling grpc
func WithEnableGRPC(enable bool) Option {
	return func(c *Config) {
		c.EnableGRPC = enable
	}
}

// WithGRPCServerAddr sets grpc server address in Config
func WithGRPCServerAddr(addr string) Option {
	return func(c *Config) {
		if addr != "" {
			c.GRPCSeverAddr = addr
		}
	}
}

// configJSON is config in json format
type configJSON struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	TrustedSubNet   string `json:"trusted_subnet"`
	EnableGRPC      bool   `json:"enable_grpc"`
	GRPCServerAddr  string `json:"grpc_server_addr"`
}

// FromFile loads config from file in JSON format
func FromFile(name string) Option {
	return func(c *Config) {
		if name == "" {
			return
		}

		b, err := os.ReadFile(name)
		if err != nil {
			return
		}

		cfg := configJSON{}

		err = json.Unmarshal(b, &cfg)
		if err != nil {
			return
		}

		WithServerAddr(cfg.ServerAddress)(c)
		WithBaseURL(cfg.BaseURL)(c)
		WithStoragePath(cfg.FileStoragePath)(c)
		WithDatabaseDSN(cfg.DatabaseDSN)(c)
		WithEnableHTTPS(cfg.EnableHTTPS)(c)
		WithTrustedSubNet(cfg.TrustedSubNet)(c)
		WithEnableGRPC(cfg.EnableGRPC)(c)
		WithGRPCServerAddr(cfg.GRPCServerAddr)(c)
	}
}

// FromEnv gets configuration from environment variables
func FromEnv() Option {
	return func(c *Config) {

		// sets base server address
		if serverAddrEnv := os.Getenv("SERVER_ADDRESS"); serverAddrEnv != "" {
			WithServerAddr(serverAddrEnv)(c)
		}
		// sets base address URL of shortened URLs
		if baseURLEnv := os.Getenv("BASE_URL"); baseURLEnv != "" {
			WithBaseURL(baseURLEnv)(c)
		}
		// sets logging level
		if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
			WithLogLevel(logLevelEnv)(c)
		}
		// sets file storage path
		if storagePathEnv := os.Getenv("FILE_STORAGE_PATH"); storagePathEnv != "" {
			WithStoragePath(storagePathEnv)(c)
		}
		// sets database source name
		if dataBaseDSNEnv := os.Getenv("DATABASE_DSN"); dataBaseDSNEnv != "" {
			WithDatabaseDSN(dataBaseDSNEnv)(c)
		}
		// sets debug mode
		if debugModeEnv := os.Getenv("DEBUG_MODE"); debugModeEnv == "true" {
			WithDebugMode(true)(c)
		}
		// sets https support
		if httpsEnv := os.Getenv("ENABLE_HTTPS"); httpsEnv == "true" {
			WithEnableHTTPS(true)(c)
		}
		// sets trusted subnet
		if subNetEnv := os.Getenv("TRUSTED_SUBNET"); subNetEnv != "" {
			WithTrustedSubNet(subNetEnv)(c)
		}
		// sets grpc support
		if grpcEnv := os.Getenv("ENABLE_GRPC"); grpcEnv == "true" {
			WithEnableGRPC(true)(c)
		}
		// sets grpc server address
		if grpcAddrEnv := os.Getenv("GRPC_SERVER_ADDRESS"); grpcAddrEnv != "" {
			WithGRPCServerAddr(grpcAddrEnv)(c)
		}
	}
}

// FromCommandLine gets configuration from command line
func FromCommandLine(args *Config) Option {
	return func(c *Config) {
		WithServerAddr(args.ServerAddr)(c)
		WithBaseURL(args.BaseURL)(c)
		WithLogLevel(args.LogLevel)(c)
		WithStoragePath(args.StoragePath)(c)
		WithDatabaseDSN(args.DataBaseDSN)(c)
		WithDebugMode(args.DebugMode)(c)
		WithEnableHTTPS(args.EnableHTTPS)(c)
		WithTrustedSubNet(args.TrustedSubNet)(c)
		WithEnableGRPC(args.EnableGRPC)(c)
		WithGRPCServerAddr(args.GRPCSeverAddr)(c)
	}
}

// parseCommandLine parses command line arguments
func parseCommandLine(cfg *Config) {
	flag.StringVar(&cfg.ServerAddr, "a", "", "server address")
	flag.StringVar(&cfg.BaseURL, "b", "", "base url")
	flag.StringVar(&cfg.LogLevel, "l", "", "log level")
	flag.StringVar(&cfg.StoragePath, "f", "", "storage path")
	flag.StringVar(&cfg.DataBaseDSN, "d", "", "database address")
	flag.BoolVar(&cfg.DebugMode, "debug", false, "enable debug mode")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "enable https")
	flag.StringVar(&cfg.ConfigPath, "config", "", "load config from file")
	flag.StringVar(&cfg.ConfigPath, "c", "", "load config from file")
	flag.StringVar(&cfg.TrustedSubNet, "t", "", "trusted subnet")
	flag.BoolVar(&cfg.EnableGRPC, "g", false, "enable grpc server")
	flag.StringVar(&cfg.GRPCSeverAddr, "p", "", "grpc server address")

	flag.Parse()
}

// New returns new Config. It parses command line, environment variables and file.
func New(opts ...Option) (*Config, error) {
	// set defaults values
	cfg := &Config{
		ServerAddr:    defaultServerAddr,
		BaseURL:       defaultBaseURL,
		LogLevel:      defaultLogLevel,
		StoragePath:   defaultStoragePath,
		DebugMode:     defaultDebugMode,
		EnableHTTPS:   defaultHTTPS,
		EnableGRPC:    defaultGRPS,
		GRPCSeverAddr: defaultGRPCServerAddr,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg, nil
}

// Initialize initializes the configuration
func Initialize() (*Config, error) {
	args := &Config{}
	// parse command line
	parseCommandLine(args)
	return New(
		// low priority
		FromFile(args.ConfigPath),
		// medium priority
		FromEnv(),
		// height priority
		FromCommandLine(args),
	)
}
