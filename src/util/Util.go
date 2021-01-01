package util

import (
	"net/http"
	"time"
)

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
