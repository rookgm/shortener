package handlers

import (
	"encoding/json"
	"github.com/rookgm/shortener/internal/client"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/models"
	"github.com/rookgm/shortener/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
)

const numWorkers = 10

// UserURL represent user's url(short and original)
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// GetUserUrlsHandler returns all urls to the user (route /api/user/urls)
func GetUserUrlsHandler(store storage.URLStorage, baseURL string, token client.AuthToken) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract user ID from request cookie
		uid := token.GetUserID(r)
		// get all user urls by id from storage
		logger.Log.Debug("trying get user urls", zap.String("id", uid))
		uurls, err := store.GetUserURLsCtx(r.Context(), uid)
		if err != nil {
			logger.Log.Error("get user urls from storage", zap.Error(err))
			http.Error(w, "can't get user urls", http.StatusInternalServerError)
			return
		}
		status := http.StatusOK

		w.Header().Set("Content-Type", "application/json")

		// user url is not exist
		if len(uurls) == 0 {
			logger.Log.Warn("user urls does not exist", zap.String("id", uid))
			status = http.StatusNoContent
			http.Error(w, "can't get user urls", http.StatusNoContent)
			return
		}

		w.WriteHeader(status)

		// output user urls
		var userURLResp []UserURL

		// prepare user urls
		for _, uurl := range uurls {
			urlPath, err := url.JoinPath(baseURL, uurl.Alias)
			if err != nil {
				logger.Log.Error("join url path", zap.Error(err))
				continue
			}
			res := UserURL{
				ShortURL:    urlPath,
				OriginalURL: uurl.URL,
			}
			userURLResp = append(userURLResp, res)
		}
		// encode user urls to json and put it to response
		if err := json.NewEncoder(w).Encode(userURLResp); err != nil {
			logger.Log.Error("cannot encode JSON body", zap.Error(err))
			return
		}
	}
}

// DeleteUserUrlsHandler deletes user urls
func DeleteUserUrlsHandler(store storage.URLStorage, token client.AuthToken, fanInCh chan<- models.UserDeleteTask) http.HandlerFunc {
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
		var aliasToDelete []string

		logger.Log.Debug("decode request")
		if err := json.NewDecoder(r.Body).Decode(&aliasToDelete); err != nil {
			logger.Log.Debug("cannot decode JSON body", zap.Error(err))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// extract user ID from request cookie
		uid := token.GetUserID(r)

		// pass user aliases to delete worker
		fanInCh <- models.UserDeleteTask{
			UID:     uid,
			Aliases: aliasToDelete,
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
