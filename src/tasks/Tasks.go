package tasks

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/database/services"
	"visual-feed-aggregator/src/util"
	"visual-feed-aggregator/src/util/logging"

	"github.com/jmoiron/sqlx"
)

// CleanupBackgroundTask removes
func CleanupBackgroundTask(stopSignal <-chan bool, lastRun map[string]time.Time, db *sqlx.DB, services *services.ServiceCollection, cutoffDays, refreshRateMinutes int64) {
	runTask(stopSignal, lastRun, db, services, cutoffDays, refreshRateMinutes, "cleanup", cleanupTask)
}

func cleanupTask(dateCutoff *time.Time, loc *time.Location, services *services.ServiceCollection) {
	cutoff := dateCutoff.AddDate(0, 0, -1)
	logging.Println(logging.Info, "Removing everything older than", cutoff)
	amount, err := services.ContentService.CleanupOldContent(&cutoff)
	if err != nil {
		logging.Println(logging.Info, err)
	} else {
		logging.Println(logging.Info, fmt.Sprintf("Cleansed %d records from content", amount))
	}
	amount, err = services.ChannelService.CleanupOrphanedChannels()
	if err != nil {
		logging.Println(logging.Info, err)
	} else {
		logging.Println(logging.Info, fmt.Sprintf("Cleansed %d orphaned channels", amount))
	}
}

// YoutubeBackgroundTask ...
func YoutubeBackgroundTask(stopSignal <-chan bool, lastRun map[string]time.Time, db *sqlx.DB, services *services.ServiceCollection, cutoffDays, refreshRateMinutes int64) {
	runChannelTask(stopSignal, lastRun, db, services, cutoffDays, refreshRateMinutes, models.KindYoutube, youtubeTask)
}

func youtubeTask(channel *models.Channel, wg *sync.WaitGroup, dateCutoff *time.Time, loc *time.Location, services *services.ServiceCollection) {
	defer wg.Done()

	resp, err := http.Get("https://youtube.com/feeds/videos.xml?" + channel.ExternalID)
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: get() error!")
		logging.Println(logging.Error, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error:", resp.Status)
		logging.Println(logging.Error, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: readall error!")
		logging.Println(logging.Error, err)
		return
	}

	type entry struct {
		XMLName   xml.Name `xml:"entry"`
		Title     string   `xml:"title"`
		VideoID   string   `xml:"videoId"`
		Published string   `xml:"published"`
	}
	type feed struct {
		XMLName xml.Name `xml:"feed"`
		Entries []entry  `xml:"entry"`
	}
	var f feed
	err = xml.Unmarshal(body, &f)
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: unmarshal error!")
		logging.Println(logging.Error, err)
		return
	}

	for idx, item := range f.Entries {
		if idx >= 50 {
			break
		}
		var content models.Content
		content.ChannelID = channel.ID
		content.Title = item.Title
		date, err := parseYoutubeTimeStr(item.Published, loc)
		if err != nil {
			logging.Println(logging.Info, err)
			break
		} else {
			content.Date = date
		}
		if content.Date.Before(*dateCutoff) {
			break
		}
		content.ExternalID = item.VideoID

		err = services.ContentService.CreateContent(&content)
		if err != nil { // content already exists (most likely)
			continue
		}

		var media models.Media
		media.ContentID = content.ID
		media.URL = "https://img.youtube.com/vi/" + item.VideoID + "/sddefault.jpg" // maxresdefault

		if err := services.MediaService.CreateMedia(&media); err != nil {
			logging.Println(logging.Error, err)
		}
	}
}

func parseYoutubeTimeStr(timestampStr string, loc *time.Location) (time.Time, error) {
	datetime, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return time.Time{}, err
	}
	return datetime.In(loc), nil
}

// RedditBackgroundTask ...
func RedditBackgroundTask(stopSignal <-chan bool, lastRun map[string]time.Time, db *sqlx.DB, services *services.ServiceCollection, cutoffDays, refreshRateMinutes int64) {
	runChannelTask(stopSignal, lastRun, db, services, cutoffDays, refreshRateMinutes, models.KindReddit, redditTask)
}

