package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PhucNguyen204/vcs-healthcheck-service/dto"
	"github.com/PhucNguyen204/vcs-healthcheck-service/pkg/logger"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.uber.org/zap"
)

type IElasticsearchClient interface {
	// Index operations
	IndexEvent(ctx context.Context, event dto.LifecycleEvent) error
	IndexMetrics(ctx context.Context, metrics dto.ContainerMetrics) error
	IndexHealthCheck(ctx context.Context, result dto.HealthCheckResult) error
	IndexUptimeEvent(ctx context.Context, event dto.LifecycleEvent) error

	// Health check
	Ping(ctx context.Context) error
}

type elasticsearchClient struct {
	client *elasticsearch.Client
	logger logger.ILogger
}

type Config struct {
	Addresses []string
	Username  string
	Password  string
}

func NewElasticsearchClient(config Config, logger logger.ILogger) (IElasticsearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: config.Addresses,
		Username:  config.Username,
		Password:  config.Password,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	es := &elasticsearchClient{
		client: client,
		logger: logger,
	}

	// Test connection
	if err := es.Ping(context.Background()); err != nil {
		logger.Warn("elasticsearch not available, will retry on operations", zap.Error(err))
	} else {
		logger.Info("elasticsearch connection established", zap.Strings("addresses", config.Addresses))
	}

	return es, nil
}

func (es *elasticsearchClient) Ping(ctx context.Context) error {
	res, err := es.client.Ping(es.client.Ping.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch ping failed: %s", res.Status())
	}
	return nil
}

func (es *elasticsearchClient) IndexEvent(ctx context.Context, event dto.LifecycleEvent) error {
	indexName := fmt.Sprintf("infra-events-%s", time.Now().Format("2006.01.02"))

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req := esapi.IndexRequest{
		Index:   indexName,
		Body:    bytes.NewReader(data),
		Refresh: "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		es.logger.Error("failed to index event", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		es.logger.Error("elasticsearch error indexing event", zap.String("status", res.Status()))
		return fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	es.logger.Debug("event indexed successfully",
		zap.String("index", indexName),
		zap.String("infrastructure_id", event.InfrastructureID),
		zap.String("action", event.Action))

	return nil
}

func (es *elasticsearchClient) IndexMetrics(ctx context.Context, metrics dto.ContainerMetrics) error {
	indexName := fmt.Sprintf("infra-metrics-%s", time.Now().Format("2006.01.02"))

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req := esapi.IndexRequest{
		Index:   indexName,
		Body:    bytes.NewReader(data),
		Refresh: "false", // Don't wait for refresh for metrics
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		es.logger.Error("failed to index metrics", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		es.logger.Error("elasticsearch error indexing metrics", zap.String("status", res.Status()))
		return fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	es.logger.Debug("metrics indexed successfully",
		zap.String("index", indexName),
		zap.String("container_name", metrics.ContainerName))

	return nil
}

func (es *elasticsearchClient) IndexHealthCheck(ctx context.Context, result dto.HealthCheckResult) error {
	indexName := fmt.Sprintf("infra-health-%s", time.Now().Format("2006.01.02"))

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal health check: %w", err)
	}

	req := esapi.IndexRequest{
		Index:   indexName,
		Body:    bytes.NewReader(data),
		Refresh: "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		es.logger.Error("failed to index health check", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		es.logger.Error("elasticsearch error indexing health check", zap.String("status", res.Status()))
		return fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	es.logger.Debug("health check indexed successfully",
		zap.String("index", indexName),
		zap.String("container_name", result.ContainerName),
		zap.String("status", result.Status))

	return nil
}

// IndexUptimeEvent stores uptime events in infrastructure-uptime-* index
// This is the index that Monitoring service queries for uptime calculations
func (es *elasticsearchClient) IndexUptimeEvent(ctx context.Context, event dto.LifecycleEvent) error {
	// Use the same index pattern as monitoring service
	indexName := fmt.Sprintf("infrastructure-uptime-%s", time.Now().Format("2006.01.02"))

	// For uptime tracking, we need to use infrastructure_id (infra_id from metadata)
	// because monitoring service queries by infrastructure_id, not cluster_id
	instanceID := event.InfrastructureID
	if event.Metadata != nil {
		if infraID, ok := event.Metadata["infra_id"].(string); ok && infraID != "" {
			instanceID = infraID
		}
	}

	// Create uptime event structure compatible with monitoring service
	uptimeEvent := map[string]interface{}{
		"instance_id":     instanceID, // Use infrastructure_id for monitoring queries
		"instance_name":   event.Metadata["cluster_name"],
		"user_id":         event.UserID,
		"type":            event.Type,
		"action":          event.Action,
		"status":          event.Status,
		"previous_status": event.PreviousStatus,
		"timestamp":       event.Timestamp,
		"message":         fmt.Sprintf("Infrastructure %s: %s", event.Action, event.Status),
		"metadata":        event.Metadata,
	}

	data, err := json.Marshal(uptimeEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal uptime event: %w", err)
	}

	req := esapi.IndexRequest{
		Index:   indexName,
		Body:    bytes.NewReader(data),
		Refresh: "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		es.logger.Error("failed to index uptime event", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		es.logger.Error("elasticsearch error indexing uptime event", zap.String("status", res.Status()))
		return fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	es.logger.Info("uptime event indexed successfully",
		zap.String("index", indexName),
		zap.String("instance_id", instanceID),
		zap.String("action", event.Action),
		zap.String("status", event.Status))

	return nil
}
