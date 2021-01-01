package pages

import (
	"net/http"
	"visual-feed-aggregator/src/server"
)

// Logout ...
func Logout(s *server.Server) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		sid := s.Sessions.SessionIDFromRequest(r)
		if sid != "" {
			s.Sessions.Store.RemoveKey(sid, "user")
		}
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
	}
}