func redditTask(channel *models.Channel, wg *sync.WaitGroup, dateCutoff *time.Time, loc *time.Location, services *services.ServiceCollection) {
	defer wg.Done()

	resp, err := util.HTTPRequest("GET", "https://reddit.com/r/"+channel.ExternalID+"/new/.json")
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: get() error!")
		logging.Println(logging.Error, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error:", resp.Status)
		logging.Println(logging.Error, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: readall error!")
		logging.Println(logging.Error, err)
		return
	}

	type galleryItem struct {
		MediaID string `json:"media_id"`
	}
	type gallery struct {
		Items []galleryItem
	}
	type data3 struct {
		Author        string
		Title         string
		Thumbnail     string
		Permalink     string
		URL           string
		CreatedUTC    float64                `json:"created_utc"`
		PostHint      string                 `json:"post_hint"`
		IsGallery     bool                   `json:"is_gallery"`
		GalleryData   gallery                `json:"gallery_data,omitempty"`
		MediaMetaData map[string]interface{} `json:"media_metadata"`
	}

	type data2 struct {
		Data data3
	}
	type data struct {
		Children []data2
	}
	type feed struct {
		Data data
	}

	var f feed
	err = json.Unmarshal(body, &f)
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: unmarshal error!")
		logging.Println(logging.Error, err)
		return
	}

	for _, item := range f.Data.Children {
		var content models.Content
		content.ChannelID = channel.ID
		content.Date = time.Unix(int64(item.Data.CreatedUTC), 0).UTC().In(loc)
		content.ExternalID = item.Data.Permalink
		content.Title = item.Data.Title

		if content.Date.Before(*dateCutoff) {
			continue
		}

		err = services.ContentService.CreateContent(&content)
		if err != nil {
			continue
		}

		if item.Data.PostHint == "image" {
			var media models.Media
			media.ContentID = content.ID
			media.URL = item.Data.URL
			if err := services.MediaService.CreateMedia(&media); err != nil {
				logging.Println(logging.Error, err)
			}
		} else if item.Data.IsGallery {
			for _, g := range item.Data.GalleryData.Items {
				urlEscaped := item.Data.MediaMetaData[g.MediaID].(map[string]interface{})["s"].(map[string]interface{})["u"].(string)

				var media models.Media
				media.ContentID = content.ID
				media.URL = html.UnescapeString(urlEscaped)
				if err := services.MediaService.CreateMedia(&media); err != nil {
					logging.Println(logging.Error, err)
				}
			}
		} else if hasImageExtension(item.Data.URL) {
			var media models.Media
			media.ContentID = content.ID
			media.URL = item.Data.URL
			if err := services.MediaService.CreateMedia(&media); err != nil {
				logging.Println(logging.Error, err)
			}
		} else if strings.HasPrefix(item.Data.Thumbnail, "http") {
			var media models.Media
			media.ContentID = content.ID
			media.URL = item.Data.Thumbnail
			if err := services.MediaService.CreateMedia(&media); err != nil {
				logging.Println(logging.Error, err)
			}
		}
	}
}

func hasImageExtension(url string) bool {
	switch filepath.Ext(url) {
	case ".png", ".jpeg", ".jpg", ".gif":
		return true
	}
	return false
}

// TwitterBackgroundTask ...
func TwitterBackgroundTask(stopSignal <-chan bool, lastRun map[string]time.Time, db *sqlx.DB, services *services.ServiceCollection, cutoffDays, refreshRateMinutes int64) {
	runChannelTask(stopSignal, lastRun, db, services, cutoffDays, refreshRateMinutes, models.KindTwitter, twitterTask)
}

func twitterTask(channel *models.Channel, wg *sync.WaitGroup, dateCutoff *time.Time, loc *time.Location, services *services.ServiceCollection) {
	defer wg.Done()

	resp, err := http.Get("https://nitter.net/" + channel.ExternalID + "/media/rss")
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: get() error!")
		logging.Println(logging.Error, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error:", resp.Status)
		logging.Println(logging.Error, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: readall error!")
		logging.Println(logging.Error, err)
		return
	}

	type item struct {
		XMLName     xml.Name `xml:"item"`
		Title       string   `xml:"title"`
		Creator     string   `xml:"creator"`
		Description string   `xml:"description"`
		PubDate     string   `xml:"pubDate"`
		Link        string   `xml:"link"`
	}
	type chnl struct {
		XMLName xml.Name `xml:"channel"`
		Items   []item   `xml:"item"`
	}
	type feed struct {
		XMLName xml.Name `xml:"rss"`
		Channel chnl     `xml:"channel"`
	}
	var f feed
	err = xml.Unmarshal(body, &f)
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: unmarshal error!")
		logging.Println(logging.Error, err)
		return
	}
	pattern := regexp.MustCompile(`(img src|video poster)="([^"]*)"`)
	for _, item := range f.Channel.Items { // 20 elements per feed
		var content models.Content
		content.ChannelID = channel.ID
		content.Title = item.Creator + "-" + item.Title
		content.ExternalID = strings.Replace(item.Link, "https://nitter.net/", "", 1)
		date, err := parseTwitterTimeStr(item.PubDate, loc)
		if err != nil {
			logging.Println(logging.Info, err)
		} else {
			content.Date = date
		}

		if content.Date.Before(*dateCutoff) {
			continue
		}

		err = services.ContentService.CreateContent(&content)
		if err != nil { // content already exists
			continue
		}

		// description contains links to media
		res := pattern.FindAllStringSubmatch(item.Description, -1)
		for i := 0; i < len(res); i++ {
			var media models.Media
			media.ContentID = content.ID
			media.URL = res[i][2]

			if err := services.MediaService.CreateMedia(&media); err != nil {
				logging.Println(logging.Error, err)
			}
		}
	}
}

