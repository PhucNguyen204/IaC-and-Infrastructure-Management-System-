package kafka

import (
	"context"
	"encoding/json"

	"github.com/PhucNguyen204/vcs-healthcheck-service/dto"
	"github.com/PhucNguyen204/vcs-healthcheck-service/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type IKafkaConsumer interface {
	Start(ctx context.Context, handler EventHandler) error
	Close() error
}

// EventHandler processes lifecycle events
type EventHandler func(ctx context.Context, event dto.LifecycleEvent) error

type kafkaConsumer struct {
	readers []*kafka.Reader
	logger  logger.ILogger
}

type ConsumerConfig struct {
	Brokers       []string
	ConsumerGroup string
	Topics        []string
}

func NewKafkaConsumer(config ConsumerConfig, logger logger.ILogger) IKafkaConsumer {
	readers := make([]*kafka.Reader, 0, len(config.Topics))

	for _, topic := range config.Topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  config.Brokers,
			GroupID:  config.ConsumerGroup,
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 10e6, // 10MB
		})
		readers = append(readers, reader)
	}

	return &kafkaConsumer{
		readers: readers,
		logger:  logger,
	}
}

func (kc *kafkaConsumer) Start(ctx context.Context, handler EventHandler) error {
	kc.logger.Info("starting kafka consumer", zap.Int("topics", len(kc.readers)))

	for _, reader := range kc.readers {
		r := reader
		go kc.consumeMessages(ctx, r, handler)
	}

	return nil
}

func (kc *kafkaConsumer) consumeMessages(ctx context.Context, reader *kafka.Reader, handler EventHandler) {
	topic := reader.Config().Topic
	kc.logger.Info("consuming from topic", zap.String("topic", topic))

	for {
		select {
		case <-ctx.Done():
			kc.logger.Info("stopping consumer for topic", zap.String("topic", topic))
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				kc.logger.Error("failed to read message", zap.Error(err), zap.String("topic", topic))
				continue
			}

			var event dto.LifecycleEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				kc.logger.Error("failed to unmarshal event",
					zap.Error(err),
					zap.String("topic", topic),
					zap.String("raw", string(msg.Value)))
				continue
			}

			kc.logger.Info("received lifecycle event",
				zap.String("topic", topic),
				zap.String("infrastructure_id", event.InfrastructureID),
				zap.String("action", event.Action),
				zap.String("status", event.Status))

			if err := handler(ctx, event); err != nil {
				kc.logger.Error("failed to handle event", zap.Error(err))
			}
		}
	}
}

func (kc *kafkaConsumer) Close() error {
	for _, reader := range kc.readers {
		if err := reader.Close(); err != nil {
			kc.logger.Error("failed to close kafka reader", zap.Error(err))
		}
	}
	return nil
}
