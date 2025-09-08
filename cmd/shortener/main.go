package main

import (
	"fmt"
	"log"

	"github.com/rookgm/shortener/internal/server"

	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/logger"
	"go.uber.org/zap"
)

// application build info
var (
	// BuildVersion is application build version
	BuildVersion = "N/A"
	// BuildDate is application build date
	BuildDate = "N/A"
	// BuildCommit is application build commit
	BuildCommit = "N/A"
)

// printBuildInfo prints application build info to stdout
func printBuildInfo() {
	fmt.Printf(
		"Build version: %s\n"+
			"Build date: %s\n"+
			"Build commit: %s\n",
		BuildVersion,
		BuildDate,
		BuildCommit)
}

func main() {

	printBuildInfo()

	cfg, err := config.Initialize()
	if err != nil {
		log.Fatalf("Cannot initialize config: %v\n", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("Cannot initialize logger: %v\n", err)
	}

	logger.Log.Info("Running server", zap.String("addr", cfg.ServerAddr))
	if err := server.Run(cfg); err != nil {
		logger.Log.Fatal("Cannot start server", zap.Error(err))
	}
}
