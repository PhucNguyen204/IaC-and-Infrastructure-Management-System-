package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/env"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type CacheInvalidator interface {
	InvalidateCluster(ctx context.Context, clusterID string) error
}

type IEventConsumer interface {
	Start(ctx context.Context) error
	Close() error
}

type eventConsumer struct {
	reader  *kafka.Reader
	cache   CacheInvalidator
	logger  logger.ILogger
	enabled bool
}

func NewEventConsumer(env env.KafkaEnv, cache CacheInvalidator, logger logger.ILogger) IEventConsumer {
	// Check if Kafka is configured
	if len(env.Brokers) == 0 || env.Brokers[0] == "" {
		logger.Warn("Kafka not configured, consumer disabled")
		return &eventConsumer{
			reader:  nil,
			cache:   cache,
			logger:  logger,
			enabled: false,
		}
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        env.Brokers,
		Topic:          env.Topic,
		GroupID:        "infrastructure-processor",
		MinBytes:       10e3,
		MaxBytes:       10e6,
		MaxWait:        500 * time.Millisecond,
		CommitInterval: time.Second,
	})

	return &eventConsumer{
		reader:  reader,
		cache:   cache,
		logger:  logger,
		enabled: true,
	}
}

func (c *eventConsumer) Start(ctx context.Context) error {
	if !c.enabled || c.reader == nil {
		c.logger.Info("Kafka consumer disabled, skipping start")
		return nil
	}

	c.logger.Info("starting kafka consumer", zap.String("group", "infrastructure-processor"))

	go func() {
		retryDelay := time.Second
		maxRetryDelay := 30 * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := c.reader.ReadMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return // Context cancelled, exit gracefully
					}
					c.logger.Warn("failed to read message, retrying...",
						zap.Error(err),
						zap.Duration("retry_delay", retryDelay))

					time.Sleep(retryDelay)
					retryDelay = min(retryDelay*2, maxRetryDelay)
					continue
				}

				// Reset retry delay on successful read
				retryDelay = time.Second

				var event InfrastructureEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					c.logger.Error("failed to unmarshal event", zap.Error(err))
					continue
				}

				if err := c.handleEvent(ctx, event); err != nil {
					c.logger.Error("failed to handle event",
						zap.String("action", event.Action),
						zap.String("instance_id", event.InstanceID),
						zap.Error(err))
				}
			}
		}
	}()

	return nil
}

func (c *eventConsumer) handleEvent(ctx context.Context, event InfrastructureEvent) error {
	c.logger.Info("handling event",
		zap.String("action", event.Action),
		zap.String("type", event.Type),
		zap.String("instance_id", event.InstanceID))

	switch event.Action {
	case "cluster.created", "cluster.updated", "cluster.scaled",
		"cluster.started", "cluster.stopped", "cluster.restarted":
		return c.cache.InvalidateCluster(ctx, event.InstanceID)

	case "node.failed", "node.recovered", "node.promoted":
		if clusterID, ok := event.Metadata["cluster_id"].(string); ok {
			return c.cache.InvalidateCluster(ctx, clusterID)
		}

	case "cluster.deleted":
		return c.cache.InvalidateCluster(ctx, event.InstanceID)
	}

	return nil
}

func (c *eventConsumer) Close() error {
	if c.reader != nil {
		c.logger.Info("closing kafka consumer")
		return c.reader.Close()
	}
	return nil
}
