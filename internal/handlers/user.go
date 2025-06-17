package handlers

import (
	"context"
	"encoding/json"
	"github.com/rookgm/shortener/internal/client"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
	"sync"
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
func DeleteUserUrlsHandler(store storage.URLStorage, token client.AuthToken) http.HandlerFunc {
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

		if err := deleteBatchUserUrlsCtx(r.Context(), store, uid, aliasToDelete); err != nil {
			logger.Log.Error("can't delete user urls", zap.String("id", uid))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

// generator writes alias to channel
func generator(done <-chan struct{}, input []string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, data := range input {
			select {
			case out <- data:
			case <-done:
				return
			}
		}
	}()

	return out
}

// fanIn combines the data streams from multiple sources into one
func fanIn(done <-chan struct{}, inputs ...<-chan string) <-chan string {
	output := make(chan string, 10)
	var wg sync.WaitGroup

	for _, input := range inputs {
		wg.Add(1)
		go func(ch <-chan string) {
			defer wg.Done()
			for val := range ch {
				select {
				case output <- val:
				case <-done:
					return
				}
			}
		}(input)
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// deleteBatchUserUrlsCtx deletes batch of user's urls
func deleteBatchUserUrlsCtx(ctx context.Context, store storage.URLStorage, userID string, aliases []string) error {
	done := make(chan struct{})
	defer close(done)

	input := generator(done, aliases)
	result := fanIn(done, input)

	var batch []string

	for res := range result {
		batch = append(batch, res)
	}

	if err := store.DeleteUserURLsCtx(ctx, userID, batch); err != nil {
		return err
	}

	return nil
}
