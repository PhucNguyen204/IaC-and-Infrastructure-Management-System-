package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/env"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type IKafkaProducer interface {
	PublishEvent(ctx context.Context, event InfrastructureEvent) error
	Close() error
	IsConnected() bool
}

type InfrastructureEvent struct {
	InstanceID string                 `json:"instance_id"`
	UserID     string                 `json:"user_id"`
	Type       string                 `json:"type"`
	Action     string                 `json:"action"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type kafkaProducer struct {
	writer    *kafka.Writer
	logger    logger.ILogger
	connected bool
	mu        sync.RWMutex
	enabled   bool
}

func NewKafkaProducer(env env.KafkaEnv, logger logger.ILogger) IKafkaProducer {
	// Check if Kafka is configured
	if len(env.Brokers) == 0 || env.Brokers[0] == "" {
		logger.Warn("Kafka not configured, running without event publishing")
		return &kafkaProducer{
			writer:    nil,
			logger:    logger,
			connected: false,
			enabled:   false,
		}
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(env.Brokers...),
		Topic:                  env.Topic,
		Balancer:               &kafka.Hash{},
		AllowAutoTopicCreation: true,
		Async:                  true,
		BatchSize:              100,
		BatchTimeout:           10 * time.Millisecond,
		RequiredAcks:           kafka.RequireOne,
		MaxAttempts:            3,
		WriteBackoffMin:        100 * time.Millisecond,
		WriteBackoffMax:        1 * time.Second,
		Compression:            kafka.Snappy,
	}

	kp := &kafkaProducer{
		writer:    writer,
		logger:    logger,
		connected: false,
		enabled:   true,
	}

	// Test connection in background
	go kp.testConnection(env.Brokers)

	return kp
}

func (kp *kafkaProducer) testConnection(brokers []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		kp.logger.Warn("Kafka connection failed, events will be logged only", zap.Error(err))
		return
	}
	defer conn.Close()

	kp.mu.Lock()
	kp.connected = true
	kp.mu.Unlock()
	kp.logger.Info("Kafka connection established", zap.Strings("brokers", brokers))
}

func (kp *kafkaProducer) IsConnected() bool {
	kp.mu.RLock()
	defer kp.mu.RUnlock()
	return kp.connected
}

func (kp *kafkaProducer) PublishEvent(ctx context.Context, event InfrastructureEvent) error {
	event.Timestamp = time.Now()

	// Log the event regardless of Kafka connection
	kp.logger.Info("infrastructure event",
		zap.String("instance_id", event.InstanceID),
		zap.String("action", event.Action),
		zap.String("type", event.Type))

	// Skip if Kafka is not enabled or connected
	if !kp.enabled || kp.writer == nil {
		return nil
	}

	if !kp.IsConnected() {
		kp.logger.Debug("Kafka not connected, skipping publish")
		return nil
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		kp.logger.Error("failed to marshal event", zap.Error(err))
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.InstanceID),
		Value: eventBytes,
		Time:  event.Timestamp,
	}

	if err := kp.writer.WriteMessages(ctx, msg); err != nil {
		kp.logger.Warn("failed to write message to kafka", zap.Error(err))
		// Don't return error - continue service operation without Kafka
		return nil
	}

	return nil
}

func (kp *kafkaProducer) Close() error {
	if kp.writer != nil {
		return kp.writer.Close()
	}
	return nil
}
