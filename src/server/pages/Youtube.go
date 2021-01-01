package pages

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/server/rest"
	"visual-feed-aggregator/src/util/logging"

	"github.com/tdewolff/parse/strconv"
)

// Youtube ...
func Youtube(s *server.Server, taskLastRunFunc TaskLastRunFunc) http.HandlerFunc {
	return RenderPage(s,
		func() ([]string, template.FuncMap, RenderPageLogic) {
			pages := []string{"main-layout.html", "sidebar.html", "youtube.html", "cardview.html"}
			funcMap := template.FuncMap{
				"fdate": formatDate,
			}
			renderLogic := func(r *http.Request, s *server.Server, sid string, user *server.GoogleUserInfo) (map[string]interface{}, error) {
				kind := models.KindYoutube

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
						"css":      []string{"components.css", "main-layout.css", "sidebar.css", "youtube.css", "cardview.css", "carousel.css"},
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

// YoutubeSettings ...
func YoutubeSettings(s *server.Server) http.HandlerFunc {
	return RenderPage(s,
		func() ([]string, template.FuncMap, RenderPageLogic) {
			pages := []string{"main-layout.html", "sidebar.html", "settings.html", "generic-settings.html"}
			renderLogic := func(r *http.Request, s *server.Server, sid string, user *server.GoogleUserInfo) (map[string]interface{}, error) {
				kind := models.KindYoutube

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
						"js":           []string{"youtube-settings.js", "settings.js"},
						"opml":         true,
						"user":         user,
						"accounts":     u.Accounts,
						"channelHint":  "examples: \n  https://www.youtube.com/channel/UC_aEa8K-EOJ3D6gOs7HcyNg \n  https://www.youtube.com/user/aaarguments",
						"channels":     channels,
						"channelCount": len(channels),
						"media":        socialMediaSvgData(kind),
					},
					nil
			}
			return pages, nil, renderLogic
		})
}

// YoutubeOpmlUpload ...
func YoutubeOpmlUpload(s *server.Server) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(1024 * 10)
		accountIDStr := r.FormValue("accountID")
		accountID, _ := strconv.ParseInt([]byte(accountIDStr))
		file, _, err := r.FormFile("opml-file")
		if err != nil {
			json.NewEncoder(rw).Encode(err)
			logging.Println(logging.Error, err)
			return
		}

		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			json.NewEncoder(rw).Encode(err)
			logging.Println(logging.Error, err)
			return
		}

		type Outline struct {
			XMLName  xml.Name  `xml:"outline"`
			Outlines []Outline `xml:"outline"`
			Title    string    `xml:"title,attr"`
			URL      string    `xml:"xmlUrl,attr"`
		}
		type Body struct {
			XMLName xml.Name `xml:"body"`
			Outline Outline  `xml:"outline"`
		}
		type Opml struct {
			XMLName xml.Name `xml:"opml"`
			Body    Body     `xml:"body"`
		}

		xmlData := []byte(strings.ToValidUTF8(string(fileBytes), ""))

		var opml Opml
		decoder := xml.NewDecoder(strings.NewReader(string(xmlData)))
		decoder.Strict = false
		err = decoder.Decode(&opml)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			logging.Println(logging.Error, err)
			return
		}
		for _, o := range opml.Body.Outline.Outlines {
			rest.DoAddChannel(s, accountID, o.URL, YoutubeMetaDataProvider)
		}
		rw.WriteHeader(http.StatusOK)
	}
}

var youtubeRegEx1 = regexp.MustCompile(`youtube\.com\/(user\/[^\/\n]*)`)
var youtubeRegEx2 = regexp.MustCompile(`(user=[^\/\n]*)`)
var youtubeRegEx3 = regexp.MustCompile(`youtube\.com\/(channel\/[^\/\n]*)`)
var youtubeRegEx4 = regexp.MustCompile(`(channel_id=[^\/\n]*)`)

func extractYoutubeExternalID(data string) string {
	ret := ""
	qry := ""
	res := youtubeRegEx1.FindAllStringSubmatch(data, -1)
	if len(res) > 0 {
		ret = res[0][1]
		qry = strings.Split(ret, "/")[1]
	}
	if ret == "" {
		res = youtubeRegEx2.FindAllStringSubmatch(data, -1)
		if len(res) > 0 {
			ret = res[0][1]
			qry = strings.Split(ret, "=")[1]
		}
	}
	if ret != "" {
		if err := httpCanGet("HEAD", "https://youtube.com/user/"+qry); err != nil {
			logging.Println(logging.Info, err)
			return ""
		}
		return ret
	}

	if ret == "" {
		res = youtubeRegEx3.FindAllStringSubmatch(data, -1)
		if len(res) > 0 {
			ret = res[0][1]
			qry = strings.Split(ret, "/")[1]
		}
	}
	if ret == "" {
		res = youtubeRegEx4.FindAllStringSubmatch(data, -1)
		if len(res) > 0 {
			ret = res[0][1]
			qry = strings.Split(ret, "=")[1]
		}
	}
	if ret != "" {
		if err := httpCanGet("HEAD", "https://youtube.com/channel/"+qry); err != nil {
			logging.Println(logging.Info, err)
			return ""
		}
		return ret
	}

	return ret
}

// YoutubeValidateChannel checks if the channel id is a valid one
func YoutubeValidateChannel(s *server.Server) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			logging.Println(logging.Info, err)
			return
		}

		if extractYoutubeExternalID(string(bodyBytes)) != "" {
			rw.WriteHeader(http.StatusOK)
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
	}
}

// YoutubeMetaDataProvider ...
func YoutubeMetaDataProvider(channelID string) (string, string, string, string) {
	externalID := extractYoutubeExternalID(channelID)
	if externalID == "" {
		return "", "", "", ""
	}
	externalIDParam := strings.Replace(externalID, "/", "=", 1)
	if !strings.HasPrefix(externalIDParam, "channel_id") {
		externalIDParam = strings.Replace(externalIDParam, "channel", "channel_id", 1)
	}
	author := queryYoutubeChannelAuthor("https://youtube.com/feeds/videos.xml?" + externalIDParam)

	return author, models.KindYoutube, "", externalIDParam
}

func queryYoutubeChannelAuthor(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		logging.Println(logging.Info, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		logging.Println(logging.Info, err)
		return ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Println(logging.Info, err)
		return ""
	}

	type author struct {
		XMLName xml.Name `xml:"author"`
		Name    string   `xml:"name"`
	}
	type feed struct {
		XMLName xml.Name `xml:"feed"`
		Author  author   `xml:"author"`
	}
	data := &feed{}
	err = xml.Unmarshal(body, data)
	if err != nil {
		logging.Println(logging.Info, err)
		return ""
	}
	return data.Author.Name
}

// YoutubeChannelValidator validates channel data for youtube
func YoutubeChannelValidator(data string) bool {
	return extractYoutubeExternalID(data) != ""
}
