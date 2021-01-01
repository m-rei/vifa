package server

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
	"visual-feed-aggregator/src/database"
	"visual-feed-aggregator/src/util/logging"

	"github.com/jmoiron/sqlx"
)

type memorySessionStore struct {
	m              sync.Mutex
	data           map[string]map[string]string
	sessionKey     []byte
	clearTimestamp time.Time
}

// NewMemorySessionStore creates an in-memory store
func NewMemorySessionStore() database.SessionStore {
	sessionKey := make([]byte, 32)
	_, err := rand.Read(sessionKey)
	if err != nil {
		return nil
	}
	return &memorySessionStore{data: make(map[string]map[string]string), sessionKey: sessionKey}
}

func (s *memorySessionStore) get(id string) (map[string]string, bool) {
	ret, ok := s.data[id]
	if ok {
		return ret, ok
	}
	return nil, false
}

func (s *memorySessionStore) Get(id, key string) (string, bool) {
	s.m.Lock()
	defer s.m.Unlock()
	ret, ok := s.get(id)
	if ok {
		return ret[key], ret[key] != ""
	}
	return "", false
}

func (s *memorySessionStore) Put(id, key, value string) {
	s.m.Lock()
	defer s.m.Unlock()
	m, ok := s.get(id)
	if !ok {
		m = make(map[string]string)
	}
	m[key] = value
	s.data[id] = m
}

func (s *memorySessionStore) RemoveKey(id, key string) {
	s.m.Lock()
	defer s.m.Unlock()
	m, ok := s.get(id)
	if ok {
		delete(m, key)
		s.data[id] = m
	}
}

func (s *memorySessionStore) RemoveSession(id string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.data, id)
}

func (s *memorySessionStore) SessionKey() []byte {
	return s.sessionKey
}

func (s *memorySessionStore) ClearStaleSessions(checkIntervalMinutes, sessionMaxAgeInactiveMinutes int) {
	doClear := false

	s.m.Lock()
	if time.Since(s.clearTimestamp).Minutes() > float64(checkIntervalMinutes) {
		s.clearTimestamp = time.Now()
		doClear = true
	}
	s.m.Unlock()
	if !doClear {
		return
	}

	utcNow := time.Now().UTC()
	for sid, m := range s.data {
		lastAccess, err := time.Parse(time.RFC3339, m["LastAccess"])
		if err == nil && utcNow.Sub(lastAccess).Minutes() > float64(sessionMaxAgeInactiveMinutes) {
			delete(s.data, sid)
		}
	}
}

type dbSessionStore struct {
	db             *sqlx.DB
	clearTimestamp time.Time
	sessionKey     []byte
}

// NewMySQLDbSessionStore creates a db session store
func NewMySQLDbSessionStore(db *sqlx.DB) database.SessionStore {
	ret := &dbSessionStore{db: db}
	if !ret.loadSessionKey() {
		return nil
	}
	return ret
}

func (s *dbSessionStore) loadSessionKey() bool {
	query := `
	SELECT v
	FROM sessions
	WHERE sid = "session" AND k = "key"
	`
	r := s.db.QueryRow(query)
	var b64Key string
	err := r.Scan(&b64Key)
	if err != nil {
		logging.Println(logging.Debug, err)
		return false
	}
	s.sessionKey, err = base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		logging.Println(logging.Debug, err)
	}
	return err == nil
}

func (s *dbSessionStore) Get(id, key string) (string, bool) {
	query := `
	SELECT v
	FROM sessions
	WHERE sid = ? AND k = ?
	`
	r := s.db.QueryRow(query, id, key)
	var ret string
	err := r.Scan(&ret)
	if err != nil {
		logging.Println(logging.Debug, err)
		return "", false
	}
	return ret, true
}

func (s *dbSessionStore) Put(id, key, value string) {
	query := `
	REPLACE INTO sessions (sid, k, v)
	VALUES (?, ?, ?)
	`
	_, err := s.db.Exec(query, id, key, value)
	if err != nil {
		logging.Println(logging.Debug, err)
	}
}

func (s *dbSessionStore) RemoveKey(id, key string) {
	query := `
	DELETE FROM sessions
	WHERE sid = ? AND k = ?
	`
	_, err := s.db.Exec(query, id, key)
	if err != nil {
		logging.Println(logging.Debug, err)
	}
}

func (s *dbSessionStore) RemoveSession(id string) {
	query := `
	DELETE FROM sessions
	WHERE sid = ?
	`
	_, err := s.db.Exec(query, id)
	if err != nil {
		logging.Println(logging.Debug, err)
	}
}

func (s *dbSessionStore) SessionKey() []byte {
	return s.sessionKey
}

func (s *dbSessionStore) ClearStaleSessions(checkIntervalMinutes, sessionMaxAgeInactiveMinutes int) {
	queries := []string{`
	CREATE TEMPORARY TABLE stale_sessions
	SELECT sid
	FROM sessions
	WHERE k = "LastAccess" AND ABS(TIMESTAMPDIFF(MINUTE, STR_TO_DATE(?, "%Y-%m-%dT%H:%i:%sZ"), STR_TO_DATE(v, "%Y-%m-%dT%H:%i:%sZ"))) > ?
	;
	`, `
	DELETE FROM sessions
	WHERE sid IN (
		SELECT sid
		FROM stale_sessions
	);
	`, `
	DROP TABLE stale_sessions;
	`,
	}
	if time.Since(s.clearTimestamp).Minutes() > float64(checkIntervalMinutes) {
		s.clearTimestamp = time.Now().UTC()

		tx, err := s.db.Begin()
		if err != nil {
			logging.Println(logging.Debug, err)
			return
		}

		_, err = tx.Exec(queries[0], s.clearTimestamp.Format(time.RFC3339), sessionMaxAgeInactiveMinutes)
		if err != nil {
			logging.Println(logging.Debug, err)
			tx.Rollback()
			return
		}
		_, err = tx.Exec(queries[1])
		if err != nil {
			logging.Println(logging.Debug, err)
			tx.Rollback()
			return
		}
		_, err = tx.Exec(queries[2])
		if err != nil {
			logging.Println(logging.Debug, err)
			tx.Rollback()
			return
		}

		if err := tx.Commit(); err != nil {
			logging.Println(logging.Debug, err)
		}
	}
}
