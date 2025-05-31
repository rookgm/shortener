package server

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/db"
	"github.com/rookgm/shortener/internal/handlers"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/middleware"
	"github.com/rookgm/shortener/internal/storage"
	"net/http"
)

func Run(config *config.Config) error {
	if config == nil {
		return errors.New("config is nil")
	}

	sdb, err := db.Open(config.DataBaseDSN)
	if err != nil {
		return errors.New("can open db")
	}
	defer sdb.Close()

	var st storage.URLStorage

	// TODO
	/*if config.DataBaseDSN != "" {
		// create db storage
	}
	*/
	if config.StoragePath != "" {
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

	router := chi.NewRouter()
	router.Use(logger.Middleware)
	router.Use(middleware.GzipMiddleware)

	router.Route("/", func(r chi.Router) {
		router.Post("/", handlers.PostHandler(st, config.BaseURL))
		router.Get("/{id}", handlers.GetHandler(st))
		router.Post("/api/shorten", handlers.APIShortenHandler(st, config.BaseURL))
		router.Get("/ping", handlers.PingHandler(sdb))
	})

	return http.ListenAndServe(config.ServerAddr, router)
}
