package submission

import (
	"time"
)

type SubmissionStat struct {
	ID              int       `db:"id" json:"id"`
	Username        string    `db:"username" json:"username"`
	Timestamp       time.Time `db:"timestamp" json:"timestamp"`
	SubmissionCount int       `db:"submission_count" json:"submission_count"`
}

type AggregatedSubmissions struct {
	Username        string
	Timestamp       time.Time
	SubmissionCount int
}

type Repository interface {
	SaveSubmissionStat(stat *SubmissionStat) error
	GetSubmissionStatsByUsername(username string, limit int) ([]*SubmissionStat, error)
}

func NewSubmissionStat(username string, timestamp time.Time, count int) *SubmissionStat {
	return &SubmissionStat{
		Username:        username,
		Timestamp:       timestamp,
		SubmissionCount: count,
	}
}

func (s *SubmissionStat) ToAggregated(username string, timestamp time.Time) *AggregatedSubmissions {
	return &AggregatedSubmissions{
		Username:        username,
		Timestamp:       timestamp,
		SubmissionCount: s.SubmissionCount,
	}
}
