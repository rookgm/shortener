package server

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/client"
	"github.com/rookgm/shortener/internal/db"
	"github.com/rookgm/shortener/internal/handlers"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/middleware"
	"github.com/rookgm/shortener/internal/storage"
	"net/http"
)

const authTokenKey = "f53ac685bbceebd75043e6be2e06ee07"

func Run(config *config.Config) error {
	if config == nil {
		return errors.New("config is nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var sdb *db.DataBase
	var st storage.URLStorage
	var err error

	if config.DataBaseDSN != "" {
		// open shortener db
		sdb, err = db.OpenCtx(ctx, config.DataBaseDSN)
		if err != nil {
			return errors.New("can open db")
		}
		defer sdb.Close()
		// create db storage
		st, err = storage.NewDBStorage(sdb)
		if err != nil {
			return err
		}
	} else if config.StoragePath != "" {
		// create file storage
		st = storage.NewFileStorage(config.StoragePath)
		// load storage from file
		if err := st.LoadFromFile(); err != nil {
			return err
		}
	} else {
		// create storage on memory
		st = storage.NewMemStorage()
	}

	key, err := hex.DecodeString(authTokenKey)
	if err != nil {
		logger.Log.Error("can not extract key")
		return err
	}

	token := client.NewAuthToken(key)

	router := chi.NewRouter()
	router.Use(logger.Middleware)
	router.Use(middleware.GzipMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return middleware.Auth(token, next)
	})

	router.Route("/", func(r chi.Router) {
		router.Post("/", handlers.PostHandler(st, config.BaseURL, token))
		router.Get("/{id}", handlers.GetHandler(st))
		router.Post("/api/shorten", handlers.APIShortenHandler(st, config.BaseURL))
		router.Get("/ping", handlers.PingHandler(sdb))
		router.Post("/api/shorten/batch", handlers.PostBatchHandler(st, config.BaseURL))
		router.Get("/api/user/urls", handlers.GetUserUrls(st, config.BaseURL, token))
	})

	return http.ListenAndServe(config.ServerAddr, router)
}
