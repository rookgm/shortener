package server

import (
	"compress/gzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {

	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"url": "https://practicum.yandex.ru"}`))
	}))

	tests := []struct {
		name           string
		acceptEncoding string
		statusCode     int
		body           string
	}{
		{
			name:           "gzip",
			acceptEncoding: "gzip",
			statusCode:     http.StatusOK,
			body:           `{"url": "https://practicum.yandex.ru"}`,
		},
		{
			name:           "no_gzip",
			acceptEncoding: "",
			statusCode:     http.StatusOK,
			body:           `{"url": "https://practicum.yandex.ru"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			respRec := httptest.NewRecorder()

			handler.ServeHTTP(respRec, req)

			res := respRec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.statusCode, res.StatusCode, "status code is not equal")

			var body []byte
			var err error

			if strings.Contains(res.Header.Get("Content-Encoding"), "gzip") {
				gr, err := gzip.NewReader(res.Body)
				require.NoError(t, err)

				body, err = io.ReadAll(gr)
				require.NoError(t, err)

				defer gr.Close()
			} else {
				body, err = io.ReadAll(res.Body)
			}

			require.NoError(t, err)
			assert.Equal(t, tt.body, string(body))

		})
	}
}
