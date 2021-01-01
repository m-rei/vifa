package pages

import (
	"html/template"
	"net/http"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/util/logging"
)

type stats struct {
	Accounts int
	Channels int
	Contents int
}

// Profile ...
func Profile(s *server.Server) http.HandlerFunc {
	return RenderPage(s, func() ([]string, template.FuncMap, RenderPageLogic) {
		pages := []string{"main-layout.html", "sidebar.html", "profile.html"}
		renderLogic := func(r *http.Request, s *server.Server, sid string, user *server.GoogleUserInfo) (map[string]interface{}, error) {
			u, err := s.Services.UserService.GetUser(user.Email)
			if err != nil {
				return nil, err
			}
			sts := []stats{
				getStats(s, u, models.KindYoutube),
				getStats(s, u, models.KindReddit),
				getStats(s, u, models.KindTwitter),
			}
			sts = append(sts, stats{
				sts[0].Accounts + sts[1].Accounts + sts[2].Accounts,
				sts[0].Channels + sts[1].Channels + sts[2].Channels,
				sts[0].Contents + sts[1].Contents + sts[2].Contents,
			})
			return map[string]interface{}{
					"title": s.Env["TITLE"],
					"css":   []string{"components.css", "main-layout.css", "sidebar.css", "profile.css"},
					"user":  user,
					"stats": sts,
					"media": socialMediaSvgData(""),
				},
				nil
		}
		return pages, nil, renderLogic
	})
}

func getStats(s *server.Server, u models.User, kind string) stats {
	var ret stats
	err := s.Services.UserService.LoadUserAccountsForSocialMedia(&u, kind)
	if err != nil {
		logging.Println(logging.Info, err)
		return ret
	}
	ret.Accounts = len(u.Accounts)

	for _, acc := range u.Accounts {
		channels, err := s.Services.ChannelService.FindChannelsByAccountIDAndKind(acc.ID, kind)
		if err != nil {
			logging.Println(logging.Info, err)
		} else {
			ret.Channels += len(channels)
		}

		contents, err := s.Services.ContentService.CountAllContentFor(u.ID, kind, acc.ID)
		if err != nil {
			logging.Println(logging.Info, err)
		} else {
			ret.Contents += int(contents)
		}
	}
	return ret
}
