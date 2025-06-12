package middleware

import (
	"github.com/rookgm/shortener/internal/client"
	"github.com/rookgm/shortener/internal/logger"
	"net/http"
	"time"
)

const authCookieName = "shortener_auth"

func Auth(authToken client.AuthToken, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("get cookie value")
		cookie, err := r.Cookie(authCookieName)
		if err != nil {
			//switch {
			// cookie not present
			logger.Log.Debug("cookie not present")
			//case errors.Is(err, http.ErrNoCookie):
			// get token
			logger.Log.Debug("create token")
			token, err := authToken.Create()
			if err != nil {
				http.Error(w, "can not create token", http.StatusInternalServerError)
				return
			}
			// create cookie
			http.SetCookie(w, &http.Cookie{
				Name:     authCookieName,
				Value:    token,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				Secure:   true,
				HttpOnly: true,
			})
			r.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
			//default:
			//	logger.Log.Error("cookie error", zap.Error(err))
			//	http.Error(w, "bad cookie", http.StatusInternalServerError)
			//}
		} else {
			// cookie exist
			// validate cookie
			logger.Log.Debug("cookie exist, validate it")
			uid, err := authToken.Verify(cookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			r.AddCookie(&http.Cookie{Name: authCookieName, Value: uid})
		}

		next.ServeHTTP(w, r)
	})
}
