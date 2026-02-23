package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
	"time"
)

type JobManager struct {
	authManager *AuthManager
	jobID       int
	ticker      *time.Ticker
	mu          sync.Mutex
	done        chan struct{}
}

type JobParams struct {
	JobID       int    `json:"job_id"`
	ServerNonce string `json:"server_nonce"`
}

func NewJobManager(authManager *AuthManager) *JobManager {
	return &JobManager{
		authManager: authManager,
		jobID:       0,
		done:        make(chan struct{}),
	}
}

func (jm *JobManager) Start() {
	jm.ticker = time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-jm.ticker.C:
				jm.updateNonceAndSendJobs()
			case <-jm.done:
				return
			}
		}
	}()
}

func (jm *JobManager) Stop() {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if jm.ticker != nil {
		jm.ticker.Stop()
		close(jm.done)
	}
}

func (jm *JobManager) updateNonceAndSendJobs() {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	serverNonce := jm.generateNonce()

	jm.jobID++

	sessions := jm.authManager.getAllSessions()

	for _, session := range sessions {
		session.SetJob(jm.jobID, serverNonce)

		jobParams := JobParams{
			JobID:       jm.jobID,
			ServerNonce: serverNonce,
		}

		request := Request{
			ID:     0,
			Method: "job",
			Params: jobParams,
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			log.Printf("error marshaling job request for user %s: %v", session.Username, err)
			continue
		}

		_, err = session.Conn.Write(jsonData)
		if err != nil {
			log.Printf("error sending job request to user %s: %v", session.Username, err)
		}
	}
}

func (jm *JobManager) generateNonce() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Printf("error generating random nonce: %v", err)
		return hex.EncodeToString([]byte(time.Now().String()))[:32]
	}
	return hex.EncodeToString(bytes)
}
