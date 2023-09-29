package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/gonah/src/domain"
)

type Kafka struct {
	cfg      *domain.Config
	logger   domain.Logger
	topics   []*kafka.Conn
	topicsMu sync.Mutex
}

func NewKafka(cfg *domain.Config, logger domain.Logger) *Kafka {
	return &Kafka{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Kafka) Connect(ctx context.Context, topic string, partition int) (*kafka.Conn, error) {
	conn, err := kafka.DialLeader(ctx, "tcp", s.cfg.Kafka.Host, topic, partition)
	if err != nil {
		return nil, err
	}

	s.topicsMu.Lock()
	s.topics = append(s.topics, conn)
	s.topicsMu.Unlock()

	go s.Consume(ctx, topic)
	return conn, nil
}

func (s *Kafka) Consume(ctx context.Context, topic string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          []string{s.cfg.Kafka.Host},
		GroupID:          "users-group",
		Topic:            topic,
		MinBytes:         1,
		MaxBytes:         10e6, // 10MB
		ReadBatchTimeout: 10 * time.Millisecond,
	})

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			break
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}

	if err := r.Close(); err != nil {
		s.logger.Error("failed to close reader", zap.Error(err))
		return
	}
	fmt.Println("close consumer")
}

func (s *Kafka) CloseTopics() {
	s.topicsMu.Lock()
	defer s.topicsMu.Unlock()
	for _, topic := range s.topics {
		topic.Close()
	}
}
