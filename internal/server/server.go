package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rookgm/shortener/internal/random"
	"io"
	"net/http"
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

func PostHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url
		url, err := io.ReadAll(r.Body)
		if err != nil || len(url) == 0 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// generating url alias
		alias := random.RandString(6)
		// put it storage
		storage[alias] = string(url)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, "http://localhost:8080/"+alias)
	}
}

func DeleteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//alias := strings.TrimLeft(r.URL.Path, "/")
		fmt.Println("test")
	}
}

func Run() error {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/", func(r chi.Router) {
		router.Post("/", PostHandler())
		router.Get("/{id}", GetHandler())
	})

	return http.ListenAndServe(`:8080`, router)
}
