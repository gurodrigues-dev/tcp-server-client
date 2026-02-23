package session

import (
	"net"
	"sync"
	"time"
)

type Session struct {
	Username    string
	Conn        net.Conn
	ServerNonce string
	JobID       int
	JobHistory  map[int]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	mu          sync.RWMutex
}

func NewSession(username string, conn net.Conn) *Session {
	now := time.Now()
	return &Session{
		Username:   username,
		Conn:       conn,
		JobHistory: make(map[int]string),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (s *Session) SetJob(jobID int, serverNonce string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.JobID = jobID
	s.ServerNonce = serverNonce
	s.JobHistory[jobID] = serverNonce
	s.UpdatedAt = time.Now()
}

func (s *Session) GetJobInfo() (int, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.JobID, s.ServerNonce
}

func (s *Session) UpdateTimestamp() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UpdatedAt = time.Now()
}

func (s *Session) Close() error {
	if s.Conn != nil {
		return s.Conn.Close()
	}
	return nil
}
