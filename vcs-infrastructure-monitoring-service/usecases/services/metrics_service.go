package services

import (
	"context"

	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/infrastructures/elasticsearch"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type IMetricsService interface {
	GetCurrentMetrics(ctx context.Context, instanceID string) (*dto.MetricsResponse, error)
	GetHistoricalMetrics(ctx context.Context, instanceID string, from, size int) ([]dto.MetricsResponse, error)
	GetLogs(ctx context.Context, instanceID string, from, size int) ([]dto.LogsResponse, error)
	AggregateMetrics(ctx context.Context, instanceID string, timeRange string) (*dto.AggregatedMetricsResponse, error)
}

type metricsService struct {
	esClient elasticsearch.IElasticsearchClient
	logger   logger.ILogger
}

func NewMetricsService(esClient elasticsearch.IElasticsearchClient, logger logger.ILogger) IMetricsService {
	return &metricsService{
		esClient: esClient,
		logger:   logger,
	}
}

func (ms *metricsService) GetCurrentMetrics(ctx context.Context, instanceID string) (*dto.MetricsResponse, error) {
	metrics, err := ms.esClient.QueryMetrics(ctx, instanceID, 0, 1)
	if err != nil {
		ms.logger.Error("failed to query metrics", zap.Error(err))
		return nil, err
	}

	if len(metrics) == 0 {
		return nil, nil
	}

	metric := metrics[0]
	return &dto.MetricsResponse{
		InstanceID:    metric.InstanceID,
		Timestamp:     metric.Timestamp.String(),
		CPUPercent:    metric.CPUPercent,
		MemoryUsed:    metric.MemoryUsed,
		MemoryLimit:   metric.MemoryLimit,
		MemoryPercent: metric.MemoryPercent,
		NetworkRx:     metric.NetworkRx,
		NetworkTx:     metric.NetworkTx,
		DiskRead:      metric.DiskRead,
		DiskWrite:     metric.DiskWrite,
		Metadata:      metric.Metadata,
	}, nil
}

func (ms *metricsService) GetHistoricalMetrics(ctx context.Context, instanceID string, from, size int) ([]dto.MetricsResponse, error) {
	metrics, err := ms.esClient.QueryMetrics(ctx, instanceID, from, size)
	if err != nil {
		ms.logger.Error("failed to query historical metrics", zap.Error(err))
		return nil, err
	}

	result := make([]dto.MetricsResponse, 0, len(metrics))
	for _, metric := range metrics {
		result = append(result, dto.MetricsResponse{
			InstanceID:    metric.InstanceID,
			Timestamp:     metric.Timestamp.String(),
			CPUPercent:    metric.CPUPercent,
			MemoryUsed:    metric.MemoryUsed,
			MemoryLimit:   metric.MemoryLimit,
			MemoryPercent: metric.MemoryPercent,
			NetworkRx:     metric.NetworkRx,
			NetworkTx:     metric.NetworkTx,
			DiskRead:      metric.DiskRead,
			DiskWrite:     metric.DiskWrite,
			Metadata:      metric.Metadata,
		})
	}

	return result, nil
}

func (ms *metricsService) GetLogs(ctx context.Context, instanceID string, from, size int) ([]dto.LogsResponse, error) {
	logs, err := ms.esClient.QueryLogs(ctx, instanceID, from, size)
	if err != nil {
		ms.logger.Error("failed to query logs", zap.Error(err))
		return nil, err
	}

	result := make([]dto.LogsResponse, 0, len(logs))
	for _, log := range logs {
		result = append(result, dto.LogsResponse{
			InstanceID: log.InstanceID,
			Timestamp:  log.Timestamp.String(),
			Message:    log.Message,
			Level:      log.Level,
			Action:     log.Action,
			Metadata:   log.Metadata,
		})
	}

	return result, nil
}

