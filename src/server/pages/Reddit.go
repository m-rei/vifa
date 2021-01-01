package pages

import (
	"html/template"
	"net/http"
	"regexp"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/util/logging"
)

// Reddit ...
func Reddit(s *server.Server, taskLastRunFunc TaskLastRunFunc) http.HandlerFunc {
	return RenderPage(s,
		func() ([]string, template.FuncMap, RenderPageLogic) {
			pages := []string{"main-layout.html", "sidebar.html", "reddit.html", "cardview.html"}
			funcMap := template.FuncMap{
				"fdate": formatDate,
			}
			renderLogic := func(r *http.Request, s *server.Server, sid string, user *server.GoogleUserInfo) (map[string]interface{}, error) {
				kind := models.KindReddit

				u, err := s.Services.UserService.GetUser(user.Email)
				if err != nil {
					return nil, err
				}
				err = s.Services.UserService.LoadUserAccountsForSocialMedia(&u, kind)
				if err != nil {
					return nil, err
				}

				snapshotTime := taskLastRunFunc(kind).Format("2006-01-02 15:04:05 MST")
				csrfToken, _ := s.Sessions.Store.Get(sid, "CsrfToken")
				return map[string]interface{}{
						"title":    s.Env["TITLE"],
						"csrf":     csrfToken,
						"css":      []string{"components.css", "main-layout.css", "sidebar.css", "reddit.css", "cardview.css", "carousel.css"},
						"js":       []string{"cardview-header.js", "carousel.js"},
						"snapshot": snapshotTime,
						"user":     user,
						"accounts": u.Accounts,
						"kind":     kind,
						"media":    socialMediaSvgData(kind),
					},
					nil
			}
			return pages, funcMap, renderLogic
		})
}

// RedditSettings ...
func RedditSettings(s *server.Server) http.HandlerFunc {
	return RenderPage(s,
		func() ([]string, template.FuncMap, RenderPageLogic) {
			pages := []string{"main-layout.html", "sidebar.html", "settings.html", "generic-settings.html"}
			renderLogic := func(r *http.Request, s *server.Server, sid string, user *server.GoogleUserInfo) (map[string]interface{}, error) {
				kind := models.KindReddit

				u, err := s.Services.UserService.GetUser(user.Email)
				if err != nil {
					return nil, err
				}
				err = s.Services.UserService.LoadUserAccountsForSocialMedia(&u, kind)
				if err != nil {
					return nil, err
				}
				channels := []models.Channel{}
				if len(u.Accounts) > 0 {
					channels, err = s.Services.ChannelService.FindChannelsByAccountIDAndKind(u.Accounts[0].ID, kind)
					if err != nil {
						logging.Println(logging.Info, err)
					}
				}
				csrfToken, _ := s.Sessions.Store.Get(sid, "CsrfToken")
				return map[string]interface{}{
						"title":        s.Env["TITLE"],
						"csrf":         csrfToken,
						"css":          []string{"components.css", "main-layout.css", "sidebar.css", "settings.css"},
						"js":           []string{"reddit-settings.js", "settings.js"},
						"user":         user,
						"accounts":     u.Accounts,
						"channelHint":  "example \n  https://www.reddit.com/r/funnygifs/",
						"channels":     channels,
						"channelCount": len(channels),
						"media":        socialMediaSvgData(kind),
					},
					nil
			}
			return pages, nil, renderLogic
		})
}

// RedditMetaDataProvider ...
func RedditMetaDataProvider(channelID string) (string, string, string, string) {
	externalID := extractRedditExternalID(channelID)
	if externalID == "" {
		return "", "", "", ""
	}
	author := externalID

	return author, models.KindReddit, "", externalID
}

var redditRegEx = regexp.MustCompile(`reddit\.com\/r\/([^\/]+)`)

func extractRedditExternalID(data string) string {
	res := redditRegEx.FindAllStringSubmatch(data, -1)
	if len(res) > 0 {
		ret := res[0][1]
		if err := httpCanGet("HEAD", "https://reddit.com/r/"+ret); err != nil {
			logging.Println(logging.Info, err)
			return ""
		}
		return ret
	}
	return ""
}

// RedditChannelValidator validates channel data for reddit
func RedditChannelValidator(data string) bool {
	return extractRedditExternalID(data) != ""
}
