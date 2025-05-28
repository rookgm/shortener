package main

import (
	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/server"
	"go.uber.org/zap"
	"log"
)

func main() {

	cfg, err := config.Init()
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
