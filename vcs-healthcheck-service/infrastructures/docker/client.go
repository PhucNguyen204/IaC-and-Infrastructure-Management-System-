package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/PhucNguyen204/vcs-healthcheck-service/dto"
	"github.com/PhucNguyen204/vcs-healthcheck-service/pkg/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type IDockerClient interface {
	// Container operations
	ListContainers(ctx context.Context, prefix string) ([]types.Container, error)
	GetContainerStats(ctx context.Context, containerID string) (*dto.ContainerMetrics, error)
	InspectContainer(ctx context.Context, containerID string) (*types.ContainerJSON, error)

	// Health check
	CheckContainerHealth(ctx context.Context, containerID string) (*dto.HealthCheckResult, error)

	// Batch operations
	CollectAllMetrics(ctx context.Context, prefix string) ([]dto.ContainerMetrics, error)
}

type dockerClient struct {
	client *client.Client
	logger logger.ILogger
}

func NewDockerClient(logger logger.ILogger) (IDockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &dockerClient{
		client: cli,
		logger: logger,
	}, nil
}

func (d *dockerClient) ListContainers(ctx context.Context, prefix string) ([]types.Container, error) {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		return nil, err
	}

	if prefix == "" {
		return containers, nil
	}

	// Filter by prefix
	var filtered []types.Container
	for _, c := range containers {
		for _, name := range c.Names {
			if strings.HasPrefix(strings.TrimPrefix(name, "/"), prefix) {
				filtered = append(filtered, c)
				break
			}
		}
	}
	return filtered, nil
}

func (d *dockerClient) GetContainerStats(ctx context.Context, containerID string) (*dto.ContainerMetrics, error) {
	stats, err := d.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	var containerStats types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&containerStats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	// Get container info for name and labels
	inspect, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Calculate CPU percentage
	cpuPercent := calculateCPUPercent(&containerStats)

	// Calculate memory percentage
	memoryPercent := 0.0
	if containerStats.MemoryStats.Limit > 0 {
		memoryPercent = float64(containerStats.MemoryStats.Usage) / float64(containerStats.MemoryStats.Limit) * 100
	}

	// Calculate network stats
	var rxBytes, txBytes, rxPackets, txPackets int64
	for _, netStats := range containerStats.Networks {
		rxBytes += int64(netStats.RxBytes)
		txBytes += int64(netStats.TxBytes)
		rxPackets += int64(netStats.RxPackets)
		txPackets += int64(netStats.TxPackets)
	}

	// Get infrastructure ID and type from labels
	infraID := inspect.Config.Labels["infrastructure_id"]
	if infraID == "" {
		infraID = inspect.Config.Labels["cluster_id"]
	}
	infraType := inspect.Config.Labels["cluster_type"]

	// Calculate disk I/O safely (array might be empty)
	var diskReadBytes, diskWriteBytes int64
	if len(containerStats.BlkioStats.IoServiceBytesRecursive) > 0 {
		diskReadBytes = int64(containerStats.BlkioStats.IoServiceBytesRecursive[0].Value)
	}
	if len(containerStats.BlkioStats.IoServiceBytesRecursive) > 1 {
		diskWriteBytes = int64(containerStats.BlkioStats.IoServiceBytesRecursive[1].Value)
	}

	metrics := &dto.ContainerMetrics{
		MetricID:         uuid.New().String(),
		InfrastructureID: infraID,
		ContainerID:      containerID,
		ContainerName:    strings.TrimPrefix(inspect.Name, "/"),
		Type:             infraType,
		Timestamp:        time.Now(),
		CPU: dto.CPUMetrics{
			UsagePercent: cpuPercent,
			Cores:        int(containerStats.CPUStats.OnlineCPUs),
		},
		Memory: dto.MemoryMetrics{
			UsedBytes:    int64(containerStats.MemoryStats.Usage),
			LimitBytes:   int64(containerStats.MemoryStats.Limit),
			UsagePercent: memoryPercent,
		},
		Network: dto.NetworkMetrics{
			RxBytes:   rxBytes,
			TxBytes:   txBytes,
			RxPackets: rxPackets,
			TxPackets: txPackets,
		},
		Disk: dto.DiskMetrics{
			ReadBytes:  diskReadBytes,
			WriteBytes: diskWriteBytes,
		},
		Health: dto.HealthStatus{
			Status:    getHealthStatus(&inspect),
			LastCheck: time.Now(),
		},
	}

	return metrics, nil
}

