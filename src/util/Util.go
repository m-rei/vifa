package util

import (
	"net/http"
	"time"
)

// NitterInstances are all the public nitter instances which can be found at:
// https://github.com/zedeus/nitter/wiki/Instances
var NitterInstances = []string{
	"nitter.net",
	"nitter.42l.fr",
	"nitter.nixnet.services",
	"nitter.13ad.de",
	"nitter.pussthecat.org",
	"nitter.mastodont.cat",
	"nitter.dark.fail",
	"nitter.tedomum.net",
	"nitter.cattube.org",
	"nitter.fdn.fr",
	"nitter.1d4.us",
	"nitter.kavin.rocks",
	"tweet.lambda.dance",
	"nitter.cc",
	"nitter.weaponizedhumiliation.com",
	"nitter.vxempire.xyz",
	"nitter.unixfox.eu",
	"nitter.domain.glass",
	"nitter.himiko.cloud",
	"nitter.eu",
	"nitter.ethibox.fr",
	"nitter.namazso.eu",
}

// UserAgent is the default user agent to circumvent bot rejection
const UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:84.0) Gecko/20100101 Firefox/84.0"

// HTTPRequest ...
func HTTPRequest(method, url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	client.Timeout = 1 * time.Minute
	return client.Do(req)
}
