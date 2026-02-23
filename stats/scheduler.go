package stats

import (
	"log"
	"time"
)

type StatsScheduler struct {
	collector *SubmissionCollector
	ticker    *time.Ticker
	done      chan struct{}
	interval  time.Duration
}

func NewStatsScheduler(collector *SubmissionCollector, interval time.Duration) *StatsScheduler {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	return &StatsScheduler{
		collector: collector,
		interval:  interval,
		done:      make(chan struct{}),
	}
}

func (s *StatsScheduler) Start() {
	log.Printf("[StatsScheduler] Starting scheduler: %v", s.interval)
	s.ticker = time.NewTicker(s.interval)

	go func() {
		for {
			select {
			case now := <-s.ticker.C:
				log.Printf("[StatsScheduler] [%s] Starting periodic flushing...", now.Format("15:04:05"))
				counters := s.collector.GetCurrentCounters()
				log.Printf("[StatsScheduler] Counters before flush: %v", counters)

				if err := s.collector.Flush(); err != nil {
					log.Printf("[StatsScheduler] Error during flush: %v", err)
				} else {
					log.Printf("[StatsScheduler] Flush completed successfully")
				}
			case <-s.done:
				return
			}
		}
	}()
}

func (s *StatsScheduler) Stop() {
	log.Printf("[StatsScheduler] Starting shutdown...")
	close(s.done)

	if s.ticker != nil {
		s.ticker.Stop()
	}

	log.Printf("[StatsScheduler] Performing final flush...")
	if err := s.FlushNow(); err != nil {
		log.Printf("[StatsScheduler] Error during final flush: %v", err)
	}

	log.Printf("[StatsScheduler] Scheduler stopped")
}

func (s *StatsScheduler) FlushNow() error {
	log.Printf("[StatsScheduler] Manual flush requested...")
	return s.collector.Flush()
}
