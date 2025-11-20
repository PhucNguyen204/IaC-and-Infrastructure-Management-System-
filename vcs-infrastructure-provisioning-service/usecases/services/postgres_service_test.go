package services

// import (
// 	"context"
// 	"testing"

// 	"github.com/docker/docker/api/types"
// 	dockerTypes "github.com/docker/docker/api/types/container"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
// 	appDocker "github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/docker"
// 	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/kafka"
// 	appLogger "github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
// 	"go.uber.org/zap"
// )

// type MockInfrastructureRepository struct {
// 	mock.Mock
// }

// func (m *MockInfrastructureRepository) Create(infra *entities.Infrastructure) error {
// 	args := m.Called(infra)
// 	return args.Error(0)
// }

// func (m *MockInfrastructureRepository) FindByID(id string) (*entities.Infrastructure, error) {
// 	args := m.Called(id)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*entities.Infrastructure), args.Error(1)
// }

// func (m *MockInfrastructureRepository) FindByUserID(userID string) ([]*entities.Infrastructure, error) {
// 	args := m.Called(userID)
// 	return args.Get(0).([]*entities.Infrastructure), args.Error(1)
// }

// func (m *MockInfrastructureRepository) Update(infra *entities.Infrastructure) error {
// 	args := m.Called(infra)
// 	return args.Error(0)
// }

// func (m *MockInfrastructureRepository) Delete(id string) error {
// 	args := m.Called(id)
// 	return args.Error(0)
// }

// type MockPostgreSQLRepository struct {
// 	mock.Mock
// }

// func (m *MockPostgreSQLRepository) Create(instance *entities.PostgreSQLInstance) error {
// 	args := m.Called(instance)
// 	return args.Error(0)
// }

// func (m *MockPostgreSQLRepository) FindByID(id string) (*entities.PostgreSQLInstance, error) {
// 	args := m.Called(id)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*entities.PostgreSQLInstance), args.Error(1)
// }

// func (m *MockPostgreSQLRepository) FindByInfrastructureID(infraID string) (*entities.PostgreSQLInstance, error) {
// 	args := m.Called(infraID)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*entities.PostgreSQLInstance), args.Error(1)
// }

// func (m *MockPostgreSQLRepository) Update(instance *entities.PostgreSQLInstance) error {
// 	args := m.Called(instance)
// 	return args.Error(0)
// }

// func (m *MockPostgreSQLRepository) Delete(id string) error {
// 	args := m.Called(id)
// 	return args.Error(0)
// }

// type MockDockerService struct {
// 	mock.Mock
// }

// func (m *MockDockerService) CreateContainer(ctx context.Context, config appDocker.ContainerConfig) (string, error) {
// 	args := m.Called(ctx, config)
// 	return args.String(0), args.Error(1)
// }

// func (m *MockDockerService) StartContainer(ctx context.Context, containerID string) error {
// 	args := m.Called(ctx, containerID)
// 	return args.Error(0)
// }

// func (m *MockDockerService) StopContainer(ctx context.Context, containerID string) error {
// 	args := m.Called(ctx, containerID)
// 	return args.Error(0)
// }

// func (m *MockDockerService) RestartContainer(ctx context.Context, containerID string) error {
// 	args := m.Called(ctx, containerID)
// 	return args.Error(0)
// }

// func (m *MockDockerService) RemoveContainer(ctx context.Context, containerID string) error {
// 	args := m.Called(ctx, containerID)
// 	return args.Error(0)
// }

// // func (m *MockDockerService) GetContainerStats(ctx context.Context, containerID string) (dockerTypes.StatsResponseReader, error) {
// // 	args := m.Called(ctx, containerID)
// // 	if reader, ok := args.Get(0).(dockerTypes.StatsResponseReader); ok {
// // 		return reader, args.Error(1)
// // 	}
// // 	return dockerTypes.StatsResponseReader{}, args.Error(1)
// // }

// func (m *MockDockerService) GetContainerLogs(ctx context.Context, containerID string, tail string) (string, error) {
// 	args := m.Called(ctx, containerID, tail)
// 	return args.String(0), args.Error(1)
// }

// func (m *MockDockerService) ExecCommand(ctx context.Context, containerID string, cmd []string) (string, error) {
// 	args := m.Called(ctx, containerID, cmd)
// 	return args.String(0), args.Error(1)
// }

// func (m *MockDockerService) InspectContainer(ctx context.Context, containerID string) (*types.ContainerJSON, error) {
// 	args := m.Called(ctx, containerID)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*types.ContainerJSON), args.Error(1)
// }

// func (m *MockDockerService) CreateNetwork(ctx context.Context, networkName string) (string, error) {
// 	args := m.Called(ctx, networkName)
// 	return args.String(0), args.Error(1)
// }

// func (m *MockDockerService) RemoveNetwork(ctx context.Context, networkID string) error {
// 	args := m.Called(ctx, networkID)
// 	return args.Error(0)
// }

// func (m *MockDockerService) CreateVolume(ctx context.Context, volumeName string) error {
// 	args := m.Called(ctx, volumeName)
// 	return args.Error(0)
// }

// func (m *MockDockerService) RemoveVolume(ctx context.Context, volumeName string) error {
// 	args := m.Called(ctx, volumeName)
// 	return args.Error(0)
// }

