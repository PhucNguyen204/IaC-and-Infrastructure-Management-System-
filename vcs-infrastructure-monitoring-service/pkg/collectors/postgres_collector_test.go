package collectors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...zap.Field)   {}
func (m *mockLogger) Info(msg string, fields ...zap.Field)    {}
func (m *mockLogger) Warn(msg string, fields ...zap.Field)    {}
func (m *mockLogger) Error(msg string, fields ...zap.Field)   {}
func (m *mockLogger) Fatal(msg string, fields ...zap.Field)   {}
func (m *mockLogger) With(fields ...zap.Field) logger.ILogger { return m }
func (m *mockLogger) Sync() error                             { return nil }
func (m *mockLogger) GetZapLogger() *zap.Logger               { return zap.NewNop() }

func TestNewPostgreSQLCollector(t *testing.T) {
	mockLogger := &mockLogger{}
	collector := NewPostgreSQLCollector(mockLogger)
	assert.NotNil(t, collector)
}

func TestCollectMetrics_ConnectionError(t *testing.T) {
	mockLogger := &mockLogger{}
	collector := NewPostgreSQLCollector(mockLogger)

	metrics, err := collector.CollectMetrics(
		context.Background(),
		"invalid-host",
		5432,
		"user",
		"password",
		"database",
	)

	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func TestPostgreSQLMetrics_Structure(t *testing.T) {
	metrics := &PostgreSQLMetrics{
		ActiveConnections: 10,
		TotalConnections:  50,
		Transactions:      1000,
		Commits:           900,
		Rollbacks:         100,
		BlocksRead:        5000,
		BlocksHit:         45000,
		TuplesReturned:    10000,
		TuplesFetched:     8000,
		TuplesInserted:    1000,
		TuplesUpdated:     500,
		TuplesDeleted:     100,
		ReplicationLag:    0,
	}

	assert.Equal(t, int64(10), metrics.ActiveConnections)
	assert.Equal(t, int64(50), metrics.TotalConnections)
	assert.Equal(t, int64(1000), metrics.Transactions)
	assert.Equal(t, int64(900), metrics.Commits)
	assert.Equal(t, int64(100), metrics.Rollbacks)
}
