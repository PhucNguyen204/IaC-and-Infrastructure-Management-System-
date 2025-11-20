package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/env"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type IElasticsearchClient interface {
	IndexLog(ctx context.Context, log LogEntry) error
	IndexMetric(ctx context.Context, metric MetricEntry) error
	QueryLogs(ctx context.Context, instanceID string, from, size int) ([]LogEntry, error)
	QueryMetrics(ctx context.Context, instanceID string, from, size int) ([]MetricEntry, error)
}

type LogEntry struct {
	InstanceID string                 `json:"instance_id"`
	UserID     string                 `json:"user_id"`
	Type       string                 `json:"type"`
	Action     string                 `json:"action"`
	Timestamp  time.Time              `json:"timestamp"`
	Message    string                 `json:"message"`
	Level      string                 `json:"level"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type MetricEntry struct {
	InstanceID    string                 `json:"instance_id"`
	Timestamp     time.Time              `json:"timestamp"`
	CPUPercent    float64                `json:"cpu_percent"`
	MemoryUsed    int64                  `json:"memory_used"`
	MemoryLimit   int64                  `json:"memory_limit"`
	MemoryPercent float64                `json:"memory_percent"`
	NetworkRx     int64                  `json:"network_rx"`
	NetworkTx     int64                  `json:"network_tx"`
	DiskRead      int64                  `json:"disk_read"`
	DiskWrite     int64                  `json:"disk_write"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type elasticsearchClient struct {
	client *elasticsearch.Client
	logger logger.ILogger
}

func NewElasticsearchClient(env env.ElasticsearchEnv, logger logger.ILogger) (IElasticsearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: env.Addresses,
		Username:  env.Username,
		Password:  env.Password,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &elasticsearchClient{
		client: client,
		logger: logger,
	}, nil
}

func (es *elasticsearchClient) IndexLog(ctx context.Context, log LogEntry) error {
	log.Timestamp = time.Now()
	indexName := fmt.Sprintf("infrastructure-logs-%s", time.Now().Format("2006.01.02"))

	data, err := json.Marshal(log)
	if err != nil {
		es.logger.Error("failed to marshal log entry", zap.Error(err))
		return err
	}

	req := esapi.IndexRequest{
		Index:   indexName,
		Body:    bytes.NewReader(data),
		Refresh: "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		es.logger.Error("failed to index log", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		es.logger.Error("elasticsearch error", zap.String("status", res.Status()))
		return fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	es.logger.Debug("log indexed successfully", zap.String("index", indexName))
	return nil
}

func (es *elasticsearchClient) IndexMetric(ctx context.Context, metric MetricEntry) error {
	metric.Timestamp = time.Now()
	indexName := fmt.Sprintf("infrastructure-metrics-%s", time.Now().Format("2006.01.02"))

	data, err := json.Marshal(metric)
	if err != nil {
		es.logger.Error("failed to marshal metric entry", zap.Error(err))
		return err
	}

	req := esapi.IndexRequest{
		Index:   indexName,
		Body:    bytes.NewReader(data),
		Refresh: "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		es.logger.Error("failed to index metric", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		es.logger.Error("elasticsearch error", zap.String("status", res.Status()))
		return fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	es.logger.Debug("metric indexed successfully", zap.String("index", indexName))
	return nil
}

func (es *elasticsearchClient) QueryLogs(ctx context.Context, instanceID string, from, size int) ([]LogEntry, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"instance_id": instanceID,
			},
		},
		"from": from,
		"size": size,
		"sort": []map[string]interface{}{
			{"timestamp": map[string]string{"order": "desc"}},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := es.client.Search(
		es.client.Search.WithContext(ctx),
		es.client.Search.WithIndex("infrastructure-logs-*"),
		es.client.Search.WithBody(bytes.NewReader(queryBytes)),
	)
	if err != nil {
		es.logger.Error("failed to search logs", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	logs := []LogEntry{}
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		logBytes, _ := json.Marshal(source)
		var log LogEntry
		json.Unmarshal(logBytes, &log)
		logs = append(logs, log)
	}

	return logs, nil
}

func (es *elasticsearchClient) QueryMetrics(ctx context.Context, instanceID string, from, size int) ([]MetricEntry, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"instance_id": instanceID,
			},
		},
		"from": from,
		"size": size,
		"sort": []map[string]interface{}{
			{"timestamp": map[string]string{"order": "desc"}},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := es.client.Search(
		es.client.Search.WithContext(ctx),
		es.client.Search.WithIndex("infrastructure-metrics-*"),
		es.client.Search.WithBody(bytes.NewReader(queryBytes)),
	)
	if err != nil {
		es.logger.Error("failed to search metrics", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	metrics := []MetricEntry{}
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		metricBytes, _ := json.Marshal(source)
		var metric MetricEntry
		json.Unmarshal(metricBytes, &metric)
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

