package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
	"visual-feed-aggregator/src/database"
	"visual-feed-aggregator/src/util/logging"

	"github.com/google/uuid"
)

// SessionManager ...
type SessionManager struct {
	Store database.SessionStore

	// cookie settings, default to a secure configuration
	secure   bool
	httpOnly bool
	sameSite http.SameSite

	staleSessionMaxMinutesInactive int
	staleSessionMux                sync.Mutex
}

// NewSessionManager ...
func NewSessionManager(store database.SessionStore) SessionManager {
	return SessionManager{
		Store:    store,
		secure:   true,
		httpOnly: true,
		sameSite: http.SameSiteLaxMode,

		staleSessionMaxMinutesInactive: 12 * 60.,
	}
}

// cookie creates a new digitally signed cookie
func cookie(SID string, digiSignKey []byte, secure, httpOnly bool, sameSite http.SameSite) *http.Cookie {
	mac := hmac.New(sha256.New, digiSignKey)
	mac.Write([]byte(SID))
	signature := mac.Sum(nil)

	return &http.Cookie{
		Name:     "sid",
		Value:    base64.StdEncoding.EncodeToString([]byte(SID + string(signature))),
		Path:     "/",
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: sameSite,
	}
}

func validSignature(msg, msgMac, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	got := mac.Sum(nil)
	return hmac.Equal(got, msgMac)
}

// SessionIDFromRequest retrieves the sid which is base64 encoded in the "sid" received
// from the http request with its digital signature
func (s *SessionManager) SessionIDFromRequest(r *http.Request) string {
	sessionCookie, err := r.Cookie("sid")
	if err != nil {
		return ""
	}
	value, err := base64.StdEncoding.DecodeString(sessionCookie.Value)
	if err != nil {
		return ""
	}

	sid := value[:len(value)-sha256.Size]
	sig := value[len(value)-sha256.Size:]

	if !validSignature(sid, sig, s.Store.SessionKey()) {
		return ""
	}

	return string(sid)
}

const staleSessionCheckIntervalMinutes = 10

// SessionMiddleware ...
func (s *SessionManager) SessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Store.ClearStaleSessions(staleSessionCheckIntervalMinutes, s.staleSessionMaxMinutesInactive)
		sid := s.SessionIDFromRequest(r)
		_, sessionAlive := s.Store.Get(sid, "LastAccess")
		if !sessionAlive {
			sid := uuid.New().String()
			s.Store.Put(sid, "Created", time.Now().UTC().Format(time.RFC3339))
			s.Store.Put(sid, "LastAccess", time.Now().UTC().Format(time.RFC3339))
			s.Store.Put(sid, "CsrfToken", uuid.New().String())
			cookie := cookie(sid, s.Store.SessionKey(), s.secure, s.httpOnly, s.sameSite)
			http.SetCookie(w, cookie)
			r.AddCookie(cookie)
		} else {
			s.Store.Put(sid, "LastAccess", time.Now().UTC().Format(time.RFC3339))
		}
		next(w, r)
	}
}

// AuthorizedMiddleware ...
func (s *SessionManager) AuthorizedMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := s.SessionIDFromRequest(r)
		if sid != "" {
			_, ok := s.Store.Get(sid, "user")
			if ok {
				next(w, r)
				return
			}
		}
		logging.Println(logging.Debug, "unauthorized request")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// CsrfTokenValidator ...
func (s *SessionManager) CsrfTokenValidator(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		csrf := r.Header.Get("csrf")
		sid := s.SessionIDFromRequest(r)
		sessionCSRF, ok := s.Store.Get(sid, "CsrfToken")
		if ok && sid != "" && csrf != sessionCSRF {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logging.Println(logging.Warn, "invalid csrf token")
			return
		}
		next(rw, r)
	}
}
