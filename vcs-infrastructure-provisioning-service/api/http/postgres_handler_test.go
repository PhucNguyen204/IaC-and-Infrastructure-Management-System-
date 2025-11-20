package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
)

type MockPostgreSQLService struct {
	mock.Mock
}

func (m *MockPostgreSQLService) CreatePostgreSQL(ctx context.Context, userID string, req dto.CreatePostgreSQLRequest) (*dto.PostgreSQLInfoResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PostgreSQLInfoResponse), args.Error(1)
}

func (m *MockPostgreSQLService) StartPostgreSQL(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostgreSQLService) StopPostgreSQL(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostgreSQLService) RestartPostgreSQL(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostgreSQLService) DeletePostgreSQL(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostgreSQLService) GetPostgreSQLInfo(ctx context.Context, id string) (*dto.PostgreSQLInfoResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PostgreSQLInfoResponse), args.Error(1)
}

func (m *MockPostgreSQLService) GetPostgreSQLLogs(ctx context.Context, id string, tail string) (string, error) {
	args := m.Called(ctx, id, tail)
	return args.String(0), args.Error(1)
}

func (m *MockPostgreSQLService) GetPostgreSQLStats(ctx context.Context, id string) (*dto.PostgreSQLStatsResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PostgreSQLStatsResponse), args.Error(1)
}

func (m *MockPostgreSQLService) BackupPostgreSQL(ctx context.Context, id string, req dto.BackupPostgreSQLRequest) (*dto.BackupPostgreSQLResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BackupPostgreSQLResponse), args.Error(1)
}

func (m *MockPostgreSQLService) RestorePostgreSQL(ctx context.Context, id string, req dto.RestorePostgreSQLRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func TestCreatePostgreSQL_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPostgreSQLService)
	handler := NewPostgreSQLHandler(mockService)

	expectedResponse := &dto.PostgreSQLInfoResponse{
		ID:           "test-id",
		Name:         "test-pg",
		Status:       "running",
		ContainerID:  "container-123",
		Version:      "15",
		Port:         5432,
		DatabaseName: "testdb",
		Username:     "testuser",
	}

	mockService.On("CreatePostgreSQL", mock.Anything, "user-123", mock.AnythingOfType("dto.CreatePostgreSQLRequest")).
		Return(expectedResponse, nil)

	router := gin.New()
	router.POST("/postgres/single", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		handler.CreatePostgreSQL(c)
	})

	reqBody := dto.CreatePostgreSQLRequest{
		Name:         "test-pg",
		Version:      "15",
		Port:         5432,
		DatabaseName: "testdb",
		Username:     "testuser",
		Password:     "testpassword",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/postgres/single", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "CREATED", response.Code)

	mockService.AssertExpectations(t)
}

func TestGetPostgreSQLInfo_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPostgreSQLService)
	handler := NewPostgreSQLHandler(mockService)

	expectedResponse := &dto.PostgreSQLInfoResponse{
		ID:           "test-id",
		Name:         "test-pg",
		Status:       "running",
		ContainerID:  "container-123",
		Version:      "15",
		Port:         5432,
		DatabaseName: "testdb",
		Username:     "testuser",
	}

	mockService.On("GetPostgreSQLInfo", mock.Anything, "test-id").Return(expectedResponse, nil)

	router := gin.New()
	router.GET("/postgres/single/:id", handler.GetPostgreSQLInfo)

	req, _ := http.NewRequest("GET", "/postgres/single/test-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "SUCCESS", response.Code)

	mockService.AssertExpectations(t)
}

func TestGetPostgreSQLLogs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPostgreSQLService)
	handler := NewPostgreSQLHandler(mockService)

	mockService.On("GetPostgreSQLLogs", mock.Anything, "test-id", "50").
		Return("log line 1\nlog line 2", nil)

	router := gin.New()
	router.GET("/postgres/single/:id/logs", handler.GetPostgreSQLLogs)

	req, _ := http.NewRequest("GET", "/postgres/single/test-id/logs?tail=50", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "SUCCESS", response.Code)

	data := response.Data.(map[string]interface{})
	assert.Contains(t, data["logs"].(string), "log line")

	mockService.AssertExpectations(t)
}

func TestGetPostgreSQLStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPostgreSQLService)
	handler := NewPostgreSQLHandler(mockService)

	statsResp := &dto.PostgreSQLStatsResponse{
		CPUPercent:    12.5,
		MemoryUsed:    100,
		MemoryLimit:   200,
		MemoryPercent: 50.0,
		NetworkRx:     1024,
		NetworkTx:     2048,
		DiskRead:      4096,
		DiskWrite:     8192,
	}

	mockService.On("GetPostgreSQLStats", mock.Anything, "test-id").
		Return(statsResp, nil)

	router := gin.New()
	router.GET("/postgres/single/:id/stats", handler.GetPostgreSQLStats)

	req, _ := http.NewRequest("GET", "/postgres/single/test-id/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "SUCCESS", response.Code)

	statsData := response.Data.(map[string]interface{})
	assert.Equal(t, statsResp.CPUPercent, statsData["cpu_percent"])

	mockService.AssertExpectations(t)
}

func TestStartPostgreSQL_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPostgreSQLService)
	handler := NewPostgreSQLHandler(mockService)

	mockService.On("StartPostgreSQL", mock.Anything, "test-id").Return(nil)

	router := gin.New()
	router.POST("/postgres/single/:id/start", handler.StartPostgreSQL)

	req, _ := http.NewRequest("POST", "/postgres/single/test-id/start", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "SUCCESS", response.Code)

	mockService.AssertExpectations(t)
}

func TestStopPostgreSQL_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPostgreSQLService)
	handler := NewPostgreSQLHandler(mockService)

	mockService.On("StopPostgreSQL", mock.Anything, "test-id").Return(nil)

	router := gin.New()
	router.POST("/postgres/single/:id/stop", handler.StopPostgreSQL)

	req, _ := http.NewRequest("POST", "/postgres/single/test-id/stop", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestDeletePostgreSQL_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPostgreSQLService)
	handler := NewPostgreSQLHandler(mockService)

	mockService.On("DeletePostgreSQL", mock.Anything, "test-id").Return(nil)

	router := gin.New()
	router.DELETE("/postgres/single/:id", handler.DeletePostgreSQL)

	req, _ := http.NewRequest("DELETE", "/postgres/single/test-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "SUCCESS", response.Code)

	mockService.AssertExpectations(t)
}
