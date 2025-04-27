package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetHandler(t *testing.T) {

	// initialize map
	id := "EwHXdJfB"
	storage[id] = "https://practicum.yandex.ru/"

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
			name:   "positive test",
			target: "/" + id,
			want: want{
				code:     http.StatusTemporaryRedirect,
				response: "https://practicum.yandex.ru/",
			},
		},
		{
			name:   "outside alias",
			target: "/",
			want: want{
				code:     http.StatusNotFound,
				response: "",
			},
		},
	}

	router := chi.NewRouter()
	router.Get("/{id}", GetHandler())

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
			name:   "positive test",
			header: "text/plain",
			body:   "http://practicum.yandex.ru/",
			want: want{
				code:        http.StatusCreated,
				response:    `http://localhost:8080/{id}`,
				contentType: "text/plain",
			},
		},
		{
			name:   "bad Content-Type",
			header: "multipart/form-data",
			body:   "http://practicum.yandex.ru/",
			want: want{
				code:        http.StatusCreated,
				response:    `http://localhost:8080/{id}`,
				contentType: "text/plain",
			},
		},
		{
			name: "empty body",
			body: "",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain",
			},
		},
	}

	handler := PostHandler()

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
			assert.NotEmpty(t, resBody, "response body is empty")
			assert.True(t, strings.HasPrefix(res.Header.Get("Content-Type"), string(test.want.contentType)), "Content-Type is not valid")
		})
	}
}
