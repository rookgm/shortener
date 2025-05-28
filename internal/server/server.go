package server

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/db"
	"github.com/rookgm/shortener/internal/handlers"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/random"
	"github.com/rookgm/shortener/internal/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var store *storage.Storage

func GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url alias from path
		alias := chi.URLParam(r, "id")

		url, err := store.Get(alias)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func PostHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// generating url alias
		alias := random.RandString(6)
		// put it storage
		if err := store.Set(alias, string(body)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		rurl, err := url.JoinPath(baseURL, alias)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(rurl))
	}
}

type APIShortenReq struct {
	URL string `json:"url"`
}
type APIShortenResp struct {
	Result string `json:"result"`
}

func APIShortenHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("check Content-Type")
		if ct := r.Header.Get("Content-Type"); ct != "" {
			st := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
			if !strings.Contains(st, "application/json") {
				msg := "Content-Type is not application/json"
				logger.Log.Debug(msg, zap.String("is", ct))
				http.Error(w, msg, http.StatusUnsupportedMediaType)
				return
			}
		}

		var req APIShortenReq

		logger.Log.Debug("decode request")
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Log.Debug("cannot decode JSON body", zap.Error(err))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// generating url alias
		alias := random.RandString(6)
		// put it storage
		if err := store.Set(alias, req.URL); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		var resp APIShortenResp
		var err error

		resp.Result, err = url.JoinPath(baseURL, alias)
		if err != nil {
			logger.Log.Debug("join url path", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Log.Debug("cannot encode JSON body", zap.Error(err))
			return
		}
	}
}

func Run(config *config.Config) error {
	if config == nil {
		return errors.New("config is nil")
	}

	sdb, err := db.Open(config.DataBaseDSN)
	if err != nil {
		return errors.New("can open db")
	}
	defer sdb.Close()

	store = storage.NewStorage(config.StoragePath)
	if store == nil {
		return errors.New("can not create storage")
	}

	if err := store.LoadFromFile(); err != nil {
		return err
	}

	router := chi.NewRouter()
	router.Use(logger.Middleware)
	router.Use(GzipMiddleware)

	router.Route("/", func(r chi.Router) {
		router.Post("/", PostHandler(config.BaseURL))
		router.Get("/{id}", GetHandler())
		router.Post("/api/shorten", APIShortenHandler(config.BaseURL))
		router.Get("/ping", handlers.PingHandler(sdb))
	})

	return http.ListenAndServe(config.ServerAddr, router)
}
