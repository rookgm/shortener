package middleware

import (
	"github.com/rookgm/shortener/internal/client"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuth(t *testing.T) {

	token := client.NewAuthToken([]byte("secretkey"))
	handler := Auth(token, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	tokenStr, _ := token.Create()

	tests := []struct {
		name        string
		tokenString string
		statusCode  int
	}{
		// token exist
		{
			name:        "authorized",
			tokenString: tokenStr,
			statusCode:  http.StatusOK,
		},
		// empty token
		{
			name:        "empty_token",
			tokenString: "",
			statusCode:  http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			w := httptest.NewRecorder()
			req.AddCookie(&http.Cookie{Name: authCookieName, Value: tt.tokenString})
			handler.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.statusCode, res.StatusCode, "status code is not equal")
		})
	}

	t.Run("create_token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode, "status code is not equal")
	})

}
