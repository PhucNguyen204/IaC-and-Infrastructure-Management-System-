package main

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"

// 	"github.com/docker/docker/api/types"
// 	"github.com/docker/docker/api/types/container"
// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	httpHandler "github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/api/http"
// 	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
// 	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
// 	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/docker"
// 	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/kafka"
// 	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/env"
// 	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
// 	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/repositories"
// 	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// )

// type MockKafkaProducer struct{}

// func (m *MockKafkaProducer) PublishEvent(ctx context.Context, event kafka.InfrastructureEvent) error {
// 	return nil
// }

// func (m *MockKafkaProducer) Close() error {
// 	return nil
// }

// type MockDockerService struct{}

// func (m *MockDockerService) CreateContainer(ctx context.Context, config docker.ContainerConfig) (string, error) {
// 	return "mock-container-id", nil
// }

// func (m *MockDockerService) StartContainer(ctx context.Context, containerID string) error {
// 	return nil
// }

// func (m *MockDockerService) StopContainer(ctx context.Context, containerID string) error {
// 	return nil
// }

// func (m *MockDockerService) RestartContainer(ctx context.Context, containerID string) error {
// 	return nil
// }

// func (m *MockDockerService) RemoveContainer(ctx context.Context, containerID string) error {
// 	return nil
// }

// func (m *MockDockerService) GetContainerStats(ctx context.Context, containerID string) (container.StatsResponseReader, error) {
// 	return container.StatsResponseReader{}, nil
// }

// func (m *MockDockerService) GetContainerLogs(ctx context.Context, containerID string, tail string) (string, error) {
// 	return "mock logs", nil
// }

// func (m *MockDockerService) ExecCommand(ctx context.Context, containerID string, cmd []string) (string, error) {
// 	return "mock output", nil
// }

// func (m *MockDockerService) InspectContainer(ctx context.Context, containerID string) (*types.ContainerJSON, error) {
// 	return nil, nil
// }

// func (m *MockDockerService) CreateNetwork(ctx context.Context, networkName string) (string, error) {
// 	return "mock-network-id", nil
// }

// func (m *MockDockerService) RemoveNetwork(ctx context.Context, networkID string) error {
// 	return nil
// }

// func (m *MockDockerService) CreateVolume(ctx context.Context, volumeName string) error {
// 	return nil
// }

// func (m *MockDockerService) RemoveVolume(ctx context.Context, volumeName string) error {
// 	return nil
// }

// func setupTestRouter() (*gin.Engine, *gorm.DB) {
// 	gin.SetMode(gin.TestMode)

// 	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
// 	db.AutoMigrate(
// 		&entities.Infrastructure{},
// 		&entities.PostgreSQLInstance{},
// 		&entities.NginxInstance{},
// 	)

// 	loggerEnv := env.LoggerEnv{
// 		Level:      "error",
// 		FilePath:   "./logs/test.log",
// 		MaxSize:    10,
// 		MaxAge:     1,
// 		MaxBackups: 1,
// 	}
// 	testLogger, _ := logger.LoadLogger(loggerEnv)

// 	mockDockerSvc := &MockDockerService{}
// 	mockKafkaProducer := &MockKafkaProducer{}

// 	infraRepo := repositories.NewInfrastructureRepository(db)
// 	pgRepo := repositories.NewPostgreSQLRepository(db)
// 	nginxRepo := repositories.NewNginxRepository(db)

// 	pgService := services.NewPostgreSQLService(infraRepo, pgRepo, mockDockerSvc, mockKafkaProducer, testLogger)
// 	nginxService := services.NewNginxService(infraRepo, nginxRepo, mockDockerSvc, mockKafkaProducer, testLogger)

// 	pgHandler := httpHandler.NewPostgreSQLHandler(pgService)
// 	nginxHandler := httpHandler.NewNginxHandler(nginxService)

// 	router := gin.New()
// 	router.GET("/health", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{"status": "ok"})
// 	})

// 	apiV1 := router.Group("/api/v1", func(c *gin.Context) {
// 		c.Set("user_id", "test-user-123")
// 		c.Next()
// 	})
// 	pgHandler.RegisterRoutes(apiV1)
// 	nginxHandler.RegisterRoutes(apiV1)

// 	return router, db
// }

// func TestHealthEndpoint(t *testing.T) {
// 	router, _ := setupTestRouter()

// 	req, _ := http.NewRequest("GET", "/health", nil)
// 	w := httptest.NewRecorder()

// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Contains(t, w.Body.String(), "ok")
// }

// func TestPostgreSQLFullWorkflow(t *testing.T) {
// 	router, db := setupTestRouter()

// 	createReq := dto.CreatePostgreSQLRequest{
// 		Name:         "test-postgres",
// 		Version:      "15",
// 		Port:         5432,
// 		DatabaseName: "testdb",
// 		Username:     "testuser",
// 		Password:     "testpassword123",
// 		CPULimit:     1000000000,
// 		MemoryLimit:  536870912,
// 	}
// 	bodyBytes, _ := json.Marshal(createReq)

// 	req, _ := http.NewRequest("POST", "/api/v1/postgres/single", bytes.NewBuffer(bodyBytes))
// 	req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()

// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusCreated, w.Code)

// 	var createResponse dto.APIResponse
// 	json.Unmarshal(w.Body.Bytes(), &createResponse)
// 	assert.True(t, createResponse.Success)
	
// 	responseData, _ := json.Marshal(createResponse.Data)
// 	var pgInfo dto.PostgreSQLInfoResponse
// 	json.Unmarshal(responseData, &pgInfo)
// 	instanceID := pgInfo.ID

// 	time.Sleep(100 * time.Millisecond)

// 	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/postgres/single/%s", instanceID), nil)
// 	getW := httptest.NewRecorder()
// 	router.ServeHTTP(getW, getReq)
// 	assert.Equal(t, http.StatusOK, getW.Code)

// 	stopReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/postgres/single/%s/stop", instanceID), nil)
// 	stopW := httptest.NewRecorder()
// 	router.ServeHTTP(stopW, stopReq)
// 	assert.Equal(t, http.StatusOK, stopW.Code)

// 	var infra entities.Infrastructure
// 	db.First(&infra, "id = ?", instanceID)
// 	assert.Equal(t, entities.StatusStopped, infra.Status)

// 	startReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/postgres/single/%s/start", instanceID), nil)
// 	startW := httptest.NewRecorder()
// 	router.ServeHTTP(startW, startReq)
// 	assert.Equal(t, http.StatusOK, startW.Code)

// 	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/postgres/single/%s", instanceID), nil)
// 	deleteW := httptest.NewRecorder()
// 	router.ServeHTTP(deleteW, deleteReq)
// 	assert.Equal(t, http.StatusOK, deleteW.Code)
// }

// func TestNginxFullWorkflow(t *testing.T) {
// 	router, db := setupTestRouter()

// 	createReq := dto.CreateNginxRequest{
// 		Name:   "test-nginx",
// 		Port:   8080,
// 		Config: "server { listen 80; }",
// 	}
// 	bodyBytes, _ := json.Marshal(createReq)

// 	req, _ := http.NewRequest("POST", "/api/v1/nginx", bytes.NewBuffer(bodyBytes))
// 	req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()

// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusCreated, w.Code)

// 	var createResponse dto.APIResponse
// 	json.Unmarshal(w.Body.Bytes(), &createResponse)
// 	assert.True(t, createResponse.Success)

// 	responseData, _ := json.Marshal(createResponse.Data)
// 	var nginxInfo dto.NginxInfoResponse
// 	json.Unmarshal(responseData, &nginxInfo)
// 	instanceID := nginxInfo.ID

// 	updateConfigReq := dto.UpdateNginxConfigRequest{
// 		Config: "server { listen 80; server_name example.com; }",
// 	}
// 	updateBytes, _ := json.Marshal(updateConfigReq)

// 	updateReq, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/nginx/%s/config", instanceID), bytes.NewBuffer(updateBytes))
// 	updateReq.Header.Set("Content-Type", "application/json")
// 	updateW := httptest.NewRecorder()
// 	router.ServeHTTP(updateW, updateReq)
// 	assert.Equal(t, http.StatusOK, updateW.Code)

// 	var nginxInstance entities.NginxInstance
// 	db.Preload("Infrastructure").First(&nginxInstance, "infrastructure_id = ?", instanceID)
// 	assert.Equal(t, "server { listen 80; server_name example.com; }", nginxInstance.Config)

// 	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/nginx/%s", instanceID), nil)
// 	deleteW := httptest.NewRecorder()
// 	router.ServeHTTP(deleteW, deleteReq)
// 	assert.Equal(t, http.StatusOK, deleteW.Code)
// }

