package database

// SessionStore interface for session stores (memory, database, ..)
type SessionStore interface {
	Get(id, key string) (string, bool)
	Put(id, key, value string)
	RemoveKey(id, key string)
	RemoveSession(id string)

	SessionKey() []byte
	ClearStaleSessions(checkIntervalMinutes, sessionMaxAgeInactiveMinutes int)
}
