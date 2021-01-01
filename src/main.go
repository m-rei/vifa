package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"time"
	"visual-feed-aggregator/src/database"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/database/services"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/server/pages"
	"visual-feed-aggregator/src/tasks"
	"visual-feed-aggregator/src/util/logging"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
)

const defaultRefreshRateMinutes = 60
const defaultCutoffDays = 7

var ssqlCrtDefault, _ = filepath.Abs(path.Join("res", "certificates", "server.crt"))
var sslKeyDefault, _ = filepath.Abs(path.Join("res", "certificates", "server.key"))

var envVars = []struct {
	name         string
	defaultValue string
}{
	{"TITLE", "vifa"},
	{"REFRESH_RATE_MINUTES", "60"},
	{"CUTOFF_DAYS", "7"},
	{"LOG_LEVEL", "INFO"},
	{"PORT", ""},
	{"CRT", ssqlCrtDefault},
	{"KEY", sslKeyDefault},
	{"DB_USER", ""},
	{"DB_PASS", ""},
	{"DB_ADDRESS", ""},
	{"GOOGLE_OAUTH2_CLIENT_ID", ""},
	{"GOOGLE_OAUTH2_CLIENT_SECRET", ""},
}

var backgroundTasks []tasks.BackgroundTask = []tasks.BackgroundTask{
	tasks.YoutubeBackgroundTask,
	tasks.RedditBackgroundTask,
	tasks.TwitterBackgroundTask,
	// tasks.InstagramBackgroundTask, // TODO disabled due to the public insta api being limited to a few requests/day
	tasks.CleanupBackgroundTask,
}

var backgroundTasksLastRun map[string]time.Time = map[string]time.Time{
	models.KindYoutube: time.Now().UTC(),
	models.KindReddit:  time.Now().UTC(),
	models.KindTwitter: time.Now().UTC(),
	// models.KindInstagram: time.Now().UTC(),
}

func main() {
	// cfg etc.
	env := loadEnvVars()
	setupLogging(env)
	db := loadDB(env)
	db.Ping()
	defer db.Close()
	sessionStore := server.NewMySQLDbSessionStore(db)
	if sessionStore == nil {
		logging.Fatalln("Could not instantiate session store")
	}
	services := services.NewMySQLServiceCollection(db)

	// server
	srv := server.NewServer(db, &services, sessionStore, oauth2Config(env), env)
	router := httprouter.New()
	pages.SetupRoutes(srv, router, getBackgroundTaskLastRun)
	pages.RedirectTLS(env["PORT"])
	go func() {
		if err := srv.Run(router); err != http.ErrServerClosed {
			logging.Println(logging.Info, err)
		}
	}()

	// background tasks
	cutoffDays, err := strconv.ParseInt(env["CUTOFF_DAYS"], 10, 64)
	if err != nil {
		cutoffDays = defaultCutoffDays
	}
	refreshRateMinutes, err := strconv.ParseInt(env["REFRESH_RATE_MINUTES"], 10, 64)
	if err != nil {
		refreshRateMinutes = defaultRefreshRateMinutes
	}
	stopSignal := make(chan bool, len(backgroundTasks))
	startBackgroundTasks(backgroundTasks, backgroundTasksLastRun, stopSignal, db, &services, cutoffDays, refreshRateMinutes)

	// graceful exit
	osSignalExit := make(chan os.Signal, 1)
	signal.Notify(osSignalExit, os.Interrupt, os.Kill)
	<-osSignalExit
	for i := 0; i < len(backgroundTasks); i++ {
		stopSignal <- true
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		logging.Println(logging.Error, err)
	}
}

func setupLogging(env map[string]string) {
	fn := time.Now().UTC().Format(time.RFC3339) + ".log"
	fn, _ = filepath.Abs(path.Join("logs", fn))
	file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		log.SetOutput(file)
	}
	lvl, ok := logging.StrToLogLevel(env["LOG_LEVEL"])
	if ok {
		logging.CurrentLogLevel = lvl
	}
}

func loadEnvVars() map[string]string {
	ret := make(map[string]string)
	for _, e := range envVars {
		env := os.Getenv(e.name)
		if env == "" {
			env = e.defaultValue
		}
		ret[e.name] = env
	}
	return ret
}

func loadDB(env map[string]string) *sqlx.DB {
	db, err := database.OpenDbWithSchema(env["DB_USER"], env["DB_PASS"], env["DB_ADDRESS"])
	if err != nil {
		logging.Fatalln(err)
	}
	return db
}

func oauth2Config(env map[string]string) oauth2.Config {
	return oauth2.Config{
		ClientID:     env["GOOGLE_OAUTH2_CLIENT_ID"],
		ClientSecret: env["GOOGLE_OAUTH2_CLIENT_SECRET"],
		Scopes:       []string{pages.GoogleScopeUserInfo},
		Endpoint:     pages.GoogleEndpoint,
	}
}

func getBackgroundTaskLastRun(kind string) time.Time {
	return backgroundTasksLastRun[kind]
}

func startBackgroundTasks(tasks []tasks.BackgroundTask, lastRun map[string]time.Time,
	stopSignal <-chan bool, db *sqlx.DB, services *services.ServiceCollection, cutoffDays, refreshRateMinutes int64) {

	for _, task := range tasks {
		go task(stopSignal, lastRun, db, services, cutoffDays, refreshRateMinutes)
	}
}
