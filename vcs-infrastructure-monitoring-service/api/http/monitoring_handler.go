package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/usecases/services"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type MonitoringHandler struct {
	metricsService services.IMetricsService
	redisClient    *redis.Client
}

func NewMonitoringHandler(metricsService services.IMetricsService, redisClient *redis.Client) *MonitoringHandler {
	return &MonitoringHandler{
		metricsService: metricsService,
		redisClient:    redisClient,
	}
}

func (h *MonitoringHandler) RegisterRoutes(r *gin.RouterGroup) {
	monitoring := r.Group("/monitoring")
	{
		monitoring.GET("/metrics/:instance_id", h.GetCurrentMetrics)
		monitoring.GET("/metrics/:instance_id/history", h.GetHistoricalMetrics)
		monitoring.GET("/metrics/:instance_id/aggregate", h.GetAggregatedMetrics)
		monitoring.GET("/logs/:instance_id", h.GetLogs)
		monitoring.GET("/health/:instance_id", h.GetHealthStatus)
		monitoring.GET("/infrastructure", h.ListInfrastructure)
	}
}

func (h *MonitoringHandler) GetCurrentMetrics(c *gin.Context) {
	instanceID := c.Param("instance_id")

	// Try to resolve container ID from Redis if instanceID is an infra/cluster ID
	resolvedID := h.resolveContainerID(c.Request.Context(), instanceID)

	metrics, err := h.metricsService.GetCurrentMetrics(c.Request.Context(), resolvedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get metrics",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Metrics retrieved successfully",
		Data:    metrics,
	})
}

func (h *MonitoringHandler) GetHistoricalMetrics(c *gin.Context) {
	instanceID := c.Param("instance_id")
	from, _ := strconv.Atoi(c.DefaultQuery("from", "0"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "100"))

	resolvedID := h.resolveContainerID(c.Request.Context(), instanceID)
	metrics, err := h.metricsService.GetHistoricalMetrics(c.Request.Context(), resolvedID, from, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get historical metrics",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Historical metrics retrieved successfully",
		Data:    metrics,
	})
}

func (h *MonitoringHandler) GetLogs(c *gin.Context) {
	instanceID := c.Param("instance_id")
	from, _ := strconv.Atoi(c.DefaultQuery("from", "0"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "100"))

	resolvedID := h.resolveContainerID(c.Request.Context(), instanceID)
	logs, err := h.metricsService.GetLogs(c.Request.Context(), resolvedID, from, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get logs",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Logs retrieved successfully",
		Data:    logs,
	})
}

func (h *MonitoringHandler) GetHealthStatus(c *gin.Context) {
	instanceID := c.Param("instance_id")

	status, err := h.redisClient.Get(context.Background(), "infra:status:"+instanceID).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get health status",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Health status retrieved successfully",
		Data: dto.HealthStatusResponse{
			InstanceID: instanceID,
			Status:     status,
		},
	})
}

func (h *MonitoringHandler) GetAggregatedMetrics(c *gin.Context) {
	instanceID := c.Param("instance_id")
	timeRange := c.DefaultQuery("range", "1h")

	resolvedID := h.resolveContainerID(c.Request.Context(), instanceID)
	metrics, err := h.metricsService.AggregateMetrics(c.Request.Context(), resolvedID, timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to aggregate metrics",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Aggregated metrics retrieved successfully",
		Data:    metrics,
	})
}

func (h *MonitoringHandler) ListInfrastructure(c *gin.Context) {
	keys, err := h.redisClient.Keys(c.Request.Context(), "infra:container:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to list infrastructure",
			Error:   err.Error(),
		})
		return
	}

	infrastructures := make([]dto.InfrastructureListResponse, 0)
	for _, key := range keys {
		containerID, err := h.redisClient.Get(c.Request.Context(), key).Result()
		if err != nil {
			continue
		}

		statusKey := "infra:status:" + containerID
		status, _ := h.redisClient.Get(c.Request.Context(), statusKey).Result()
		if status == "" {
			status = "unknown"
		}

		infraID := key[len("infra:container:"):]
		infrastructures = append(infrastructures, dto.InfrastructureListResponse{
			ID:     infraID,
			Name:   infraID,
			Type:   "unknown",
			Status: status,
		})
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Infrastructure list retrieved successfully",
		Data:    infrastructures,
	})
}

// resolveContainerID tries to resolve an infra/cluster ID to its actual container ID
// by looking up in Redis. If not found, returns the original ID (assuming it's already a container ID)
func (h *MonitoringHandler) resolveContainerID(ctx context.Context, instanceID string) string {
	if h.redisClient == nil {
		return instanceID
	}

	// Try to find container ID from Redis using infra:container:{instanceID} key
	containerKey := "infra:container:" + instanceID
	containerID, err := h.redisClient.Get(ctx, containerKey).Result()
	if err == nil && containerID != "" {
		return containerID
	}

	// Not found, assume instanceID is already a container ID
	return instanceID
}
