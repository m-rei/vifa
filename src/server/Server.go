package server

import (
	"context"
	"encoding/json"
	"net/http"
	"visual-feed-aggregator/src/database"
	"visual-feed-aggregator/src/database/services"
	"visual-feed-aggregator/src/server/middleware"
	"visual-feed-aggregator/src/util/logging"

	"github.com/jmoiron/sqlx"
	"github.com/tdewolff/minify/v2"
	"golang.org/x/oauth2"
)

// Server ...
type Server struct {
	Sessions  middleware.SessionManager
	DB        *sqlx.DB
	Minifier  *minify.M
	Env       map[string]string
	OAuth2Cfg oauth2.Config
	Services  *services.ServiceCollection

	certFile string
	keyFile  string
	server   http.Server
}

// GoogleUserInfo ...
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`

	Username string
}

// NewServer creates and configures a new server instance
func NewServer(db *sqlx.DB, services *services.ServiceCollection, sessionStore database.SessionStore, oauth2Cfg oauth2.Config, env map[string]string) *Server {
	res := Server{
		Sessions:  middleware.NewSessionManager(sessionStore),
		DB:        db,
		Minifier:  middleware.NewMinifier(),
		Env:       env,
		OAuth2Cfg: oauth2Cfg,

		Services: services,

		certFile: env["CRT"],
		keyFile:  env["KEY"],
	}
	return &res
}

// Run calls listen and serve / serveTLS depending on the flags
func (s *Server) Run(handler http.Handler) error {
	logging.Println(logging.Info, "listening on port:", s.Env["PORT"])
	s.server = http.Server{Addr: ":" + s.Env["PORT"], Handler: handler}
	return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
}

// Stop stops the server
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// GoogleUserInfoFromSession ...
func GoogleUserInfoFromSession(s *Server, sid string) *GoogleUserInfo {
	if sid == "" {
		return nil
	}
	var gui GoogleUserInfo
	u, ok := s.Sessions.Store.Get(sid, "user")
	if !ok {
		return nil
	}
	if err := json.Unmarshal([]byte(u), &gui); err != nil {
		return nil
	}
	return &gui
}
