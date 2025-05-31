package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// valid content types
var validContentTypes = []string{"application/json", "text/html"}

type ContentTypeChecker struct {
	m    map[string]struct{}
	once sync.Once
}

// IsValid is checks the acceptable content type
func (ch *ContentTypeChecker) IsValid(s string) bool {
	ch.once.Do(func() {
		ch.m = make(map[string]struct{})
		for _, v := range validContentTypes {
			ch.m[v] = struct{}{}
		}
	})

	_, ok := ch.m[s]
	return ok
}

// content type checker
var ctChecker ContentTypeChecker

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	// only gzip
	if strings.Compare(c.Header().Get("Content-Encoding"), "gzip") == 0 {
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	// check content type
	if ctChecker.IsValid(c.Header().Get("Content-Type")) {
		// set content-type
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	// only gzip
	if strings.Compare(c.Header().Get("Content-Encoding"), "gzip") == 0 {
		return c.zw.Close()
	}
	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		curw := w
		// compress response
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			zw := newCompressWriter(w)
			curw = zw
			defer zw.Close()
		}
		// uncompressed request
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		next.ServeHTTP(curw, r)
	})
}
