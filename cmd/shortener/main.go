package main

import (
	"fmt"
	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/server"
	"go.uber.org/zap"
	"log"
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
	if err := server.Run(cfg); err != nil {
		logger.Log.Fatal("Cannot run server", zap.Error(err))
	}
}
