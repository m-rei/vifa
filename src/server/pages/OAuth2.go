package pages

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/util/logging"

	"golang.org/x/oauth2"
)

// GoogleEndpoint ...
var (
	GoogleEndpoint oauth2.Endpoint = oauth2.Endpoint{
		AuthURL:  "https://accounts.google.com/o/oauth2/auth",
		TokenURL: "https://oauth2.googleapis.com/token",
	}
	GoogleScopeUserInfo = "https://www.googleapis.com/auth/userinfo.email"

	googleOauth2RequestURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// Oauth2LoginHandler ...
func Oauth2LoginHandler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := s.Sessions.SessionIDFromRequest(r)
		_, loggedIn := s.Sessions.Store.Get(sid, "user")
		if loggedIn {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		} else {
			s.OAuth2Cfg.RedirectURL = "https://" + r.Host + "/login/oauth2/callback"
			uri := s.OAuth2Cfg.AuthCodeURL("12312")
			http.Redirect(w, r, uri, http.StatusTemporaryRedirect)
		}
	}
}

// Oauth2LoginCallbackHandler ...
func Oauth2LoginCallbackHandler(s *server.Server, callbackURL, redirectURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != "12312" {
			http.Error(w, http.ErrBodyReadAfterClose.Error(), http.StatusInternalServerError)
			return
		}
		code := r.URL.Query().Get("code")
		s.OAuth2Cfg.RedirectURL = "https://" + r.Host + callbackURL
		token, err := s.OAuth2Cfg.Exchange(context.Background(), code)
		if err != nil {
			logging.Println(logging.Info, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest(http.MethodGet, googleOauth2RequestURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header.Add("Authorization", "OAuth "+token.AccessToken)
		hc := http.Client{}
		res, err := hc.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userInfo := server.GoogleUserInfo{}
		err = json.Unmarshal(bytes, &userInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userInfo.Username = strings.Split(userInfo.Email, "@")[0]
		logging.Println(logging.Debug, userInfo)

		user, created, err := s.Services.UserService.CreateUserIfNotExists(userInfo.Email, userInfo.Picture)
		logging.Println(logging.Debug, fmt.Sprintf("User %v was created %t (err: %v)", user, created, err))

		sid := s.Sessions.SessionIDFromRequest(r)
		userInfoJSON, err := json.Marshal(userInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.Sessions.Store.Put(sid, "user", string(userInfoJSON))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}
