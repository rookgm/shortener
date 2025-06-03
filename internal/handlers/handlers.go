package handlers

import (
	"encoding/json"
	"errors"
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

		rurl, err := store.GetURLCtx(r.Context(), alias)
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

		statusCode := http.StatusCreated

		// put it storage
		if err := store.StoreURLCtx(r.Context(), iurl); err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				statusCode = http.StatusConflict
				ourl, err := store.GetAliasCtx(r.Context(), iurl.URL)
				if err != nil {
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				iurl = ourl
			} else {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(statusCode)
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

		statusCode := http.StatusCreated

		// put it storage
		if err := store.StoreURLCtx(r.Context(), iurl); err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				statusCode = http.StatusConflict
				ourl, err := store.GetAliasCtx(r.Context(), iurl.URL)
				if err != nil {
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				iurl = ourl
			} else {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

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

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func PostBatchHandler(store storage.URLStorage, baseURL string) http.HandlerFunc {
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

		var batchReq []BatchRequest
		var batchResp []BatchResponse

		logger.Log.Debug("decode batch request")
		if err := json.NewDecoder(r.Body).Decode(&batchReq); err != nil {
			logger.Log.Debug("cannot decode JSON body", zap.Error(err))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if len(batchReq) == 0 {
			logger.Log.Debug("batch request is empty")
			http.Error(w, "batch request is empty", http.StatusBadRequest)
			return
		}

		var batchURL []models.ShrURL
		// prepare batch urls
		for _, breq := range batchReq {
			iurl := models.ShrURL{
				Alias: random.RandString(6),
				URL:   breq.OriginalURL,
			}
			batchURL = append(batchURL, iurl)
		}

		if err := store.StoreBatchURLCtx(r.Context(), batchURL); err != nil {
			logger.Log.Debug("can't save batch urls")
			http.Error(w, "can't save batch urls", http.StatusBadRequest)
			return
		}

		// forming batch result
		for _, breq := range batchReq {
			rurl, err := store.GetAliasCtx(r.Context(), breq.OriginalURL)
			if err != nil {
				continue
			}

			surl, err := url.JoinPath(baseURL, rurl.Alias)
			if err != nil {
				logger.Log.Debug("join url path", zap.Error(err))
				continue
			}

			batchResp = append(batchResp, BatchResponse{
				CorrelationID: breq.CorrelationID,
				ShortURL:      surl,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(batchResp); err != nil {
			logger.Log.Debug("cannot encode JSON body", zap.Error(err))
			return
		}
	}
}
