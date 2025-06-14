package middleware

import (
	"github.com/rookgm/shortener/internal/client"
	"github.com/rookgm/shortener/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const authCookieName = "auth_shortener"

func Auth(authToken client.AuthToken, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("try get auth cookie")
		cookie, err := r.Cookie(authCookieName)
		if err != nil {
			logger.Log.Warn("cookie not exist")
			logger.Log.Debug("create token")
			token, err := authToken.Create()
			if err != nil {
				logger.Log.Error("can not create token", zap.Error(err))
				http.Error(w, "can not create token", http.StatusInternalServerError)
				return
			}
			// adds a Set-Cookie header
			logger.Log.Debug("set cookie with new token")
			authSetCookie(w, token)
			// add new token to request cookie for using in handlers
			logger.Log.Debug("add new token to request cookie")
			r.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
		} else {
			logger.Log.Debug("cookie exist, verify it")
			uid, err := authToken.Verify(cookie.Value)
			if err != nil {
				logger.Log.Warn("cannot verify token")
				// create token
				token, err := authToken.Create()
				if err != nil {
					logger.Log.Error("can not create token", zap.Error(err))
					http.Error(w, "can not create token", http.StatusInternalServerError)
					return
				}
				// adds a Set-Cookie header
				logger.Log.Debug("set cookie with new token")
				authSetCookie(w, token)
				// add new token to request cookie for using in handlers
				logger.Log.Debug("add new token to request cookie")
				r.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
				next.ServeHTTP(w, r)
			}

			if uid == "" {
				logger.Log.Error("user id is empty")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			logger.Log.Debug("add uid to request cookie")
			r.AddCookie(&http.Cookie{Name: authCookieName, Value: uid})
		}
		next.ServeHTTP(w, r)
	})
}

func authSetCookie(w http.ResponseWriter, tokenString string) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
}
