package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/gurodrigues-dev/tcp-server-client/session"
)

type SubmissionParams struct {
	JobID       int    `json:"job_id"`
	ClientNonce string `json:"client_nonce"`
	Result      string `json:"result"`
}

type SubmissionInfo struct {
	Username  string
	JobID     int
	Timestamp time.Time
	Result    string
}

func NewSubmissionInfo(username string, jobID int, result string) *SubmissionInfo {
	return &SubmissionInfo{
		Username:  username,
		JobID:     jobID,
		Timestamp: time.Now(),
		Result:    result,
	}
}

type SubmissionManager struct {
	authManager         *AuthManager
	clientNoncesUsed    map[string]map[string]bool
	lastSubmissionTime  map[string]time.Time
	resultsByUserMinute map[string][]SubmissionInfo
	mu                  sync.RWMutex
}

func NewSubmissionManager(authManager *AuthManager) *SubmissionManager {
	return &SubmissionManager{
		authManager:         authManager,
		clientNoncesUsed:    make(map[string]map[string]bool),
		lastSubmissionTime:  make(map[string]time.Time),
		resultsByUserMinute: make(map[string][]SubmissionInfo),
	}
}

func (sm *SubmissionManager) HandleSubmit(session *session.Session, submission *SubmissionParams) (*Response, error) {
	if session == nil {
		return &Response{Error: "Invalid Session"}, fmt.Errorf("session is nil")
	}

	if submission == nil {
		return &Response{Error: "Invalid Submission"}, fmt.Errorf("submission is nil")
	}

	sessionKey := session.Username
	if sessionKey == "" {
		return &Response{Error: "Session has no username"}, fmt.Errorf("session has no username")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if lastTime, exists := sm.lastSubmissionTime[sessionKey]; exists {
		if time.Since(lastTime) < time.Second {
			return &Response{Error: "Rate Limit for submissions"}, nil
		}
	}

	jobID, _ := session.GetJobInfo()
	if jobID != submission.JobID {
		return &Response{Error: "Job does not exist"}, nil
	}

	if _, exists := sm.clientNoncesUsed[sessionKey]; !exists {
		sm.clientNoncesUsed[sessionKey] = make(map[string]bool)
	}

	if sm.clientNoncesUsed[sessionKey][submission.ClientNonce] {
		return &Response{Error: "Duplicate submission"}, nil
	}

	expectedHash := ComputeSHA256(session.ServerNonce, submission.ClientNonce)
	if expectedHash != submission.Result {
		return &Response{Error: "Invalid result"}, nil
	}

	sm.lastSubmissionTime[sessionKey] = time.Now()
	sm.clientNoncesUsed[sessionKey][submission.ClientNonce] = true

	minuteKey := fmt.Sprintf("%s:%s", sessionKey, time.Now().Format("200601021504"))
	submissionInfo := NewSubmissionInfo(sessionKey, submission.JobID, submission.Result)
	sm.resultsByUserMinute[minuteKey] = append(sm.resultsByUserMinute[minuteKey], *submissionInfo)

	return &Response{Result: true}, nil
}

func ComputeSHA256(serverNonce, clientNonce string) string {
	data := serverNonce + clientNonce
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
