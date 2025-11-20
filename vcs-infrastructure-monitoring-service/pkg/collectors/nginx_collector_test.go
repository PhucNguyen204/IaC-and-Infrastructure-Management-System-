package collectors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type mockNginxLogger struct{}

func (m *mockNginxLogger) Debug(msg string, fields ...zap.Field)   {}
func (m *mockNginxLogger) Info(msg string, fields ...zap.Field)    {}
func (m *mockNginxLogger) Warn(msg string, fields ...zap.Field)    {}
func (m *mockNginxLogger) Error(msg string, fields ...zap.Field)   {}
func (m *mockNginxLogger) Fatal(msg string, fields ...zap.Field)   {}
func (m *mockNginxLogger) With(fields ...zap.Field) logger.ILogger { return m }
func (m *mockNginxLogger) Sync() error                             { return nil }
func (m *mockNginxLogger) GetZapLogger() *zap.Logger               { return zap.NewNop() }

func TestNewNginxCollector(t *testing.T) {
	mockLogger := &mockNginxLogger{}
	collector := NewNginxCollector(mockLogger)
	assert.NotNil(t, collector)
}

func TestCollectMetrics_Success(t *testing.T) {
	nginxStatus := `Active connections: 42
server accepts handled requests
 1000 1000 5000
Reading: 5 Writing: 10 Waiting: 27`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(nginxStatus))
	}))
	defer server.Close()

	mockLogger := &mockNginxLogger{}
	collector := NewNginxCollector(mockLogger)

	metrics, err := collector.CollectMetrics(context.Background(), server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(42), metrics.ActiveConnections)
	assert.Equal(t, int64(1000), metrics.Accepts)
	assert.Equal(t, int64(1000), metrics.Handled)
	assert.Equal(t, int64(5000), metrics.Requests)
	assert.Equal(t, int64(5), metrics.Reading)
	assert.Equal(t, int64(10), metrics.Writing)
	assert.Equal(t, int64(27), metrics.Waiting)
}

func TestCollectMetrics_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	mockLogger := &mockNginxLogger{}
	collector := NewNginxCollector(mockLogger)

	metrics, err := collector.CollectMetrics(context.Background(), server.URL)

	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func TestCollectMetrics_InvalidURL(t *testing.T) {
	mockLogger := &mockNginxLogger{}
	collector := NewNginxCollector(mockLogger)

	metrics, err := collector.CollectMetrics(context.Background(), "http://invalid-url-that-does-not-exist:99999")

	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func TestParseNginxStatus(t *testing.T) {
	mockLogger := &mockNginxLogger{}
	collector := &nginxCollector{logger: mockLogger}

	status := `Active connections: 100
server accepts handled requests
 10000 9500 50000
Reading: 10 Writing: 20 Waiting: 70`

	metrics, err := collector.parseNginxStatus(status)

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(100), metrics.ActiveConnections)
	assert.Equal(t, int64(10000), metrics.Accepts)
	assert.Equal(t, int64(9500), metrics.Handled)
	assert.Equal(t, int64(50000), metrics.Requests)
	assert.Equal(t, int64(10), metrics.Reading)
	assert.Equal(t, int64(20), metrics.Writing)
	assert.Equal(t, int64(70), metrics.Waiting)
}
