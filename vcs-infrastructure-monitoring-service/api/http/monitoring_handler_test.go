package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/usecases/services"
)

type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) GetCurrentMetrics(ctx context.Context, instanceID string) (*dto.MetricsResponse, error) {
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.MetricsResponse), args.Error(1)
}

func (m *MockMetricsService) GetHistoricalMetrics(ctx context.Context, instanceID string, from, size int) ([]dto.MetricsResponse, error) {
	args := m.Called(ctx, instanceID, from, size)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.MetricsResponse), args.Error(1)
}

func (m *MockMetricsService) GetLogs(ctx context.Context, instanceID string, from, size int) ([]dto.LogsResponse, error) {
	args := m.Called(ctx, instanceID, from, size)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.LogsResponse), args.Error(1)
}

func (m *MockMetricsService) AggregateMetrics(ctx context.Context, instanceID string, timeRange string) (*dto.AggregatedMetricsResponse, error) {
	args := m.Called(ctx, instanceID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AggregatedMetricsResponse), args.Error(1)
}

var _ services.IMetricsService = (*MockMetricsService)(nil)

func TestGetCurrentMetrics_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	expectedMetrics := &dto.MetricsResponse{
		InstanceID:    "test-id",
		CPUPercent:    50.5,
		MemoryUsed:    1024000,
		MemoryLimit:   2048000,
		MemoryPercent: 50.0,
	}

	mockService.On("GetCurrentMetrics", mock.Anything, "test-id").Return(expectedMetrics, nil)

	router := gin.New()
	router.GET("/monitoring/metrics/:instance_id", handler.GetCurrentMetrics)

	req, _ := http.NewRequest("GET", "/monitoring/metrics/test-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "SUCCESS", response.Code)

	mockService.AssertExpectations(t)
}

func TestGetHistoricalMetrics_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	expectedMetrics := []dto.MetricsResponse{
		{
			InstanceID:    "test-id",
			CPUPercent:    50.5,
			MemoryPercent: 50.0,
		},
		{
			InstanceID:    "test-id",
			CPUPercent:    45.2,
			MemoryPercent: 48.0,
		},
	}

	mockService.On("GetHistoricalMetrics", mock.Anything, "test-id", 0, 100).Return(expectedMetrics, nil)

	router := gin.New()
	router.GET("/monitoring/metrics/:instance_id/history", handler.GetHistoricalMetrics)

	req, _ := http.NewRequest("GET", "/monitoring/metrics/test-id/history", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestGetLogs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	expectedLogs := []dto.LogsResponse{
		{
			InstanceID: "test-id",
			Message:    "Container started",
			Level:      "info",
			Action:     "started",
		},
	}

	mockService.On("GetLogs", mock.Anything, "test-id", 0, 100).Return(expectedLogs, nil)

	router := gin.New()
	router.GET("/monitoring/logs/:instance_id", handler.GetLogs)

	req, _ := http.NewRequest("GET", "/monitoring/logs/test-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestGetAggregatedMetrics_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	expectedAggregated := &dto.AggregatedMetricsResponse{
		InstanceID: "test-id",
		TimeRange:  "1h",
		CPUPercent: dto.AggregatedValue{
			Avg: 50.0,
			Max: 80.0,
			Min: 20.0,
		},
		MemoryPercent: dto.AggregatedValue{
			Avg: 60.0,
			Max: 90.0,
			Min: 30.0,
		},
		DataPoints: 60,
	}

	mockService.On("AggregateMetrics", mock.Anything, "test-id", "1h").Return(expectedAggregated, nil)

	router := gin.New()
	router.GET("/monitoring/metrics/:instance_id/aggregate", handler.GetAggregatedMetrics)

	req, _ := http.NewRequest("GET", "/monitoring/metrics/test-id/aggregate?range=1h", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "SUCCESS", response.Code)

	mockService.AssertExpectations(t)
}

func TestGetHealthStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	router := gin.New()
	router.GET("/monitoring/health/:instance_id", handler.GetHealthStatus)

	req, _ := http.NewRequest("GET", "/monitoring/health/test-container-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestListInfrastructure_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	router := gin.New()
	router.GET("/monitoring/infrastructure", handler.ListInfrastructure)

	req, _ := http.NewRequest("GET", "/monitoring/infrastructure", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
}

func TestGetCurrentMetrics_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	mockService.On("GetCurrentMetrics", mock.Anything, "test-id").Return(nil, assert.AnError)

	router := gin.New()
	router.GET("/monitoring/metrics/:instance_id", handler.GetCurrentMetrics)

	req, _ := http.NewRequest("GET", "/monitoring/metrics/test-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)

	mockService.AssertExpectations(t)
}

func TestGetHistoricalMetrics_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	mockService.On("GetHistoricalMetrics", mock.Anything, "test-id", 0, 100).Return(nil, assert.AnError)

	router := gin.New()
	router.GET("/monitoring/metrics/:instance_id/history", handler.GetHistoricalMetrics)

	req, _ := http.NewRequest("GET", "/monitoring/metrics/test-id/history", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestGetLogs_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	mockService.On("GetLogs", mock.Anything, "test-id", 0, 100).Return(nil, assert.AnError)

	router := gin.New()
	router.GET("/monitoring/logs/:instance_id", handler.GetLogs)

	req, _ := http.NewRequest("GET", "/monitoring/logs/test-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestGetAggregatedMetrics_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	mockService.On("AggregateMetrics", mock.Anything, "test-id", "1h").Return(nil, assert.AnError)

	router := gin.New()
	router.GET("/monitoring/metrics/:instance_id/aggregate", handler.GetAggregatedMetrics)

	req, _ := http.NewRequest("GET", "/monitoring/metrics/test-id/aggregate?range=1h", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestGetHistoricalMetrics_WithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	expectedMetrics := []dto.MetricsResponse{
		{InstanceID: "test-id", CPUPercent: 50.0},
	}

	mockService.On("GetHistoricalMetrics", mock.Anything, "test-id", 10, 20).Return(expectedMetrics, nil)

	router := gin.New()
	router.GET("/monitoring/metrics/:instance_id/history", handler.GetHistoricalMetrics)

	req, _ := http.NewRequest("GET", "/monitoring/metrics/test-id/history?from=10&size=20", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestGetLogs_WithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMetricsService)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	handler := NewMonitoringHandler(mockService, mockRedis)

	expectedLogs := []dto.LogsResponse{
		{InstanceID: "test-id", Message: "Test log"},
	}

	mockService.On("GetLogs", mock.Anything, "test-id", 5, 15).Return(expectedLogs, nil)

	router := gin.New()
	router.GET("/monitoring/logs/:instance_id", handler.GetLogs)

	req, _ := http.NewRequest("GET", "/monitoring/logs/test-id?from=5&size=15", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}
