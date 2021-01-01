package pages

import (
	"fmt"
	"net/http"
	"visual-feed-aggregator/src/util/logging"
)

// NotFound ...
func NotFound(reditURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logging.Println(logging.Debug, fmt.Sprintf("Site '%s' not found", r.RequestURI))
		http.Redirect(w, r, reditURL, http.StatusMovedPermanently)
	}
}
