package middleware

import "net/http"

// DefaultHeaders disables caching and sets a few security headers
func DefaultHeaders(f http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		f(rw, r)
		rw.Header().Set("Content-Security-Policy", "img-src https: data:;"+
			"font-src https: data:;"+
			"connect-src https: data:;")
		rw.Header().Set("X-XSS-Protection", "1; mode=blockFilter") // for legacy browsers, which dont implement content security policy fully yet
		rw.Header().Set("X-Content-Type-Options", "nosniff")       // disable browser content-type auto detection through reading first few bytes
		rw.Header().Set("X-Frame-Options", "no-cache")             // prevent being displayed in an iframe

	}
}
