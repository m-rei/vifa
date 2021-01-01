package middleware

import (
	"net/http"
)

type injectedResponseWriter struct {
	http.ResponseWriter
}

// AutoDetectContentType ...
func AutoDetectContentType(f http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		f(injectedResponseWriter{ResponseWriter: rw}, r)
	}
}

func (irw injectedResponseWriter) Write(b []byte) (int, error) {
	irw.ResponseWriter.Header().Set("Content-Type", http.DetectContentType(b))
	return irw.ResponseWriter.Write(b)
}
