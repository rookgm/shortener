package main

import (
	"context"
	"errors"
	"log"

	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/db"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/server"
	"github.com/rookgm/shortener/internal/storage"
	"go.uber.org/zap"
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

	store, fl, sdb, err := initStorage(cfg)
	if err != nil {
		logger.Log.Fatal("Cannot initialize storage", zap.Error(err))
	}

	if err := server.Run(cfg, store, fl, sdb); err != nil {
		logger.Log.Fatal("Cannot start server", zap.Error(err))
	}
}

func initStorage(cfg *config.Config) (store storage.URLStorage, sl server.StorageLoader, sdb *db.DataBase, err error) {
	if cfg.DataBaseDSN != "" {
		// open shortener db
		sdb, err := db.OpenCtx(context.Background(), cfg.DataBaseDSN)
		if err != nil {
			return nil, nil, nil, errors.New("can open db")
		}
		defer sdb.Close()
		// create db storage
		store, err = storage.NewDBStorage(sdb)
		if err != nil {
			return nil, nil, nil, err
		}
	} else if cfg.StoragePath != "" {
		// create file storage
		// load storage from file
		fs := storage.NewFileStorage(cfg.StoragePath)
		store = fs
		sl = fs
	} else {
		// create storage on memory
		store = storage.NewMemStorage()
	}

	return store, sl, sdb, nil
}
