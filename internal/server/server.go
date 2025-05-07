package server

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/random"
	"io"
	"net/http"
	"net/url"
)

var storage = map[string]string{}

func GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url alias from path
		alias := chi.URLParam(r, "id")

		url, ok := storage[alias]
		if !ok {
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
		storage[alias] = string(body)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		rurl, err := url.JoinPath(baseURL, alias)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		io.WriteString(w, rurl)
	}
}

func Run(config *config.Config) error {
	if config == nil {
		return errors.New("config is nil")
	}

	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/", func(r chi.Router) {
		router.Post("/", PostHandler(config.BaseURL))
		router.Get("/{id}", GetHandler())
	})

	return http.ListenAndServe(config.ServerAddr, router)
}
