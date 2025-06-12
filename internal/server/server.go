package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/db"
	"github.com/rookgm/shortener/internal/handlers"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/middleware"
	"github.com/rookgm/shortener/internal/storage"
)

type StorageLoader interface {
	LoadFromFile() error
}

func Run(config *config.Config, store storage.URLStorage, sl StorageLoader, sdb *db.DataBase) error {
	if config == nil {
		return errors.New("config is nil")
	}

	if sl != nil {
		if err := sl.LoadFromFile(); err != nil {
			return err
		}
	}

	router := chi.NewRouter()
	router.Use(logger.Middleware)
	router.Use(middleware.GzipMiddleware)

	router.Route("/", func(r chi.Router) {
		router.Post("/", handlers.PostHandler(store, config.BaseURL))
		router.Get("/{id}", handlers.GetHandler(store))
		router.Post("/api/shorten", handlers.APIShortenHandler(store, config.BaseURL))
		router.Get("/ping", handlers.PingHandler(sdb))
		router.Post("/api/shorten/batch", handlers.PostBatchHandler(store, config.BaseURL))
	})

	return http.ListenAndServe(config.ServerAddr, router)
}
