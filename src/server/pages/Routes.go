package pages

import (
	"net/http"
	"strings"
	"time"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/server/middleware"
	"visual-feed-aggregator/src/server/rest"
	"visual-feed-aggregator/src/util/logging"

	"github.com/julienschmidt/httprouter"
)

// TaskLastRunFunc retrieves the last-run time, used for caching purposes (last-modified)
type TaskLastRunFunc func(kind string) time.Time

// SetupRoutes sets up all routes.
// This is the central place for all routes!
func SetupRoutes(s *server.Server, router *httprouter.Router, taskLastRunFunc TaskLastRunFunc) {
	middlewares := []func(http.HandlerFunc) http.HandlerFunc{
		s.Sessions.SessionMiddleware,
		middleware.AutoDetectContentType,
		middleware.Minifier{M: s.Minifier}.MinifierMiddleware,
		middleware.GzipMiddleware,
		middleware.Recover,
	}
	middlewaresEx := []func(http.HandlerFunc) http.HandlerFunc{
		s.Sessions.SessionMiddleware,
		s.Sessions.AuthorizedMiddleware,
		middleware.DefaultHeaders,
		middleware.AutoDetectContentType,
		middleware.Minifier{M: s.Minifier}.MinifierMiddleware,
		middleware.GzipMiddleware,
		middleware.Recover,
	}
	middlewaresExCSRF := []func(http.HandlerFunc) http.HandlerFunc{
		s.Sessions.SessionMiddleware,
		s.Sessions.AuthorizedMiddleware,
		s.Sessions.CsrfTokenValidator,
		middleware.DefaultHeaders,
		middleware.AutoDetectContentType,
		middleware.Minifier{M: s.Minifier}.MinifierMiddleware,
		middleware.GzipMiddleware,
		middleware.Recover,
	}

	// routes
	router.HandlerFunc(http.MethodGet, "/", use(Welcome(s), middlewares...))
	router.HandlerFunc(http.MethodGet, "/profile", use(Profile(s), middlewaresEx...))
	router.HandlerFunc(http.MethodGet, "/logout", use(Logout(s), middlewaresEx...))

	router.HandlerFunc(http.MethodGet, "/youtube", use(Youtube(s, taskLastRunFunc), middlewaresEx...))
	router.HandlerFunc(http.MethodGet, "/youtube-settings", use(YoutubeSettings(s), middlewaresEx...))
	router.HandlerFunc(http.MethodPost, "/opmlupload", use(YoutubeOpmlUpload(s), middlewaresExCSRF...))
	router.HandlerFunc(http.MethodGet, "/reddit", use(Reddit(s, taskLastRunFunc), middlewaresEx...))
	router.HandlerFunc(http.MethodGet, "/reddit-settings", use(RedditSettings(s), middlewaresEx...))
	router.HandlerFunc(http.MethodGet, "/twitter", use(Twitter(s, taskLastRunFunc), middlewaresEx...))
	router.HandlerFunc(http.MethodGet, "/twitter-settings", use(TwitterSettings(s), middlewaresEx...))
	// router.HandlerFunc(http.MethodGet, "/instagram", use(Instagram(s, taskLastRunFunc TaskLastRunFunc), middlewaresEx...)) // TODO disabled due to the public insta api being limited to a few requests/day
	// router.HandlerFunc(http.MethodGet, "/instagram-settings", use(InstagramSettings(s), middlewaresEx...))

	router.HandlerFunc(http.MethodGet, "/partial-renderer/cards", use(Cards(s, taskLastRunFunc), middlewaresEx...))
	router.HandlerFunc(http.MethodPost, "/partial-renderer/settings-account-selection", use(SettingAccountSelection(s), middlewaresEx...))
	router.HandlerFunc(http.MethodPost, "/partial-renderer/settings-channel-table", use(SettingChannelTable(s), middlewaresEx...))

	router.HandlerFunc(http.MethodPost, "/api/v1/account", use(rest.AddAccount(s), middlewaresExCSRF...))
	router.HandlerFunc(http.MethodDelete, "/api/v1/account", use(rest.DeleteAccount(s), middlewaresExCSRF...))
	router.HandlerFunc(http.MethodPost, "/api/v1/channel", use(rest.AddChannel(s, channelMetaDataProviderFactory), middlewaresExCSRF...))
	router.HandlerFunc(http.MethodDelete, "/api/v1/channel", use(rest.DeleteChannel(s), middlewaresExCSRF...))
	router.HandlerFunc(http.MethodHead, "/api/v1/channel", use(rest.ValidateChannel(s, channelDataValidatorFactory), middlewaresExCSRF...))
	router.HandlerFunc(http.MethodGet, "/api/v1/content", use(rest.ContentCount(s), middlewaresExCSRF...))

	router.HandlerFunc(http.MethodGet, "/login/oauth2", use(Oauth2LoginHandler(s), middlewares...))
	router.HandlerFunc(http.MethodGet, "/login/oauth2/callback",
		use(Oauth2LoginCallbackHandler(s, "/login/oauth2/callback", "/profile"), middlewares...))

	// static files
	router.HEAD("/static/*filepath", staticFileHandler(s))
	router.GET("/static/*filepath", staticFileHandler(s))

	// error handlers
	router.NotFound = NotFound("/profile")
}

// RedirectTLS redirects any trafic from :80 to port (443)
// requires admin rights to bind & listen to port 80
// otherwise will fail silently
func RedirectTLS(port string) {
	go func() {
		err := http.ListenAndServe(":80", redirectTLS(port))
		if err != nil {
			logging.Println(logging.Info, err)
		}
	}()
}

func redirectTLS(port string) http.HandlerFunc {
	if port == "443" {
		port = ""
	} else {
		port = ":" + port
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		raSplit := strings.Split(r.RemoteAddr, ":")
		if len(raSplit) == 2 {
			ip := raSplit[0]
			http.Redirect(rw, r, "https://"+ip+port+r.RequestURI, http.StatusMovedPermanently)
			return
		}
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		logging.Println(logging.Info, "RemoteAddr could not be resolved")
	}
}

func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	if len(middleware) == 0 {
		return h
	}
	wrapped := h
	for i := 0; i < len(middleware); i++ {
		wrapped = middleware[i](wrapped)
	}
	return wrapped
}

func staticFileHandler(s *server.Server) httprouter.Handle {
	fileServer := http.FileServer(http.Dir("./static"))
	fileServer = s.Minifier.Middleware(fileServer)
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.URL.Path = p.ByName("filepath")
		fileServer.ServeHTTP(w, r)
	}
}
