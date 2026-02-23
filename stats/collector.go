package stats

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gurodrigues-dev/tcp-server-client/internal/messaging"
	"github.com/gurodrigues-dev/tcp-server-client/submission"
)

type AggregatedStat struct {
	Username        string
	Timestamp       time.Time
	SubmissionCount int
}

type SubmissionCollector struct {
	counters           map[string]int
	lastSubmissionTime map[string]time.Time
	mu                 sync.RWMutex
	producer           messaging.Producer
	exchangeTopic      string
	exchangeName       string
}

func NewSubmissionCollector(topic, exchange string, producer messaging.Producer) *SubmissionCollector {
	return &SubmissionCollector{
		counters:           make(map[string]int),
		lastSubmissionTime: make(map[string]time.Time),
		producer:           producer,
		exchangeTopic:      topic,
		exchangeName:       exchange,
	}
}

func (sc *SubmissionCollector) RecordSubmission(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.counters[username]++

	sc.lastSubmissionTime[username] = time.Now()

	return nil
}

func (sc *SubmissionCollector) Flush() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if len(sc.counters) == 0 {
		return nil
	}

	var errors []error

	now := time.Now()
	minuteStart := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		0,
		0,
		now.Location(),
	)

	for username, count := range sc.counters {
		if count <= 0 {
			continue
		}

		msg, err := submission.NewSubmissionStatsMessage(username, minuteStart, count)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to create message for user %s: %w", username, err))
			continue
		}

		if err := sc.producer.Publish(sc.exchangeTopic, sc.exchangeName, msg); err != nil {
			errors = append(errors, fmt.Errorf("failed to publish message for user %s: %w", username, err))
			continue
		}

		log.Printf("Successfully published submission stats for user %s: count=%d, timestamp=%s", username, count, minuteStart.Format(time.RFC3339))
	}

	sc.counters = make(map[string]int)
	sc.lastSubmissionTime = make(map[string]time.Time)

	if len(errors) > 0 {
		return fmt.Errorf("flush completed with %d errors: %v", len(errors), errors)
	}

	return nil
}

func (sc *SubmissionCollector) GetAggregatedStats(username string) ([]*AggregatedStat, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var stats []*AggregatedStat

	now := time.Now()
	minuteStart := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		0,
		0,
		now.Location(),
	)

	if username != "" {
		count, exists := sc.counters[username]
		if !exists || count <= 0 {
			return stats, nil
		}

		aggregated := &AggregatedStat{
			Username:        username,
			Timestamp:       minuteStart,
			SubmissionCount: count,
		}

		stats = append(stats, aggregated)
	} else {
		for user, count := range sc.counters {
			if count <= 0 {
				continue
			}

			aggregated := &AggregatedStat{
				Username:        user,
				Timestamp:       minuteStart,
				SubmissionCount: count,
			}

			stats = append(stats, aggregated)
		}
	}

	return stats, nil
}

func (sc *SubmissionCollector) Close() error {
	_ = sc.Flush()

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.counters = make(map[string]int)
	sc.lastSubmissionTime = make(map[string]time.Time)

	return sc.producer.Close()
}

func (sc *SubmissionCollector) GetCurrentCounters() map[string]int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	countersCopy := make(map[string]int)
	for k, v := range sc.counters {
		countersCopy[k] = v
	}

	return countersCopy
}

func (sc *SubmissionCollector) GetLastSubmissionTime(username string) (time.Time, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	t, exists := sc.lastSubmissionTime[username]
	return t, exists
}
