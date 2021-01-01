package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	statusCode *int
}

// GzipMiddleware ...
func GzipMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}
		var buf bytes.Buffer
		grw := gzipResponseWriter{Writer: &buf, ResponseWriter: w}
		sc := new(int)
		*sc = http.StatusOK
		grw.statusCode = sc
		fn(grw, r)

		detectedContentType := http.DetectContentType(buf.Bytes())
		if canCompress(detectedContentType) {
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Type", detectedContentType)
			w.WriteHeader(*grw.statusCode)

			var gzipWriter *gzip.Writer
			gzipWriter, _ = gzip.NewWriterLevel(w, gzip.BestCompression)
			gzipWriter.Write(buf.Bytes())
			gzipWriter.Close()
		} else {
			w.Write(buf.Bytes())
		}
	}
}

func (grw gzipResponseWriter) Write(b []byte) (int, error) {
	return grw.Writer.Write(b)
}

func (grw gzipResponseWriter) WriteHeader(statusCode int) {
	*grw.statusCode = statusCode
}

func canCompress(contentType string) bool {
	if strings.HasPrefix(contentType, "audio/") || strings.HasPrefix(contentType, "video/") ||
		(strings.HasPrefix(contentType, "image/") && contentType != "image/svg+xml" && contentType != "image/bmp") {
		return false
	}
	return true
}
