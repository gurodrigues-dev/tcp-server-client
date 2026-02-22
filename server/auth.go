package server

import (
	"errors"
	"sync"

	"github.com/gurodrigues-dev/tcp-server-client/session"
)

type AuthManager struct {
	sessions map[string]*session.Session
	mu       sync.RWMutex
}

type AuthParams struct {
	Username string `json:"username"`
}

type Authenticator interface {
	Authorize(session *session.Session, authParams *AuthParams) error
	GetSession(username string) (*session.Session, bool)
	RemoveSession(username string)
	GetAllSessions() []*session.Session
}

func NewAuthManager() Authenticator {
	return &AuthManager{
		sessions: make(map[string]*session.Session),
	}
}

func (am *AuthManager) Authorize(session *session.Session, authParams *AuthParams) error {
	if authParams == nil || authParams.Username == "" {
		return errors.New("username cannot be empty")
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	session.Username = authParams.Username

	am.sessions[authParams.Username] = session

	return nil
}

func (am *AuthManager) GetSession(username string) (*session.Session, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	session, exists := am.sessions[username]
	return session, exists
}

func (am *AuthManager) RemoveSession(username string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	delete(am.sessions, username)
}

func (am *AuthManager) GetAllSessions() []*session.Session {
	am.mu.RLock()
	defer am.mu.RUnlock()

	sessions := make([]*session.Session, 0, len(am.sessions))
	for _, session := range am.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}