func parseTwitterTimeStr(timestampStr string, loc *time.Location) (time.Time, error) {
	datetime, err := time.Parse("Mon, _2 Jan 2006 15:04:05 MST", timestampStr)
	if err != nil {
		return time.Time{}, err
	}
	return datetime.In(loc), nil
}

// InstagramBackgroundTask ...
func InstagramBackgroundTask(stopSignal <-chan bool, lastRun map[string]time.Time, db *sqlx.DB, services *services.ServiceCollection, cutoffDays, refreshRateMinutes int64) {
	runChannelTask(stopSignal, lastRun, db, services, cutoffDays, refreshRateMinutes, models.KindInstagram, instagramTask)
}

func instagramTask(channel *models.Channel, wg *sync.WaitGroup, dateCutoff *time.Time, loc *time.Location, services *services.ServiceCollection) {
	defer wg.Done()

	resp, err := util.HTTPRequest("GET", "https://www.instagram.com/"+channel.ExternalID+"/?__a=1")
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: get() error!")
		logging.Println(logging.Error, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error:", resp.Status)
		logging.Println(logging.Error, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: readall error!")
		logging.Println(logging.Error, err)
		return
	}

	type captionNodeData struct {
		Text string
	}
	type captionNode struct {
		Node captionNodeData
	}
	type edgeMediaToCaption struct {
		Edges []captionNode
	}
	type sidecarNodeData struct {
		DisplayURL string `json:"display_url"`
	}
	type sidecarNode struct {
		Node sidecarNodeData
	}
	type edgeSidecarToChildren struct {
		Edges []sidecarNode
	}
	type nodeData struct {
		Shortcode             string
		Typename              string                `json:"__typename"`            // GraphSidecar (multiple media), GraphImage (media type #1), GraphVideo (media type #2)
		DisplayURL            string                `json:"display_url"`           // for graph image
		EdgeMediaToCaption    edgeMediaToCaption    `json:"edge_media_to_caption"` // should have always a minimum of 1?!
		EdgeSidecarToChildren edgeSidecarToChildren `json:"edge_sidecar_to_children"`
		TakenAtTimestamp      float64               `json:"taken_at_timestamp"`
	}
	type node struct {
		Node nodeData
	}
	type edgeOwnerToTimelineMedia struct {
		Edges []node
	}
	type user struct {
		EdgeOwnerToTimelineMedia edgeOwnerToTimelineMedia `json:"edge_owner_to_timeline_media"`
	}
	type graphql struct {
		User user
	}
	type feed struct {
		Graphql graphql
	}

	var f feed
	err = json.Unmarshal(body, &f)
	if err != nil {
		logging.Println(logging.Error, "Channel:", channel.Name, "--Error: unmarshal error!")
		logging.Println(logging.Error, err)
		return
	}
	for _, edge := range f.Graphql.User.EdgeOwnerToTimelineMedia.Edges {
		var content models.Content
		content.ChannelID = channel.ID
		content.Date = time.Unix(int64(edge.Node.TakenAtTimestamp), 0).UTC().In(loc)
		content.ExternalID = edge.Node.Shortcode // https://instagram.com/p/ + shortcode
		if content.ExternalID == "" {
			continue
		}
		content.Title = ""
		for _, capEdge := range edge.Node.EdgeMediaToCaption.Edges {
			content.Title += capEdge.Node.Text
		}

		if content.Date.Before(*dateCutoff) {
			break
		}

		err = services.ContentService.CreateContent(&content)
		if err != nil {
			continue
		}

		switch edge.Node.Typename {
		case "GraphVideo", "GraphImage":
			var media models.Media
			media.ContentID = content.ID
			media.URL = edge.Node.DisplayURL
			if err := services.MediaService.CreateMedia(&media); err != nil {
				logging.Println(logging.Error, err)
			}
		case "GraphSidecar":
			for _, sidecar := range edge.Node.EdgeSidecarToChildren.Edges {
				var media models.Media
				media.ContentID = content.ID
				media.URL = sidecar.Node.DisplayURL
				if err := services.MediaService.CreateMedia(&media); err != nil {
					logging.Println(logging.Error, err)
				}
			}
		}
	}
}
