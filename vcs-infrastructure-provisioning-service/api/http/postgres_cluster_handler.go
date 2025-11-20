package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
	"go.uber.org/zap"
)

type PostgreSQLClusterHandler struct {
	clusterService services.IPostgreSQLClusterService
	logger         logger.ILogger
}

func NewPostgreSQLClusterHandler(
	clusterService services.IPostgreSQLClusterService,
	logger logger.ILogger,
) *PostgreSQLClusterHandler {
	return &PostgreSQLClusterHandler{
		clusterService: clusterService,
		logger:         logger,
	}
}

// CreateCluster creates a new PostgreSQL cluster
// @Summary Create PostgreSQL cluster
// @Tags PostgreSQL Cluster
// @Accept json
// @Produce json
// @Param request body dto.CreateClusterRequest true "Cluster configuration"
// @Success 201 {object} dto.ClusterInfoResponse
// @Router /api/v1/postgres/cluster [post]
func (h *PostgreSQLClusterHandler) CreateCluster(c *gin.Context) {
	var req dto.CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "system" // Default for testing
	}

	// Use background context with timeout for long-running cluster creation
	// This prevents HTTP request timeout from canceling replica creation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cluster, err := h.clusterService.CreateCluster(ctx, userID.(string), req)
	if err != nil {
		h.logger.Error("failed to create cluster", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cluster)
}

// GetClusterInfo retrieves cluster information
// @Summary Get cluster info
// @Tags PostgreSQL Cluster
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.ClusterInfoResponse
// @Router /api/v1/postgres/cluster/{id} [get]
func (h *PostgreSQLClusterHandler) GetClusterInfo(c *gin.Context) {
	clusterID := c.Param("id")

	cluster, err := h.clusterService.GetClusterInfo(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to get cluster info", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cluster)
}

// StartCluster starts a stopped cluster
// @Summary Start cluster
// @Tags PostgreSQL Cluster
// @Param id path string true "Cluster ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/postgres/cluster/{id}/start [post]
func (h *PostgreSQLClusterHandler) StartCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.clusterService.StartCluster(c.Request.Context(), clusterID); err != nil {
		h.logger.Error("failed to start cluster", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cluster started successfully"})
}

// StopCluster stops a running cluster
// @Summary Stop cluster
// @Tags PostgreSQL Cluster
// @Param id path string true "Cluster ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/postgres/cluster/{id}/stop [post]
func (h *PostgreSQLClusterHandler) StopCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.clusterService.StopCluster(c.Request.Context(), clusterID); err != nil {
		h.logger.Error("failed to stop cluster", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cluster stopped successfully"})
}

// RestartCluster restarts a cluster
// @Summary Restart cluster
// @Tags PostgreSQL Cluster
// @Param id path string true "Cluster ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/postgres/cluster/{id}/restart [post]
func (h *PostgreSQLClusterHandler) RestartCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.clusterService.RestartCluster(c.Request.Context(), clusterID); err != nil {
		h.logger.Error("failed to restart cluster", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cluster restarted successfully"})
}

// DeleteCluster deletes a cluster and all resources
// @Summary Delete cluster
// @Tags PostgreSQL Cluster
// @Param id path string true "Cluster ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/postgres/cluster/{id} [delete]
func (h *PostgreSQLClusterHandler) DeleteCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.clusterService.DeleteCluster(c.Request.Context(), clusterID); err != nil {
		h.logger.Error("failed to delete cluster", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cluster deleted successfully"})
}

// ScaleCluster scales cluster up or down
// @Summary Scale cluster
// @Tags PostgreSQL Cluster
// @Accept json
// @Produce json
// @Param id path string true "Cluster ID"
// @Param request body dto.ScaleClusterRequest true "Target node count"
// @Success 200 {object} map[string]string
// @Router /api/v1/postgres/cluster/{id}/scale [post]
func (h *PostgreSQLClusterHandler) ScaleCluster(c *gin.Context) {
	clusterID := c.Param("id")

	var req dto.ScaleClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.clusterService.ScaleCluster(c.Request.Context(), clusterID, req); err != nil {
		h.logger.Error("failed to scale cluster", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cluster scaled successfully"})
}

// PromoteReplica promotes a replica to primary (manual failover)
// @Summary Manual failover
// @Tags PostgreSQL Cluster
// @Accept json
// @Produce json
// @Param id path string true "Cluster ID"
// @Param request body dto.TriggerFailoverRequest true "New primary node ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/postgres/cluster/{id}/failover [post]
func (h *PostgreSQLClusterHandler) PromoteReplica(c *gin.Context) {
	clusterID := c.Param("id")

	var req dto.TriggerFailoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.clusterService.PromoteReplica(c.Request.Context(), clusterID, req.NewPrimaryNodeID); err != nil {
		h.logger.Error("failed to promote replica", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "failover completed successfully"})
}

// GetReplicationStatus shows replication status for all nodes
// @Summary Get replication status
// @Tags PostgreSQL Cluster
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.ReplicationStatusResponse
// @Router /api/v1/postgres/cluster/{id}/replication [get]
func (h *PostgreSQLClusterHandler) GetReplicationStatus(c *gin.Context) {
	clusterID := c.Param("id")

	status, err := h.clusterService.GetReplicationStatus(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to get replication status", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)

	/*
		status, err := h.clusterService.GetReplicationStatus(c.Request.Context(), clusterID)
		if err != nil {
			h.logger.Error("failed to get replication status", zap.String("cluster_id", clusterID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, status)
	*/
}

// GetClusterStats returns aggregated stats
// @Summary Get cluster statistics
// @Tags PostgreSQL Cluster
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.ClusterStatsResponse
// @Router /api/v1/postgres/cluster/{id}/stats [get]
func (h *PostgreSQLClusterHandler) GetClusterStats(c *gin.Context) {
	clusterID := c.Param("id")

	stats, err := h.clusterService.GetClusterStats(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to get cluster stats", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetClusterLogs retrieves logs from all nodes
// @Summary Get cluster logs
// @Tags PostgreSQL Cluster
// @Produce json
// @Param id path string true "Cluster ID"
// @Param tail query string false "Number of lines" default(100)
// @Success 200 {object} dto.ClusterLogsResponse
// @Router /api/v1/postgres/cluster/{id}/logs [get]
func (h *PostgreSQLClusterHandler) GetClusterLogs(c *gin.Context) {
	clusterID := c.Param("id")
	tail := c.DefaultQuery("tail", "100")

	logs, err := h.clusterService.GetClusterLogs(c.Request.Context(), clusterID, tail)
	if err != nil {
		h.logger.Error("failed to get cluster logs", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}
