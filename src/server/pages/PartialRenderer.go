package pages

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/util/logging"
)

// Cards is a partial renderer for the cards view
func Cards(s *server.Server, taskLastRunFunc TaskLastRunFunc) http.HandlerFunc {
	var init sync.Once
	var tpl *template.Template
	var tplErr error
	return func(rw http.ResponseWriter, r *http.Request) {
		// Caching via (resp) Last-Modified --> (req)If-Modified-Since --> conditional resp
		timeLayout := "Mon, 02 Jan 2006 15:04:05 GMT"
		accountID := r.URL.Query().Get("id")
		accountKind := r.URL.Query().Get("kind")
		lastModified := taskLastRunFunc(accountKind)
		ifModifiedSinceStr := r.Header.Get("If-Modified-Since")
		if ifModifiedSinceStr != "" {
			ifModifiedSince, err := time.Parse(timeLayout, ifModifiedSinceStr)
			if err == nil {
				ifModifiedSince = ifModifiedSince.Add(1 * time.Second)
				if !lastModified.After(ifModifiedSince) {
					rw.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}

		init.Do(func() {
			funcMap := template.FuncMap{
				"fdate": formatDate,
			}
			tpl, tplErr = template.New("cardview.html").Funcs(funcMap).ParseFiles(templates("cardview.html")...)
			if tplErr == nil {
				tpl, tplErr = tpl.Parse(`{{template "cards" .}}`)
			}
		})
		if tplErr != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Error, tplErr)
			return
		}

		sid := s.Sessions.SessionIDFromRequest(r)
		user := server.GoogleUserInfoFromSession(s, sid)

		u, err := s.Services.UserService.GetUser(user.Email)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}
		// paging
		pageStr := r.URL.Query().Get("page")
		page, err := strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			logging.Println(logging.Warn, err)
			page = 0
		}
		countStr := r.URL.Query().Get("count")
		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			logging.Println(logging.Warn, err)
			count = 30
		}
		if count < 0 || count > 30 {
			count = 30
		}
		page *= count
		// content retrieval
		var contents []models.Content
		if accountID == "*" {
			err = s.Services.UserService.LoadUserAccountsForSocialMedia(&u, accountKind)
			if err != nil {
				http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				logging.Println(logging.Error, err)
				return
			}
			contents, err = s.Services.ContentService.LoadContentFor(u.ID, accountKind, -1, page, count)
			if err != nil {
				http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				logging.Println(logging.Error, err)
				return
			}
		} else {
			accID, err := strconv.ParseInt(accountID, 10, 64)
			if err != nil {
				http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				logging.Println(logging.Error, err)
				return
			}
			contents, err = s.Services.ContentService.LoadContentFor(u.ID, accountKind, accID, page, count)
			if err != nil {
				http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				logging.Println(logging.Error, err)
				return
			}
		}
		// content adjustment
		loc := time.Now().Location()
		for idx := range contents {
			contents[idx].ExternalID = ContentURL(contents[idx].ExternalID, accountKind)
			contents[idx].Date = contents[idx].Date.In(loc)
		}

		// rendering
		var buf bytes.Buffer
		err = tpl.Execute(io.Writer(&buf), map[string]interface{}{
			"contents": contents,
		})
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}

		// response
		rw.Header().Set("Last-Modified", lastModified.Format(timeLayout))
		rw.Header().Set("Cache-Control", "private")

		rw.Write(buf.Bytes())
	}
}

// SettingAccountSelection is a partial renderer for setting account selection
func SettingAccountSelection(s *server.Server) http.HandlerFunc {
	var init sync.Once
	var tpl *template.Template
	var tplErr error
	return func(rw http.ResponseWriter, r *http.Request) {
		init.Do(func() {
			tpl, tplErr = tpl.ParseFiles(templates("generic-settings.html")...)
			if tplErr == nil {
				tpl, tplErr = tpl.Parse(`{{template "account-select" .}}`)
			}
		})
		if tplErr != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Error, tplErr)
			return
		}

		sid := s.Sessions.SessionIDFromRequest(r)
		user := server.GoogleUserInfoFromSession(s, sid)

		var accountData struct {
			ID   int64
			Kind string
		}
		err := json.NewDecoder(r.Body).Decode(&accountData)

		u, err := s.Services.UserService.GetUser(user.Email)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			logging.Println(logging.Error, err)
			return
		}
		err = s.Services.UserService.LoadUserAccountsForSocialMedia(&u, accountData.Kind)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			logging.Println(logging.Error, err)
			return
		}

		var buf bytes.Buffer
		err = tpl.Execute(io.Writer(&buf), map[string]interface{}{
			"selection": accountData.ID,
			"accounts":  u.Accounts,
		})
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}

		rw.Write(buf.Bytes())
	}
}

// SettingChannelTable is a partial renderer for setting channel table
func SettingChannelTable(s *server.Server) http.HandlerFunc {
	var init sync.Once
	var tpl *template.Template
	var tplErr error
	return func(rw http.ResponseWriter, r *http.Request) {
		init.Do(func() {
			tpl, tplErr = tpl.ParseFiles(templates("generic-settings.html")...)
			if tplErr == nil {
				tpl, tplErr = tpl.Parse(`{{template "channel-table" .}}`)
			}
		})
		if tplErr != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Error, tplErr)
			return
		}

		var channelData struct {
			AccountID string
			Kind      string
		}
		err := json.NewDecoder(r.Body).Decode(&channelData)
		var accountID int64 = 0
		if channelData.AccountID != "" {
			accountID, err = strconv.ParseInt(channelData.AccountID, 10, 64)
			if err != nil {
				http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				logging.Println(logging.Error, err)
				return
			}
		}

		var channels []models.Channel
		if accountID > 0 {
			channels, err = s.Services.ChannelService.FindChannelsByAccountIDAndKind(accountID, channelData.Kind)
			if err != nil {
				http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				logging.Println(logging.Error, err)
				return
			}
		}

		var buf bytes.Buffer
		err = tpl.Execute(io.Writer(&buf), map[string]interface{}{
			"channels":     channels,
			"channelCount": len(channels),
		})
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}

		rw.Write(buf.Bytes())
	}
}
