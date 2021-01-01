package pages

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/util"
	"visual-feed-aggregator/src/util/logging"
)

// Instagram ...
func Instagram(s *server.Server, taskLastRunFunc TaskLastRunFunc) http.HandlerFunc {
	return RenderPage(s,
		func() ([]string, template.FuncMap, RenderPageLogic) {
			pages := []string{"main-layout.html", "sidebar.html", "instagram.html", "cardview.html"}
			funcMap := template.FuncMap{
				"fdate": formatDate,
			}
			renderLogic := func(r *http.Request, s *server.Server, sid string, user *server.GoogleUserInfo) (map[string]interface{}, error) {
				kind := models.KindInstagram

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
					"css":      []string{"components.css", "main-layout.css", "sidebar.css", "instagram.css", "cardview.css", "carousel.css"},
					"js":       []string{"cardview-header.js", "carousel.js"},
					"snapshot": snapshotTime,
					"user":     user,
					"accounts": u.Accounts,
					"kind":     kind,
					"media":    socialMediaSvgData(kind),
				}, nil
			}
			return pages, funcMap, renderLogic
		})
}

// InstagramSettings ...
func InstagramSettings(s *server.Server) http.HandlerFunc {
	return RenderPage(s,
		func() ([]string, template.FuncMap, RenderPageLogic) {
			pages := []string{"main-layout.html", "sidebar.html", "settings.html", "generic-settings.html"}
			renderLogic := func(r *http.Request, s *server.Server, sid string, user *server.GoogleUserInfo) (map[string]interface{}, error) {
				kind := models.KindInstagram

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
						logging.Println(logging.Error, err)
					}
				}
				csrfToken, _ := s.Sessions.Store.Get(sid, "CsrfToken")
				return map[string]interface{}{
						"title":        s.Env["TITLE"],
						"csrf":         csrfToken,
						"css":          []string{"components.css", "main-layout.css", "sidebar.css", "settings.css"},
						"js":           []string{"instagram-settings.js", "settings.js"},
						"user":         user,
						"accounts":     u.Accounts,
						"channelHint":  "example \n  https://www.instagram.com/fundotcom_/?hl=en",
						"channels":     channels,
						"channelCount": len(channels),
						"media":        socialMediaSvgData(kind),
					},
					nil
			}
			return pages, nil, renderLogic
		})
}

// InstagramMetaDataProvider ...
func InstagramMetaDataProvider(channelID string) (string, string, string, string) {
	externalID := extractInstagramExternalID(channelID)
	if externalID == "" {
		return "", "", "", ""
	}
	profilePic := queryInstagramProfilePic(externalID)

	return externalID, models.KindInstagram, profilePic, externalID
}

func queryInstagramProfilePic(externalID string) string {
	url := "https://instagram.com/" + externalID + "/?__a=1"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logging.Println(logging.Info, err)
		return ""
	}
	req.Header.Set("User-Agent", util.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		logging.Println(logging.Info, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		logging.Println(logging.Info, resp.StatusCode, resp.Status)
		return ""
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Println(logging.Info, err)
		return ""
	}

	type user struct {
		ProfilePicURLHD string `json:"profile_pic_url_hd"`
	}
	type graphql struct {
		User user `json:"user"`
	}
	type instagramJSON struct {
		GraphQL graphql `json:"graphql"`
	}
	var instaJSON instagramJSON
	err = json.Unmarshal(data, &instaJSON)
	if err != nil {
		logging.Println(logging.Info, err)
		return ""
	}

	return instaJSON.GraphQL.User.ProfilePicURLHD
}

var instaRegEx = regexp.MustCompile(`instagram\.com\/([^\/]+)`)

func extractInstagramExternalID(data string) string {
	res := instaRegEx.FindAllStringSubmatch(data, -1)
	if len(res) > 0 {
		ret := res[0][1]
		if err := httpCanGet("HEAD", "https://instagram.com/"+ret+"/?__a=1"); err != nil {
			logging.Println(logging.Info, err)
			return ""
		}
		return ret
	}
	return ""
}

// InstagramChannelValidator validates channel data for instagram
func InstagramChannelValidator(data string) bool {
	return extractInstagramExternalID(data) != ""
}
