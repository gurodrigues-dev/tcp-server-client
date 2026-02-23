package submission

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type SubmissionStatsMessage struct {
	Username        string    `json:"username"`
	Timestamp       time.Time `json:"timestamp"`
	SubmissionCount int       `json:"submission_count"`
	MessageType     string    `json:"message_type"`
}

func NewSubmissionStatsMessage(username string, timestamp time.Time, count int) (*SubmissionStatsMessage, error) {
	if username == "" {
		return nil, errors.New("username cant be empty")
	}

	if timestamp.IsZero() {
		return nil, errors.New("timestamp cant be zero")
	}

	if count <= 0 {
		return nil, errors.New("count must be greater than zero")
	}

	return &SubmissionStatsMessage{
		Username:        username,
		Timestamp:       timestamp,
		SubmissionCount: count,
		MessageType:     "submission_stats",
	}, nil
}

func (m *SubmissionStatsMessage) Serialize() ([]byte, error) {
	if m == nil {
		return nil, errors.New("message cannot be nil")
	}

	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("error serializing message: %w", err)
	}

	return data, nil
}

func Deserialize(data []byte) (*SubmissionStatsMessage, error) {
	if len(data) == 0 {
		return nil, errors.New("data cannot be empty")
	}

	var msg SubmissionStatsMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("error deserializing message: %w", err)
	}

	if msg.MessageType != "submission_stats" {
		return nil, fmt.Errorf("invalid message type: expected 'submission_stats', got '%s'", msg.MessageType)
	}

	return &msg, nil
}

func (m *SubmissionStatsMessage) String() string {
	if m == nil {
		return "SubmissionStatsMessage(nil)"
	}

	return fmt.Sprintf(
		"SubmissionStatsMessage{Username: %s, Timestamp: %s, SubmissionCount: %d, MessageType: %s}",
		m.Username,
		m.Timestamp.Format(time.RFC3339),
		m.SubmissionCount,
		m.MessageType,
	)
}