func (ms *metricsService) AggregateMetrics(ctx context.Context, instanceID string, timeRange string) (*dto.AggregatedMetricsResponse, error) {
	size := 100
	if timeRange == "1h" {
		size = 60
	} else if timeRange == "24h" {
		size = 1440
	} else if timeRange == "7d" {
		size = 10080
	}

	metrics, err := ms.esClient.QueryMetrics(ctx, instanceID, 0, size)
	if err != nil {
		ms.logger.Error("failed to query metrics for aggregation", zap.Error(err))
		return nil, err
	}

	if len(metrics) == 0 {
		return &dto.AggregatedMetricsResponse{
			InstanceID: instanceID,
			TimeRange:  timeRange,
			DataPoints: 0,
		}, nil
	}

	var cpuSum, memSum, netRxSum, netTxSum, diskReadSum, diskWriteSum float64
	var cpuMax, memMax, netRxMax, netTxMax, diskReadMax, diskWriteMax float64
	cpuMin := metrics[0].CPUPercent
	memMin := metrics[0].MemoryPercent
	netRxMin := float64(metrics[0].NetworkRx)
	netTxMin := float64(metrics[0].NetworkTx)
	diskReadMin := float64(metrics[0].DiskRead)
	diskWriteMin := float64(metrics[0].DiskWrite)

	for _, m := range metrics {
		cpuSum += m.CPUPercent
		memSum += m.MemoryPercent
		netRxSum += float64(m.NetworkRx)
		netTxSum += float64(m.NetworkTx)
		diskReadSum += float64(m.DiskRead)
		diskWriteSum += float64(m.DiskWrite)

		if m.CPUPercent > cpuMax {
			cpuMax = m.CPUPercent
		}
		if m.CPUPercent < cpuMin {
			cpuMin = m.CPUPercent
		}

		if m.MemoryPercent > memMax {
			memMax = m.MemoryPercent
		}
		if m.MemoryPercent < memMin {
			memMin = m.MemoryPercent
		}

		if float64(m.NetworkRx) > netRxMax {
			netRxMax = float64(m.NetworkRx)
		}
		if float64(m.NetworkRx) < netRxMin {
			netRxMin = float64(m.NetworkRx)
		}

		if float64(m.NetworkTx) > netTxMax {
			netTxMax = float64(m.NetworkTx)
		}
		if float64(m.NetworkTx) < netTxMin {
			netTxMin = float64(m.NetworkTx)
		}

		if float64(m.DiskRead) > diskReadMax {
			diskReadMax = float64(m.DiskRead)
		}
		if float64(m.DiskRead) < diskReadMin {
			diskReadMin = float64(m.DiskRead)
		}

		if float64(m.DiskWrite) > diskWriteMax {
			diskWriteMax = float64(m.DiskWrite)
		}
		if float64(m.DiskWrite) < diskWriteMin {
			diskWriteMin = float64(m.DiskWrite)
		}
	}

	count := float64(len(metrics))

	return &dto.AggregatedMetricsResponse{
		InstanceID: instanceID,
		TimeRange:  timeRange,
		CPUPercent: dto.AggregatedValue{
			Avg: cpuSum / count,
			Max: cpuMax,
			Min: cpuMin,
		},
		MemoryPercent: dto.AggregatedValue{
			Avg: memSum / count,
			Max: memMax,
			Min: memMin,
		},
		NetworkRx: dto.AggregatedValue{
			Avg: netRxSum / count,
			Max: netRxMax,
			Min: netRxMin,
		},
		NetworkTx: dto.AggregatedValue{
			Avg: netTxSum / count,
			Max: netTxMax,
			Min: netTxMin,
		},
		DiskRead: dto.AggregatedValue{
			Avg: diskReadSum / count,
			Max: diskReadMax,
			Min: diskReadMin,
		},
		DiskWrite: dto.AggregatedValue{
			Avg: diskWriteSum / count,
			Max: diskWriteMax,
			Min: diskWriteMin,
		},
		DataPoints: len(metrics),
	}, nil
}
