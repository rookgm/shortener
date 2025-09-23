package handlers

import (
	"encoding/json"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

// StatsResponse is handler response in JSON format
type StatsResponse struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}

// StatsHandler returns the number of shortened urls and users in the service.
//
// Request
// GET /api/internal/stats HTTP/1.1
// Host: localhost:8080
// Content-Type: text/plain
//
// # Response
//
// HTTP/1.1 200 OK
// Content-Type: application/json
// Content-Length: 21
//
// {"urls":1,"users":1}
func StatsHandler(store storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp StatsResponse
		var err error

		// get the number of shortened urls
		resp.Urls, err = store.GetURLCountCtx(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// get user count
		resp.Users, err = store.GetUserCountCtx(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Log.Debug("cannot encode JSON body", zap.Error(err))
			return
		}
	}
}
