package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Log is no-op Logger
var Log *zap.Logger = zap.NewNop()

type (
	responseData struct {
		status int
		size   int
	}

	responseWrite struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Initialize init logger
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

// Write writes data
func (rw *responseWrite) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.responseData.size += size
	return size, err
}

// WriteHeader writes header with status code
func (rw *responseWrite) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.responseData.status = statusCode
}

// Middleware is middleware logger
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ts := time.Now()
		responseData := &responseData{
			status: http.StatusOK,
			size:   0,
		}

		lrw := responseWrite{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lrw, r)

		dt := time.Since(ts)

		Log.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Duration("duration", dt),
		)
	})
}
