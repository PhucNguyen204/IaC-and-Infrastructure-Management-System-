package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/infrastructures/elasticsearch"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/env"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type MockElasticsearchClient struct {
	mock.Mock
}

func (m *MockElasticsearchClient) IndexLog(ctx context.Context, log elasticsearch.LogEntry) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockElasticsearchClient) IndexMetric(ctx context.Context, metric elasticsearch.MetricEntry) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockElasticsearchClient) QueryLogs(ctx context.Context, instanceID string, from, size int) ([]elasticsearch.LogEntry, error) {
	args := m.Called(ctx, instanceID, from, size)
	return args.Get(0).([]elasticsearch.LogEntry), args.Error(1)
}

func (m *MockElasticsearchClient) QueryMetrics(ctx context.Context, instanceID string, from, size int) ([]elasticsearch.MetricEntry, error) {
	args := m.Called(ctx, instanceID, from, size)
	return args.Get(0).([]elasticsearch.MetricEntry), args.Error(1)
}

type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...zap.Field)                  {}
func (m *MockLogger) Info(msg string, fields ...zap.Field)                   {}
func (m *MockLogger) Warn(msg string, fields ...zap.Field)                   {}
func (m *MockLogger) Error(msg string, fields ...zap.Field)                  {}
func (m *MockLogger) Fatal(msg string, fields ...zap.Field)                  {}
func (m *MockLogger) With(fields ...zap.Field) logger.ILogger                { return m }
func (m *MockLogger) Sync() error                                            { return nil }
func (m *MockLogger) GetZapLogger() *zap.Logger                              { return zap.NewNop() }

func TestNewKafkaConsumer(t *testing.T) {
	mockES := new(MockElasticsearchClient)
	mockLogger := &MockLogger{}

	kafkaEnv := env.KafkaEnv{
		Brokers: []string{"localhost:9092"},
		GroupID: "test-group",
		Topics:  []string{"test-topic"},
	}

	consumer := NewKafkaConsumer(kafkaEnv, mockES, mockLogger)

	assert.NotNil(t, consumer)
}

func TestKafkaConsumer_Start(t *testing.T) {
	mockES := new(MockElasticsearchClient)
	mockLogger := &MockLogger{}

	kafkaEnv := env.KafkaEnv{
		Brokers: []string{"localhost:9092"},
		GroupID: "test-group",
		Topics:  []string{"test-topic"},
	}

	consumer := NewKafkaConsumer(kafkaEnv, mockES, mockLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := consumer.Start(ctx)
	assert.NoError(t, err)

	<-ctx.Done()

	err = consumer.Close()
	assert.NoError(t, err)
}

func TestInfrastructureEvent_Structure(t *testing.T) {
	event := InfrastructureEvent{
		InstanceID: "test-id",
		UserID:     "user-id",
		Type:       "postgres",
		Action:     "created",
		Timestamp:  time.Now().Format(time.RFC3339),
		Metadata: map[string]interface{}{
			"version": "15",
			"port":    5432,
		},
	}

	assert.Equal(t, "test-id", event.InstanceID)
	assert.Equal(t, "user-id", event.UserID)
	assert.Equal(t, "postgres", event.Type)
	assert.Equal(t, "created", event.Action)
	assert.NotEmpty(t, event.Timestamp)
	assert.NotNil(t, event.Metadata)
}

