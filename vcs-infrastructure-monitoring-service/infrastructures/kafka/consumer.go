package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/infrastructures/elasticsearch"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/env"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type IKafkaConsumer interface {
	Start(ctx context.Context) error
	Close() error
}

type InfrastructureEvent struct {
	InstanceID string                 `json:"instance_id"`
	UserID     string                 `json:"user_id"`
	Type       string                 `json:"type"`
	Action     string                 `json:"action"`
	Timestamp  string                 `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type kafkaConsumer struct {
	readers  []*kafka.Reader
	esClient elasticsearch.IElasticsearchClient
	logger   logger.ILogger
}

func NewKafkaConsumer(env env.KafkaEnv, esClient elasticsearch.IElasticsearchClient, logger logger.ILogger) IKafkaConsumer {
	readers := make([]*kafka.Reader, 0, len(env.Topics))
	for _, topic := range env.Topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers: env.Brokers,
			GroupID: env.GroupID,
			Topic:   topic,
		})
		readers = append(readers, reader)
	}

	return &kafkaConsumer{
		readers:  readers,
		esClient: esClient,
		logger:   logger,
	}
}

func (kc *kafkaConsumer) Start(ctx context.Context) error {
	kc.logger.Info("kafka consumer started", zap.Int("topics", len(kc.readers)))

	for _, reader := range kc.readers {
		r := reader
		go kc.consumeMessages(ctx, r)
	}

	return nil
}

func (kc *kafkaConsumer) consumeMessages(ctx context.Context, reader *kafka.Reader) {
	for {
		select {
		case <-ctx.Done():
			kc.logger.Info("kafka consumer stopped for topic")
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				kc.logger.Error("failed to read kafka message", zap.Error(err))
				continue
			}

			var event InfrastructureEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				kc.logger.Error("failed to unmarshal event", zap.Error(err), zap.String("topic", msg.Topic))
				continue
			}

			kc.logger.Info("received event",
				zap.String("topic", msg.Topic),
				zap.String("instance_id", event.InstanceID),
				zap.String("action", event.Action))

			logEntry := elasticsearch.LogEntry{
				InstanceID: event.InstanceID,
				UserID:     event.UserID,
				Type:       event.Type,
				Action:     event.Action,
				Message:    fmt.Sprintf("%s %s", event.Type, event.Action),
				Level:      "info",
				Metadata:   event.Metadata,
			}

			if err := kc.esClient.IndexLog(ctx, logEntry); err != nil {
				kc.logger.Error("failed to index log", zap.Error(err))
			} else {
				kc.logger.Info("indexed event to elasticsearch",
					zap.String("instance_id", event.InstanceID),
					zap.String("action", event.Action))
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