func (d *dockerClient) InspectContainer(ctx context.Context, containerID string) (*types.ContainerJSON, error) {
	inspect, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}
	return &inspect, nil
}

func (d *dockerClient) CheckContainerHealth(ctx context.Context, containerID string) (*dto.HealthCheckResult, error) {
	start := time.Now()

	inspect, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return &dto.HealthCheckResult{
			ContainerID:   containerID,
			Status:        dto.StatusUnknown,
			Timestamp:     time.Now(),
			CheckDuration: time.Since(start).Milliseconds(),
			Message:       fmt.Sprintf("failed to inspect: %v", err),
		}, nil
	}

	status := dto.StatusHealthy
	message := "Container is running"

	if !inspect.State.Running {
		status = dto.StatusUnhealthy
		message = fmt.Sprintf("Container not running, state: %s", inspect.State.Status)
	} else if inspect.State.Health != nil {
		switch inspect.State.Health.Status {
		case "healthy":
			status = dto.StatusHealthy
			message = "Health check passed"
		case "unhealthy":
			status = dto.StatusUnhealthy
			if len(inspect.State.Health.Log) > 0 {
				lastLog := inspect.State.Health.Log[len(inspect.State.Health.Log)-1]
				message = lastLog.Output
			}
		case "starting":
			status = dto.StatusUnknown
			message = "Health check starting"
		}
	}

	infraID := inspect.Config.Labels["infrastructure_id"]
	if infraID == "" {
		infraID = inspect.Config.Labels["cluster_id"]
	}

	return &dto.HealthCheckResult{
		InfrastructureID: infraID,
		ContainerID:      containerID,
		ContainerName:    strings.TrimPrefix(inspect.Name, "/"),
		Type:             inspect.Config.Labels["cluster_type"],
		Status:           status,
		Timestamp:        time.Now(),
		CheckDuration:    time.Since(start).Milliseconds(),
		Message:          message,
	}, nil
}

func (d *dockerClient) CollectAllMetrics(ctx context.Context, prefix string) ([]dto.ContainerMetrics, error) {
	containers, err := d.ListContainers(ctx, prefix)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	metricsChan := make(chan dto.ContainerMetrics, len(containers))
	errorsChan := make(chan error, len(containers))

	for _, c := range containers {
		wg.Add(1)
		go func(containerID string) {
			defer wg.Done()
			metrics, err := d.GetContainerStats(ctx, containerID)
			if err != nil {
				errorsChan <- err
				return
			}
			metricsChan <- *metrics
		}(c.ID)
	}

	wg.Wait()
	close(metricsChan)
	close(errorsChan)

	var allMetrics []dto.ContainerMetrics
	for metrics := range metricsChan {
		allMetrics = append(allMetrics, metrics)
	}

	// Log errors but don't fail
	for err := range errorsChan {
		d.logger.Warn("failed to collect metrics for container", zap.Error(err))
	}

	return allMetrics, nil
}

// Helper functions

func calculateCPUPercent(stats *types.StatsJSON) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta > 0 && cpuDelta > 0 {
		cpuPercent := (cpuDelta / systemDelta) * float64(stats.CPUStats.OnlineCPUs) * 100.0
		return cpuPercent
	}
	return 0.0
}

func getHealthStatus(inspect *types.ContainerJSON) string {
	if !inspect.State.Running {
		return dto.StatusUnhealthy
	}
	if inspect.State.Health == nil {
		return dto.StatusUnknown
	}
	switch inspect.State.Health.Status {
	case "healthy":
		return dto.StatusHealthy
	case "unhealthy":
		return dto.StatusUnhealthy
	default:
		return dto.StatusUnknown
	}
}
