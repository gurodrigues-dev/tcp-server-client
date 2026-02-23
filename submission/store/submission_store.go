package store

import (
	"database/sql"
	"fmt"

	"github.com/gurodrigues-dev/tcp-server-client/submission"
)

type SubmissionRepository struct {
	db *sql.DB
}

func NewSubmissionRepository(db *sql.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) SaveSubmissionStat(stat *submission.SubmissionStat) error {
	query := `
		INSERT INTO submissions (username, timestamp, submission_count)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(query, stat.Username, stat.Timestamp, stat.SubmissionCount)
	if err != nil {
		return fmt.Errorf("failed to save submission stat: %w", err)
	}

	return nil
}

func (r *SubmissionRepository) GetSubmissionStatsByUsername(username string, limit int) ([]*submission.SubmissionStat, error) {
	query := `
		SELECT id, username, timestamp, submission_count
		FROM submissions
		WHERE username = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, username, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query submission stats: %w", err)
	}
	defer rows.Close()

	var stats []*submission.SubmissionStat
	for rows.Next() {
		stat := &submission.SubmissionStat{}
		err := rows.Scan(&stat.ID, &stat.Username, &stat.Timestamp, &stat.SubmissionCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan submission stat: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning submission stats: %w", err)
	}

	return stats, nil
}
