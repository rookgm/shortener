package handlers

import (
	"github.com/rookgm/shortener/internal/db"
	"net/http"
)

func PingHandler(sdb *db.DataBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := sdb.PingCtx(r.Context()); err != nil {
			// failed: 500 Internal Server Error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 200 OK
		w.WriteHeader(http.StatusOK)
	}
}
