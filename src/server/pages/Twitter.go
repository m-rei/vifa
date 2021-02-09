package pages

import (
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/util"
	"visual-feed-aggregator/src/util/logging"
)

// Twitter ...
func Twitter(s *server.Server, taskLastRunFunc TaskLastRunFunc) http.HandlerFunc {
	return RenderPage(s,
		func() ([]string, template.FuncMap, RenderPageLogic) {
			pages := []string{"main-layout.html", "sidebar.html", "twitter.html", "cardview.html"}
			funcMap := template.FuncMap{
				"fdate": formatDate,
			}
			renderLogic := func(r *http.Request, s *server.Server, sid string, user *server.GoogleUserInfo) (map[string]interface{}, error) {
				kind := models.KindTwitter

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
						"css":      []string{"components.css", "main-layout.css", "sidebar.css", "twitter.css", "cardview.css", "carousel.css"},
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

// TwitterSettings ...
func TwitterSettings(s *server.Server) http.HandlerFunc {
	return RenderPage(s,
		func() ([]string, template.FuncMap, RenderPageLogic) {
			pages := []string{"main-layout.html", "sidebar.html", "settings.html", "generic-settings.html"}
			renderLogic := func(r *http.Request, s *server.Server, sid string, user *server.GoogleUserInfo) (map[string]interface{}, error) {
				kind := models.KindTwitter

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
						"js":           []string{"twitter-settings.js", "settings.js"},
						"user":         user,
						"accounts":     u.Accounts,
						"channelHint":  "example \n  https://twitter.com/golang",
						"channels":     channels,
						"channelCount": len(channels),
						"media":        socialMediaSvgData(kind),
					},
					nil
			}
			return pages, nil, renderLogic
		})
}

// TwitterMetaDataProvider ...
func TwitterMetaDataProvider(channelID string) (string, string, string, string) {
	externalID := extractTwitterExternalID(channelID)
	if externalID == "" {
		return "", "", "", ""
	}
	profilePic := queryTwitterProfilePic(externalID)

	return externalID, models.KindTwitter, profilePic, externalID
}

func queryTwitterProfilePic(externalID string) string {
	var resp *http.Response = nil
	var err error
	url := ""
	for _, ni := range util.NitterInstances {
		url = "https://" + ni + "/" + externalID + "/rss"
		resp, err = http.Get(url)
		if resp.StatusCode >= 200 && resp.StatusCode < 400 && err == nil {
			break
		}
	}
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

	type image struct {
		URL string `xml:"url"`
	}
	type channel struct {
		XMLName xml.Name `xml:"channel"`
		Image   image    `xml:"image"`
	}
	type rss struct {
		XMLName xml.Name `xml:"rss"`
		Channel channel  `xml:"channel"`
	}
	var nitterXML rss
	err = xml.Unmarshal(data, &nitterXML)
	if err != nil {
		logging.Println(logging.Info, err)
		return ""
	}

	return nitterXML.Channel.Image.URL
}

var twitterRegEx = regexp.MustCompile(`twitter\.com\/([^\/?]+)`)

func extractTwitterExternalID(data string) string {
	res := twitterRegEx.FindAllStringSubmatch(data, -1)
	if len(res) > 0 {
		ret := res[0][1]
		var err error
		for _, ni := range util.NitterInstances {
			err = httpCanGet("GET", "https://"+ni+"/"+ret)
			if err == nil {
				break
			}
		}
		if err != nil { // head is not supported
			logging.Println(logging.Info, err)
			return ""
		}
		return ret
	}
	return ""
}

// TwitterChannelValidator validates channel data for twitter
func TwitterChannelValidator(data string) bool {
	return extractTwitterExternalID(data) != ""
}
