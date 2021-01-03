package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/util/logging"
)

// ContentCount returns the number of contents for one social media & one account
func ContentCount(s *server.Server) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		sid := s.Sessions.SessionIDFromRequest(r)
		googleUser := server.GoogleUserInfoFromSession(s, sid)
		user, err := s.Services.UserService.GetUser(googleUser.Email)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}

		accountIDStr := r.URL.Query().Get("accountID")
		kindIDStr := r.URL.Query().Get("kind")

		if accountIDStr == "*" {
			accountIDStr = "-1"
		}

		accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
		if err != nil {
			logging.Println(logging.Error, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		count, err := s.Services.ContentService.CountAllContentFor(user.ID, kindIDStr, accountID)
		if err != nil {
			logging.Println(logging.Error, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := struct {
			Count int64
		}{Count: count}
		json.NewEncoder(rw).Encode(resp)
	}
}
