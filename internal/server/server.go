package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/rookgm/shortener/internal/random"
)

var storage = map[string]string{}

func GetHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// get url alias from path
	alias := strings.TrimLeft(r.URL.Path, "/")

	url, ok := storage[alias]
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	// get url
	url, err := io.ReadAll(r.Body)
	if err != nil || len(url) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// generating url alias
	alias := random.RandString(6)
	// put it storage
	storage[alias] = string(url)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, "http://localhost:8080/"+alias)

}

func Run() error {

	mux := http.NewServeMux()

	mux.HandleFunc("/", PostHandler)
	mux.HandleFunc("/{id}", GetHandler)

	return http.ListenAndServe(`:8080`, mux)
}
