package middleware

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckTrustedSubNet(t *testing.T) {
	handler := CheckTrustedSubNet("192.168.1.0/24", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	tests := []struct {
		name       string
		clientIP   string
		statusCode int
	}{
		{
			name:       "trusted_client_ip_return_200",
			clientIP:   "192.168.1.1",
			statusCode: http.StatusOK,
		},
		{
			name:       "not_trusted_client_ip_return_403",
			clientIP:   "127.0.0.1",
			statusCode: http.StatusForbidden,
		},
		{
			name:       "bad_client_ip_return_400",
			clientIP:   "",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			req.Header.Set("X-Real-IP", tt.clientIP)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.statusCode, res.StatusCode, "status code is not equal")
		})
	}
}
