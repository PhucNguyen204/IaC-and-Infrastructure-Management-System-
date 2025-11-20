package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/infrastructures/elasticsearch"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/collectors"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type IHealthCheckService interface {
	Start(ctx context.Context) error
	CheckContainer(ctx context.Context, containerID string) (string, error)
}

type healthCheckService struct {
	dockerCollector collectors.IDockerCollector
	esClient        elasticsearch.IElasticsearchClient
	redisClient     *redis.Client
	logger          logger.ILogger
}

func NewHealthCheckService(
	dockerCollector collectors.IDockerCollector,
	esClient elasticsearch.IElasticsearchClient,
	redisClient *redis.Client,
	logger logger.ILogger,
) IHealthCheckService {
	return &healthCheckService{
		dockerCollector: dockerCollector,
		esClient:        esClient,
		redisClient:     redisClient,
		logger:          logger,
	}
}

func (hcs *healthCheckService) Start(ctx context.Context) error {
	hcs.logger.Info("health check service started")

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				hcs.logger.Info("health check service stopped")
				return
			case <-ticker.C:
				hcs.performHealthChecks(ctx)
			}
		}
	}()

	return nil
}

func (hcs *healthCheckService) performHealthChecks(ctx context.Context) {
	keys, err := hcs.redisClient.Keys(ctx, "infra:container:*").Result()
	if err != nil {
		hcs.logger.Error("failed to get container keys from redis", zap.Error(err))
		return
	}

	for _, key := range keys {
		containerID, err := hcs.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		status, err := hcs.CheckContainer(ctx, containerID)
		if err != nil {
			hcs.logger.Error("health check failed",
				zap.String("container_id", containerID),
				zap.Error(err))
			status = "unhealthy"
		}

		statusKey := "infra:status:" + containerID
		hcs.redisClient.Set(ctx, statusKey, status, time.Hour)

		stats, err := hcs.dockerCollector.CollectContainerStats(ctx, containerID)
		if err != nil {
			hcs.logger.Error("failed to collect stats",
				zap.String("container_id", containerID),
				zap.Error(err))
			continue
		}

		metricEntry := elasticsearch.MetricEntry{
			InstanceID:    containerID,
			CPUPercent:    stats.CPUPercent,
			MemoryUsed:    stats.MemoryUsed,
			MemoryLimit:   stats.MemoryLimit,
			MemoryPercent: stats.MemoryPercent,
			NetworkRx:     stats.NetworkRx,
			NetworkTx:     stats.NetworkTx,
			DiskRead:      stats.DiskRead,
			DiskWrite:     stats.DiskWrite,
		}

		if err := hcs.esClient.IndexMetric(ctx, metricEntry); err != nil {
			hcs.logger.Error("failed to index metric", zap.Error(err))
		}
	}
}

func (hcs *healthCheckService) CheckContainer(ctx context.Context, containerID string) (string, error) {
	stats, err := hcs.dockerCollector.CollectContainerStats(ctx, containerID)
	if err != nil {
		return "unhealthy", err
	}

	if stats.CPUPercent > 90 || stats.MemoryPercent > 90 {
		return "warning", nil
	}

	return "healthy", nil
}

