package app

import (
	"io"
	"net/http"
	"strings"

	"github.com/rookgm/shortener/internal/random"
)

var storage = map[string]string{}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом GET
	if r.Method != http.MethodGet {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if val := r.Header.Get("Content-Type"); strings.Compare(val, "text/plain") != 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	/*parse request*/

	// get url alias from path
	alias := strings.TrimLeft(r.URL.Path, "/")

	url, ok := storage[alias]
	if !ok {
		return
	}

	/*write response*/
	w.WriteHeader(http.StatusTemporaryRedirect)
	io.WriteString(w, "Location: "+url)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// TODO uncomment
	if val := r.Header.Get("Content-Type"); strings.Compare(val, "text/plain") != 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	/*parse request*/
	// get url
	url, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	// generating url alias
	alias := random.RandString(6)
	// put it storage
	storage[alias] = string(url)

	/*write response*/
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
