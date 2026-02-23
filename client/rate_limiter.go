package client

import (
	"sync"
	"time"
)

type RateLimiter struct {
	lastSubmission time.Time
	firstPending   time.Time
	mu             sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if rl.lastSubmission.IsZero() {
		rl.lastSubmission = now
		rl.firstPending = now
		return
	}

	elapsed := now.Sub(rl.lastSubmission)

	if elapsed < time.Second {
		sleepDuration := time.Second - elapsed
		rl.mu.Unlock()
		time.Sleep(sleepDuration)
		rl.mu.Lock()
		rl.lastSubmission = time.Now()
	} else {
		rl.lastSubmission = now
	}

	if rl.firstPending.IsZero() || now.Sub(rl.firstPending) >= time.Minute {
		rl.firstPending = now
	}
}

func (rl *RateLimiter) CanSubmit() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.firstPending.IsZero() {
		return true
	}

	return time.Since(rl.firstPending) >= time.Minute
}

func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.lastSubmission = time.Time{}
	rl.firstPending = time.Time{}
}
