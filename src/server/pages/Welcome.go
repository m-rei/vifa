package pages

import (
	"html/template"
	"net/http"

	"visual-feed-aggregator/src/server"
)

// Welcome ...
func Welcome(s *server.Server) http.HandlerFunc {
	return RenderPage(s, func() ([]string, template.FuncMap, RenderPageLogic) {
		pages := []string{"main-layout.html", "sidebar.html", "welcome.html"}
		renderLogic := func(r *http.Request, s *server.Server, sid string, user *server.GoogleUserInfo) (map[string]interface{}, error) {
			return map[string]interface{}{
					"title": s.Env["TITLE"],
					"css":   []string{"components.css", "main-layout.css", "sidebar.css", "welcome.css"},
					"user":  user,
					"media": socialMediaSvgData(""),
				},
				nil
		}
		return pages, nil, renderLogic
	})
}
