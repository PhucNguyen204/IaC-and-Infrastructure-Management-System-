package collectors

// import (
// 	"context"
// 	"io"
// 	"strings"
// 	"testing"

// 	"github.com/docker/docker/api/types/container"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
// 	"go.uber.org/zap"
// )

// type MockDockerClient struct {
// 	mock.Mock
// }

// func (m *MockDockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
// 	args := m.Called(ctx, containerID, stream)
// 	return args.Get(0).(container.StatsResponseReader), args.Error(1)
// }

// type MockLogger struct{}

// func (m *MockLogger) Debug(msg string, fields ...zap.Field)   {}
// func (m *MockLogger) Info(msg string, fields ...zap.Field)    {}
// func (m *MockLogger) Warn(msg string, fields ...zap.Field)    {}
// func (m *MockLogger) Error(msg string, fields ...zap.Field)   {}
// func (m *MockLogger) Fatal(msg string, fields ...zap.Field)   {}
// func (m *MockLogger) With(fields ...zap.Field) logger.ILogger { return m }
// func (m *MockLogger) Sync() error                             { return nil }
// func (m *MockLogger) GetZapLogger() *zap.Logger               { return zap.NewNop() }

// func TestCollectContainerStats_Success(t *testing.T) {
// 	statsJSON := `{
// 		"cpu_stats": {
// 			"cpu_usage": {"total_usage": 1000000},
// 			"system_cpu_usage": 10000000
// 		},
// 		"precpu_stats": {
// 			"cpu_usage": {"total_usage": 500000},
// 			"system_cpu_usage": 5000000
// 		},
// 		"memory_stats": {
// 			"usage": 104857600,
// 			"limit": 1073741824,
// 			"max_usage": 104857600
// 		},
// 		"networks": {
// 			"eth0": {"rx_bytes": 1024, "tx_bytes": 2048}
// 		},
// 		"blkio_stats": {
// 			"io_service_bytes_recursive": [
// 				{"op": "Read", "value": 4096},
// 				{"op": "Write", "value": 8192}
// 			]
// 		}
// 	}`

// 	mockClient := new(MockDockerClient)
// 	mockLogger := &MockLogger{}

// 	statsReader := container.StatsResponseReader{
// 		Body:   io.NopCloser(strings.NewReader(statsJSON)),
// 		OSType: "linux",
// 	}

// 	mockClient.On("ContainerStats", mock.Anything, "test-container", false).Return(statsReader, nil)

// 	dc := &dockerCollector{
// 		client: nil,
// 		logger: mockLogger,
// 	}

// 	stats, err := dc.CollectContainerStats(context.Background(), "test-container")

// 	assert.NoError(t, err)
// 	assert.NotNil(t, stats)
// 	assert.Greater(t, stats.CPUPercent, 0.0)
// 	assert.Equal(t, int64(104857600), stats.MemoryUsed)
// 	assert.Equal(t, int64(1073741824), stats.MemoryLimit)
// 	assert.Greater(t, stats.MemoryPercent, 0.0)
// 	assert.Equal(t, int64(1024), stats.NetworkRx)
// 	assert.Equal(t, int64(2048), stats.NetworkTx)
// 	assert.Equal(t, int64(4096), stats.DiskRead)
// 	assert.Equal(t, int64(8192), stats.DiskWrite)
// }

// func TestCollectContainerStats_InvalidJSON(t *testing.T) {
// 	invalidJSON := `{invalid json`

// 	mockLogger := &MockLogger{}

// 	statsReader := container.StatsResponseReader{
// 		Body:   io.NopCloser(strings.NewReader(invalidJSON)),
// 		OSType: "linux",
// 	}

// 	mockClient := new(MockDockerClient)
// 	mockClient.On("ContainerStats", mock.Anything, "test-container", false).Return(statsReader, nil)

// 	dc := &dockerCollector{
// 		client: nil,
// 		logger: mockLogger,
// 	}

// 	_, err := dc.CollectContainerStats(context.Background(), "test-container")

// 	assert.Error(t, err)
// }
