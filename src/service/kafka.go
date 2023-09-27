package service

import (
	"context"
	"sync"

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

	go s.Consume(ctx, topic, partition)
	return conn, nil
}

func (s *Kafka) Consume(ctx context.Context, topic string, partition int) {
	conn, err := kafka.DialLeader(ctx, "tcp", s.cfg.Kafka.Host, topic, partition)
	if err != nil {
		s.logger.Error("cannot connect to topic "+topic, zap.Error(err))
		return
	}

	s.topicsMu.Lock()
	s.topics = append(s.topics, conn)
	s.topicsMu.Unlock()

	b := make([]byte, 10e3) // 10KB max per message
	for {
		n, err := conn.Read(b)
		if err != nil {
			//s.logger.Error("cannot read from topic "+topic, zap.Error(err))
			return
		}
		s.logger.Info("consumer", zap.ByteString("users", b[:n]))
	}
}

func (s *Kafka) CloseTopics() {
	s.topicsMu.Lock()
	defer s.topicsMu.Unlock()
	for _, topic := range s.topics {
		topic.Close()
	}
}
