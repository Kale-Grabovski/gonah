package service

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/gonah/src/domain"
)

const (
	producerFlushMs   = 300
	consumerTimeoutMs = 100
	closeSleepMs      = 500 // should be > producerFlushMs + consumerTimeoutMs
)

type Kafka struct {
	cfg          *domain.Config
	logger       domain.Logger
	consumerDone chan struct{}
	producerDone chan struct{}
}

func NewKafka(cfg *domain.Config, logger domain.Logger) *Kafka {
	return &Kafka{
		cfg:          cfg,
		logger:       logger,
		consumerDone: make(chan struct{}),
		producerDone: make(chan struct{}),
	}
}

func (s *Kafka) GetProducer(topic string, partition int32, ch <-chan []byte) error {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": s.cfg.Kafka.Host})
	if err != nil {
		return err
	}

	go func() {
		if partition == 0 {
			partition = kafka.PartitionAny
		}
		for {
			select {
			case m := <-ch:
				err = producer.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: partition},
					Value:          m,
				}, nil)
				if err != nil {
					s.logger.Error("cannot producer a message", zap.Error(err))
				}
			case <-s.producerDone:
				producer.Flush(producerFlushMs)
				producer.Close()
				s.logger.Info("producer closed")
				return
			default:
			}
		}
	}()
	go s.Consume(topic)
	return nil
}

func (s *Kafka) Consume(topic string) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": s.cfg.Kafka.Host,
		"group.id":          "pizdec",
		"auto.offset.reset": "latest",
		"fetch.min.bytes":   "1",
	})
	if err != nil {
		s.logger.Error("cannot create consumer", zap.Error(err))
		return
	}

	consumer.SubscribeTopics([]string{topic}, nil)

	for {
		msg, err := consumer.ReadMessage(consumerTimeoutMs * time.Millisecond)
		if err == nil {
			s.logger.Info("kafka consumer", zap.ByteString("msg", msg.Value))
		} else if !err.(kafka.Error).IsTimeout() {
			s.logger.Error("consumer error", zap.Error(err))
			break
		}

		select {
		case <-s.consumerDone:
			consumer.Close()
			s.logger.Info("consumer closed")
			return
		default:
		}
	}
}

func (s *Kafka) Close() {
	s.consumerDone <- struct{}{}
	s.producerDone <- struct{}{}
	time.Sleep(closeSleepMs * time.Millisecond)
}
