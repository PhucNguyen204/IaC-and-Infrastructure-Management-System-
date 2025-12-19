package services

import (
	"context"
	"sync"
	"time"

	"github.com/PhucNguyen204/vcs-healthcheck-service/dto"
	"github.com/PhucNguyen204/vcs-healthcheck-service/infrastructures/docker"
	"github.com/PhucNguyen204/vcs-healthcheck-service/infrastructures/elasticsearch"
	"github.com/PhucNguyen204/vcs-healthcheck-service/infrastructures/kafka"
	"github.com/PhucNguyen204/vcs-healthcheck-service/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type IHealthCheckService interface {
	Start(ctx context.Context) error
	Stop() error
}

type healthCheckService struct {
	kafkaConsumer kafka.IKafkaConsumer
	esClient      elasticsearch.IElasticsearchClient
	dockerClient  docker.IDockerClient
	logger        logger.ILogger

	// Tracked infrastructures
	tracked     map[string]*trackedInfra
	trackedLock sync.RWMutex

	// Configuration
	metricsInterval     time.Duration
	healthCheckInterval time.Duration

	// Control
	stopChan chan struct{}
	wg       sync.WaitGroup
}

type trackedInfra struct {
	InfrastructureID string
	Type             string
	Status           string
	CreatedAt        time.Time
	LastCheck        time.Time
	ContainerPrefix  string
	Metadata         map[string]interface{}
}

func NewHealthCheckService(
	kafkaConsumer kafka.IKafkaConsumer,
	esClient elasticsearch.IElasticsearchClient,
	dockerClient docker.IDockerClient,
	logger logger.ILogger,
) IHealthCheckService {
	return &healthCheckService{
		kafkaConsumer:       kafkaConsumer,
		esClient:            esClient,
		dockerClient:        dockerClient,
		logger:              logger,
		tracked:             make(map[string]*trackedInfra),
		metricsInterval:     30 * time.Second,
		healthCheckInterval: 60 * time.Second,
		stopChan:            make(chan struct{}),
	}
}

func (s *healthCheckService) Start(ctx context.Context) error {
	s.logger.Info("starting health check service")

	// Start Kafka consumer
	if err := s.kafkaConsumer.Start(ctx, s.handleEvent); err != nil {
		return err
	}

	// Start metrics collector
	s.wg.Add(1)
	go s.metricsCollectorLoop(ctx)

	// Start health checker
	s.wg.Add(1)
	go s.healthCheckerLoop(ctx)

	s.logger.Info("health check service started successfully")
	return nil
}

func (s *healthCheckService) Stop() error {
	s.logger.Info("stopping health check service")
	close(s.stopChan)
	s.wg.Wait()
	s.kafkaConsumer.Close()
	s.logger.Info("health check service stopped")
	return nil
}

// handleEvent processes lifecycle events from Kafka
func (s *healthCheckService) handleEvent(ctx context.Context, event dto.LifecycleEvent) error {
	// Extract infrastructure_id from metadata if not present in event
	// Provisioning service sends infra_id in metadata
	if event.InfrastructureID == "" && event.Metadata != nil {
		if infraID, ok := event.Metadata["infra_id"].(string); ok {
			event.InfrastructureID = infraID
		}
	}

	// Extract status from metadata if not present
	if event.Status == "" && event.Metadata != nil {
		if status, ok := event.Metadata["status"].(string); ok {
			event.Status = status
		}
	}

	s.logger.Info("handling lifecycle event",
		zap.String("infrastructure_id", event.InfrastructureID),
		zap.String("action", event.Action),
		zap.String("status", event.Status))

	// Assign event ID if not present
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Index event to Elasticsearch (for logging/audit)
	if err := s.esClient.IndexEvent(ctx, event); err != nil {
		s.logger.Error("failed to index event to elasticsearch", zap.Error(err))
		// Don't return error - continue processing
	}

	// Index uptime event for monitoring service uptime calculations
	// This goes to infrastructure-uptime-* index
	if err := s.esClient.IndexUptimeEvent(ctx, event); err != nil {
		s.logger.Error("failed to index uptime event to elasticsearch", zap.Error(err))
	}

	// Update tracking based on action
	switch event.Action {
	case dto.ActionCreated, dto.ActionStarted,
		dto.ActionClusterCreated, dto.ActionClusterStarted,
		dto.ActionNginxCreated, dto.ActionNginxStarted,
		dto.ActionDinDCreated, dto.ActionDinDStarted,
		dto.ActionClickHouseCreated, dto.ActionClickHouseStarted:
		s.trackInfrastructure(event)
	case dto.ActionStopped,
		dto.ActionClusterStopped,
		dto.ActionNginxStopped,
		dto.ActionDinDStopped,
		dto.ActionClickHouseStopped:
		s.updateInfraStatus(event.InfrastructureID, dto.StatusStopped)
	case dto.ActionDeleted,
		dto.ActionClusterDeleted,
		dto.ActionNginxDeleted,
		dto.ActionDinDDeleted,
		dto.ActionClickHouseDeleted:
		s.untrackInfrastructure(event.InfrastructureID)
	case dto.ActionNodeAdded, dto.ActionNodeRemoved:
		// Just index the event, no tracking change needed
	case dto.ActionFailover:
		// Index failover event for uptime calculation
		s.updateInfraStatus(event.InfrastructureID, dto.StatusRunning)
	}

	return nil
}

