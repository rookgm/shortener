package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/rookgm/shortener/internal/models"
	"github.com/rookgm/shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetHandler(t *testing.T) {

	fileName := "storage_test.json"
	defer os.Remove(fileName)

	st := storage.NewFileStorage(fileName)

	url := models.ShrURL{
		Alias: "EwHXdJfB",
		URL:   "https://practicum.yandex.ru/",
	}

	err := st.StoreURLCtx(context.Background(), url)
	require.NoError(t, err)

	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name   string
		target string
		want   want
	}{
		{
			name:   "positive_test",
			target: "/" + url.Alias,
			want: want{
				code:     http.StatusTemporaryRedirect,
				response: "https://practicum.yandex.ru/",
			},
		},
		{
			name:   "outside_alias",
			target: "/",
			want: want{
				code:     http.StatusNotFound,
				response: "",
			},
		},
	}

	router := chi.NewRouter()
	router.Get("/{id}", GetHandler(st))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.target, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.response, res.Header.Get("Location"))
		})
	}
}

func TestPostHandler(t *testing.T) {

	st := storage.NewMemStorage()

	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name   string
		header string
		body   string
		want   want
	}{
		{
			name:   "positive_test",
			header: "text/plain",
			body:   "http://practicum.yandex.ru/",
			want: want{
				code:        http.StatusCreated,
				response:    `http://localhost:8080/{id}`,
				contentType: "text/plain",
			},
		},
		{
			name:   "other_Content-Type",
			header: "multipart/form-data",
			body:   "http://practicum.yandex.ru/test",
			want: want{
				code:        http.StatusCreated,
				response:    `http://localhost:8080/{id}`,
				contentType: "text/plain",
			},
		},
		{
			name: "empty_body",
			body: "",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain",
			},
		},
		{
			name:   "test_status_conflict",
			header: "text/plain",
			body:   "http://practicum.yandex.ru/",
			want: want{
				code:        http.StatusConflict,
				response:    `http://localhost:8080/{id}`,
				contentType: "text/plain",
			},
		},
	}

	handler := PostHandler(st, "http://localhost:8080")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.header)
			w := httptest.NewRecorder()

			handler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.NotEmpty(t, resBody, "body body is empty")
			assert.True(t, strings.HasPrefix(res.Header.Get("Content-Type"), string(test.want.contentType)), "Content-Type is not valid")
		})
	}
}

func TestApiShortenHandler(t *testing.T) {

	st := storage.NewMemStorage()

	type want struct {
		code        int
		contentType string
		body        APIShortenResp
	}

	tests := []struct {
		name        string
		contentType string
		body        APIShortenReq
		want        want
	}{
		{
			name:        "positive_test",
			contentType: "application/json",
			body:        APIShortenReq{URL: "https://practicum.yandex.ru/"},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
				body:        APIShortenResp{Result: "http://localhost:8080/"},
			},
		},
		{
			name:        "unsupported_media_type",
			contentType: "text/plain",
			want: want{
				code:        http.StatusUnsupportedMediaType,
				contentType: "text/plain",
			},
		},
	}

	handler := APIShortenHandler(st, "http://localhost:8080")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(test.body)
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
			request.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()

			handler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)

			assert.NotEmpty(t, resBody, "body body is empty")

			assert.True(t, strings.HasPrefix(res.Header.Get("Content-Type"), string(test.want.contentType)), "Content-Type is not valid")

			var resp APIShortenResp

			json.Unmarshal(resBody, &resp)

			assert.True(t, strings.HasPrefix(resp.Result, test.want.body.Result), "body result is not equal")
		})
	}
}
