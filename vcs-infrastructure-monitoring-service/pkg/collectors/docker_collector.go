package collectors

import (
	"context"
	"encoding/json"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type IDockerCollector interface {
	CollectContainerStats(ctx context.Context, containerID string) (*ContainerStats, error)
}

type ContainerStats struct {
	CPUPercent    float64
	MemoryUsed    int64
	MemoryLimit   int64
	MemoryPercent float64
	NetworkRx     int64
	NetworkTx     int64
	DiskRead      int64
	DiskWrite     int64
}

type dockerCollector struct {
	client *client.Client
	logger logger.ILogger
}

func NewDockerCollector(dockerClient *client.Client, logger logger.ILogger) IDockerCollector {
	return &dockerCollector{
		client: dockerClient,
		logger: logger,
	}
}

func (dc *dockerCollector) CollectContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	stats, err := dc.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		dc.logger.Error("failed to get container stats", zap.String("container_id", containerID), zap.Error(err))
		return nil, err
	}
	defer stats.Body.Close()

	var v types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil && err != io.EOF {
		dc.logger.Error("failed to decode container stats", zap.Error(err))
		return nil, err
	}

	containerStats := &ContainerStats{
		MemoryUsed:  int64(v.MemoryStats.Usage),
		MemoryLimit: int64(v.MemoryStats.Limit),
	}

	if v.MemoryStats.Limit > 0 {
		containerStats.MemoryPercent = float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit) * 100.0
	}

	cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(v.CPUStats.SystemUsage - v.PreCPUStats.SystemUsage)
	if systemDelta > 0 && cpuDelta > 0 {
		containerStats.CPUPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	for _, network := range v.Networks {
		containerStats.NetworkRx += int64(network.RxBytes)
		containerStats.NetworkTx += int64(network.TxBytes)
	}

	for _, io := range v.BlkioStats.IoServiceBytesRecursive {
		if io.Op == "Read" {
			containerStats.DiskRead += int64(io.Value)
		} else if io.Op == "Write" {
			containerStats.DiskWrite += int64(io.Value)
		}
	}

	return containerStats, nil
}