func (s *healthCheckService) trackInfrastructure(event dto.LifecycleEvent) {
	s.trackedLock.Lock()
	defer s.trackedLock.Unlock()

	// Determine container prefix based on infrastructure type and metadata
	prefix := s.determineContainerPrefix(event.Type, event.Metadata)

	s.tracked[event.InfrastructureID] = &trackedInfra{
		InfrastructureID: event.InfrastructureID,
		Type:             event.Type,
		Status:           event.Status,
		CreatedAt:        event.Timestamp,
		LastCheck:        time.Now(),
		ContainerPrefix:  prefix,
		Metadata:         event.Metadata,
	}

	s.logger.Info("tracking new infrastructure",
		zap.String("infrastructure_id", event.InfrastructureID),
		zap.String("type", event.Type),
		zap.String("prefix", prefix))
}

// determineContainerPrefix returns the appropriate container name prefix based on infrastructure type
func (s *healthCheckService) determineContainerPrefix(infraType string, metadata map[string]interface{}) string {
	clusterName := ""
	if name, ok := metadata["cluster_name"].(string); ok {
		clusterName = name
	} else if name, ok := metadata["name"].(string); ok {
		clusterName = name
	}

	switch infraType {
	case dto.TypePostgresCluster:
		if clusterName != "" {
			return "pg-cluster-" + clusterName
		}
		return "pg-cluster-"

	case dto.TypeNginxCluster:
		if clusterName != "" {
			return clusterName + "-nginx"
		}
		return "nginx-"

	case dto.TypeDinD:
		if clusterName != "" {
			return "dind-" + clusterName
		}
		return "dind-"

	case dto.TypeClickHouse:
		if clusterName != "" {
			return "clickhouse-" + clusterName
		}
		return "clickhouse-"

	default:
		// Fallback: use cluster name or generic prefix
		if clusterName != "" {
			return clusterName
		}
		return "infra-"
	}
}

func (s *healthCheckService) updateInfraStatus(infraID, status string) {
	s.trackedLock.Lock()
	defer s.trackedLock.Unlock()

	if infra, ok := s.tracked[infraID]; ok {
		infra.Status = status
		infra.LastCheck = time.Now()
	}
}

func (s *healthCheckService) untrackInfrastructure(infraID string) {
	s.trackedLock.Lock()
	defer s.trackedLock.Unlock()

	delete(s.tracked, infraID)
	s.logger.Info("untracked infrastructure", zap.String("infrastructure_id", infraID))
}

// metricsCollectorLoop collects container metrics periodically
func (s *healthCheckService) metricsCollectorLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.metricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.collectMetrics(ctx)
		}
	}
}

func (s *healthCheckService) collectMetrics(ctx context.Context) {
	s.trackedLock.RLock()
	infraList := make([]*trackedInfra, 0, len(s.tracked))
	for _, infra := range s.tracked {
		if infra.Status == dto.StatusRunning {
			infraList = append(infraList, infra)
		}
	}
	s.trackedLock.RUnlock()

	for _, infra := range infraList {
		metrics, err := s.dockerClient.CollectAllMetrics(ctx, infra.ContainerPrefix)
		if err != nil {
			s.logger.Warn("failed to collect metrics",
				zap.String("infrastructure_id", infra.InfrastructureID),
				zap.Error(err))
			continue
		}

		for _, m := range metrics {
			m.InfrastructureID = infra.InfrastructureID
			m.Type = infra.Type
			if err := s.esClient.IndexMetrics(ctx, m); err != nil {
				s.logger.Warn("failed to index metrics", zap.Error(err))
			}
		}

		s.logger.Debug("collected metrics",
			zap.String("infrastructure_id", infra.InfrastructureID),
			zap.Int("container_count", len(metrics)))
	}
}

// healthCheckerLoop performs health checks periodically
func (s *healthCheckService) healthCheckerLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.performHealthChecks(ctx)
		}
	}
}

func (s *healthCheckService) performHealthChecks(ctx context.Context) {
	s.trackedLock.RLock()
	infraList := make([]*trackedInfra, 0, len(s.tracked))
	for _, infra := range s.tracked {
		infraList = append(infraList, infra)
	}
	s.trackedLock.RUnlock()

	for _, infra := range infraList {
		containers, err := s.dockerClient.ListContainers(ctx, infra.ContainerPrefix)
		if err != nil {
			s.logger.Warn("failed to list containers for health check",
				zap.String("infrastructure_id", infra.InfrastructureID),
				zap.Error(err))
			continue
		}

		for _, container := range containers {
			result, err := s.dockerClient.CheckContainerHealth(ctx, container.ID)
			if err != nil {
				s.logger.Warn("health check failed",
					zap.String("container_id", container.ID),
					zap.Error(err))
				continue
			}

			result.InfrastructureID = infra.InfrastructureID
			result.Type = infra.Type

			if err := s.esClient.IndexHealthCheck(ctx, *result); err != nil {
				s.logger.Warn("failed to index health check", zap.Error(err))
			}
		}

		// Update last check time
		s.trackedLock.Lock()
		if tracked, ok := s.tracked[infra.InfrastructureID]; ok {
			tracked.LastCheck = time.Now()
		}
		s.trackedLock.Unlock()
	}
}
