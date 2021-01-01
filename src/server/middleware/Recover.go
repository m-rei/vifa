package middleware

import (
	"net/http"
	"visual-feed-aggregator/src/util/logging"
)

// Recover ...
func Recover(f http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				logging.Println(logging.Error, "!Recovery!")
				logging.Println(logging.Error, r)
			}
		}()
		f(rw, r)
	}
}
