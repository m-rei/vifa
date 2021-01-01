package pages

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"path"
	"sync"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/util/logging"
)

// RenderPageLogic ...
type RenderPageLogic func(*http.Request, *server.Server, string, *server.GoogleUserInfo) (map[string]interface{}, error)

// RenderPageLogicFactory ...
type RenderPageLogicFactory func() ([]string, template.FuncMap, RenderPageLogic)

// RenderPage is the generic page renderer
func RenderPage(s *server.Server, f RenderPageLogicFactory) http.HandlerFunc {
	var init sync.Once
	var tpl *template.Template
	var tplErr error
	pages, funcMap, businessLogic := f()
	return func(rw http.ResponseWriter, r *http.Request) {
		init.Do(func() {
			if funcMap != nil {
				tpl, tplErr = template.New(path.Base(pages[0])).Funcs(funcMap).ParseFiles(templates(pages...)...)
			} else {
				tpl, tplErr = tpl.ParseFiles(templates(pages...)...)
			}
		})
		if tplErr != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Error, tplErr)
			return
		}

		sid := s.Sessions.SessionIDFromRequest(r)
		user := server.GoogleUserInfoFromSession(s, sid)

		data, err := businessLogic(r, s, sid, user)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}

		var buf bytes.Buffer
		err = tpl.Execute(io.Writer(&buf), data)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}

		rw.Write(buf.Bytes())
	}
}