// type MockKafkaProducer struct {
// 	mock.Mock
// }

// func (m *MockKafkaProducer) PublishEvent(ctx context.Context, event kafka.InfrastructureEvent) error {
// 	args := m.Called(ctx, event)
// 	return args.Error(0)
// }

// func (m *MockKafkaProducer) Close() error {
// 	if len(m.ExpectedCalls) == 0 {
// 		return nil
// 	}
// 	args := m.Called()
// 	return args.Error(0)
// }

// type MockLogger struct {
// 	mock.Mock
// }

// func (m *MockLogger) Debug(msg string, fields ...zap.Field) { m.Called(msg, fields) }
// func (m *MockLogger) Info(msg string, fields ...zap.Field)  { m.Called(msg, fields) }
// func (m *MockLogger) Warn(msg string, fields ...zap.Field)  { m.Called(msg, fields) }
// func (m *MockLogger) Error(msg string, fields ...zap.Field) { m.Called(msg, fields) }
// func (m *MockLogger) Fatal(msg string, fields ...zap.Field) { m.Called(msg, fields) }

// func (m *MockLogger) Sync() error {
// 	args := m.Called()
// 	return args.Error(0)
// }

// func (m *MockLogger) With(fields ...zap.Field) appLogger.ILogger {
// 	args := m.Called(fields)
// 	if args.Get(0) == nil {
// 		return nil
// 	}
// 	return args.Get(0).(appLogger.ILogger)
// }

// func TestGetPostgreSQLInfo_Success(t *testing.T) {
// 	mockInfraRepo := new(MockInfrastructureRepository)
// 	mockPgRepo := new(MockPostgreSQLRepository)
// 	mockLogger := new(MockLogger)

// 	service := NewPostgreSQLService(mockInfraRepo, mockPgRepo, nil, nil, mockLogger)

// 	infraID := "test-infra-id"
// 	infra := &entities.Infrastructure{
// 		ID:     infraID,
// 		Name:   "test-pg",
// 		Type:   entities.TypePostgreSQLSingle,
// 		Status: entities.StatusRunning,
// 		UserID: "user-123",
// 	}

// 	instance := &entities.PostgreSQLInstance{
// 		ID:               "test-instance-id",
// 		InfrastructureID: infraID,
// 		ContainerID:      "container-123",
// 		Version:          "15",
// 		Port:             5432,
// 		DatabaseName:     "testdb",
// 		Username:         "testuser",
// 		Infrastructure:   *infra,
// 	}

// 	mockInfraRepo.On("FindByID", infraID).Return(infra, nil)
// 	mockPgRepo.On("FindByInfrastructureID", infraID).Return(instance, nil)
// 	mockLogger.On("Error", mock.Anything, mock.Anything).Maybe()
// 	mockLogger.On("Info", mock.Anything, mock.Anything).Maybe()

// 	ctx := context.Background()
// 	resp, err := service.GetPostgreSQLInfo(ctx, infraID)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, resp)
// 	assert.Equal(t, infraID, resp.ID)
// 	assert.Equal(t, "test-pg", resp.Name)
// 	assert.Equal(t, "running", resp.Status)

// 	mockInfraRepo.AssertExpectations(t)
// 	mockPgRepo.AssertExpectations(t)
// }

// func TestStopPostgreSQL_Success(t *testing.T) {
// 	mockInfraRepo := new(MockInfrastructureRepository)
// 	mockPgRepo := new(MockPostgreSQLRepository)
// 	mockDockerSvc := new(MockDockerService)
// 	mockKafka := new(MockKafkaProducer)
// 	mockLogger := new(MockLogger)

// 	service := NewPostgreSQLService(mockInfraRepo, mockPgRepo, mockDockerSvc, mockKafka, mockLogger)

// 	infraID := "test-infra-id"
// 	containerID := "container-123"

// 	infra := &entities.Infrastructure{
// 		ID:     infraID,
// 		Name:   "test-pg",
// 		Status: entities.StatusRunning,
// 		UserID: "user-123",
// 	}

// 	instance := &entities.PostgreSQLInstance{
// 		ID:               "test-instance-id",
// 		InfrastructureID: infraID,
// 		ContainerID:      containerID,
// 		Infrastructure:   *infra,
// 	}

// 	mockInfraRepo.On("FindByID", infraID).Return(infra, nil)
// 	mockPgRepo.On("FindByInfrastructureID", infraID).Return(instance, nil)
// 	mockDockerSvc.On("StopContainer", mock.Anything, containerID).Return(nil)
// 	mockInfraRepo.On("Update", mock.Anything).Return(nil)
// 	mockKafka.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)
// 	mockLogger.On("Error", mock.Anything, mock.Anything).Maybe()
// 	mockLogger.On("Info", mock.Anything, mock.Anything).Maybe()

// 	ctx := context.Background()
// 	err := service.StopPostgreSQL(ctx, infraID)

// 	assert.NoError(t, err)

// 	mockInfraRepo.AssertExpectations(t)
// 	mockPgRepo.AssertExpectations(t)
// 	mockDockerSvc.AssertExpectations(t)
// 	mockKafka.AssertExpectations(t)
// }
