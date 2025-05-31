package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/rookgm/shortener/internal/db"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/models"
	"github.com/rookgm/shortener/internal/random"
	"github.com/rookgm/shortener/internal/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func GetHandler(store storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url alias from path
		alias := chi.URLParam(r, "id")

		rurl, err := store.GetURL(alias)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, rurl.URL, http.StatusTemporaryRedirect)
	}
}

func PostHandler(store storage.URLStorage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		iurl := models.ShrURL{
			Alias: random.RandString(6),
			URL:   string(body),
		}

		// put it storage
		if err := store.StoreURL(iurl); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		rurl, err := url.JoinPath(baseURL, iurl.Alias)
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

func APIShortenHandler(store storage.URLStorage, baseURL string) http.HandlerFunc {
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

		iurl := models.ShrURL{
			Alias: random.RandString(6),
			URL:   req.URL,
		}

		// put it storage
		if err := store.StoreURL(iurl); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		var resp APIShortenResp
		var err error

		resp.Result, err = url.JoinPath(baseURL, iurl.Alias)
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

func PingHandler(sdb *db.DataBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := sdb.PingCtx(r.Context()); err != nil {
			// failed: 500 Internal Server Error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 200 OK
		w.WriteHeader(http.StatusOK)
	}
}
