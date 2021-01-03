package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/util/logging"
)

// AddAccount ...
func AddAccount(s *server.Server) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var addAccountRequest struct {
			AccountName string
			Kind        string
		}
		json.NewDecoder(r.Body).Decode(&addAccountRequest)
		sid := s.Sessions.SessionIDFromRequest(r)
		googleUser := server.GoogleUserInfoFromSession(s, sid)
		user, err := s.Services.UserService.GetUser(googleUser.Email)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}
		err = s.Services.UserService.LoadUserAccounts(&user)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}
		for _, acc := range user.Accounts {
			if acc.Name == addAccountRequest.AccountName && acc.Kind == addAccountRequest.Kind {
				rw.WriteHeader(http.StatusConflict)
				logging.Println(logging.Error, "account already exists")
				return
			}
		}
		acc, err := s.Services.AccountService.AddAccount(user.ID, addAccountRequest.AccountName, addAccountRequest.Kind)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}
		resp := struct {
			ID int64
		}{ID: acc.ID}
		json.NewEncoder(rw).Encode(resp)
	}
}

// DeleteAccount ...
func DeleteAccount(s *server.Server) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var delAccountRequest struct {
			AccountID string
			Kind      string
		}
		json.NewDecoder(r.Body).Decode(&delAccountRequest)
		sid := s.Sessions.SessionIDFromRequest(r)
		googleUser := server.GoogleUserInfoFromSession(s, sid)
		user, err := s.Services.UserService.GetUser(googleUser.Email)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}
		accountID, _ := strconv.ParseInt(delAccountRequest.AccountID, 10, 64)
		err = s.Services.AccountService.RemoveAccount(accountID)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}
		err = s.Services.UserService.LoadUserAccountsForSocialMedia(&user, delAccountRequest.Kind)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}
		var lastID int64 = -1
		if len(user.Accounts) > 0 {
			lastID = user.Accounts[len(user.Accounts)-1].ID
		}
		resp := struct {
			ID     int64
			LastID int64
		}{ID: accountID, LastID: lastID}
		json.NewEncoder(rw).Encode(resp)
	}
}
