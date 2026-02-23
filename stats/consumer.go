package stats

import (
	"log"

	"github.com/gurodrigues-dev/tcp-server-client/internal/messaging"
	"github.com/gurodrigues-dev/tcp-server-client/submission"
	"github.com/gurodrigues-dev/tcp-server-client/submission/store"
)

type StatsConsumer struct {
	topic      string
	consumer   messaging.Consumer
	repository *store.SubmissionRepository
	done       chan struct{}
}

func NewStatsConsumer(topic string, consumer messaging.Consumer, repository *store.SubmissionRepository) *StatsConsumer {
	return &StatsConsumer{
		topic:      topic,
		consumer:   consumer,
		repository: repository,
		done:       make(chan struct{}),
	}
}

func (s *StatsConsumer) Start() error {
	if err := s.consumer.Subscribe(s.topic, s.handler); err != nil {
		return err
	}

	if err := s.consumer.Start(); err != nil {
		return err
	}

	log.Println("[StatsConsumer] Succesfully started")
	return nil
}

func (s *StatsConsumer) Stop() error {
	if err := s.consumer.Stop(); err != nil {
		return err
	}
	log.Println("[StatsConsumer] Stopped successfully")
	return nil
}

func (s *StatsConsumer) Close() error {
	close(s.done)
	if err := s.consumer.Close(); err != nil {
		return err
	}
	log.Println("[StatsConsumer] Resources released")
	return nil
}

func (s *StatsConsumer) handler(data []byte) error {
	msg, err := submission.Deserialize(data)
	if err != nil {
		log.Printf("[StatsConsumer] Failed to deserialize message: %v", err)
		return err
	}

	submissionStat := submission.NewSubmissionStat(msg.Username, msg.Timestamp, msg.SubmissionCount)

	if err := s.repository.SaveSubmissionStat(submissionStat); err != nil {
		log.Printf("[StatsConsumer] Failed to save statistic to the database: %v", err)
		return err
	}

	log.Printf("[StatsConsumer] Statistic saved successfully: %s", msg)
	return nil
}
