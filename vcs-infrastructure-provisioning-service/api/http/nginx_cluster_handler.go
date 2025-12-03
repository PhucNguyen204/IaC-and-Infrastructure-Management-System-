package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
)

// NginxClusterHandler handles HTTP requests for Nginx cluster operations
type NginxClusterHandler struct {
	clusterService services.INginxClusterService
	logger         logger.ILogger
}

// NewNginxClusterHandler creates a new Nginx cluster handler
func NewNginxClusterHandler(clusterService services.INginxClusterService, logger logger.ILogger) *NginxClusterHandler {
	return &NginxClusterHandler{
		clusterService: clusterService,
		logger:         logger,
	}
}

// RegisterRoutes registers all nginx cluster routes
func (h *NginxClusterHandler) RegisterRoutes(r *gin.RouterGroup) {
	clusterGroup := r.Group("/nginx/cluster")
	{
		// Cluster lifecycle
		clusterGroup.POST("", h.CreateCluster)
		clusterGroup.GET("/:id", h.GetClusterInfo)
		clusterGroup.DELETE("/:id", h.DeleteCluster)
		clusterGroup.POST("/:id/start", h.StartCluster)
		clusterGroup.POST("/:id/stop", h.StopCluster)
		clusterGroup.POST("/:id/restart", h.RestartCluster)

		// Node operations
		clusterGroup.POST("/:id/nodes", h.AddNode)
		clusterGroup.DELETE("/:id/nodes/:nodeId", h.RemoveNode)

		// Configuration
		clusterGroup.PUT("/:id/config", h.UpdateClusterConfig)

		// Upstreams
		clusterGroup.GET("/:id/upstreams", h.ListUpstreams)
		clusterGroup.POST("/:id/upstreams", h.AddUpstream)
		clusterGroup.PUT("/:id/upstreams/:upstreamId", h.UpdateUpstream)
		clusterGroup.DELETE("/:id/upstreams/:upstreamId", h.DeleteUpstream)

		// Server blocks
		clusterGroup.GET("/:id/server-blocks", h.ListServerBlocks)
		clusterGroup.POST("/:id/server-blocks", h.AddServerBlock)
		clusterGroup.DELETE("/:id/server-blocks/:blockId", h.DeleteServerBlock)

		// Health & Monitoring
		clusterGroup.GET("/:id/health", h.GetClusterHealth)
		clusterGroup.GET("/:id/metrics", h.GetClusterMetrics)

		// Failover
		clusterGroup.POST("/:id/failover", h.TriggerFailover)
		clusterGroup.GET("/:id/failover-history", h.GetFailoverHistory)
	}
}

// CreateCluster creates a new Nginx HA cluster
// @Summary Create Nginx HA Cluster
// @Tags Nginx Cluster
// @Accept json
// @Produce json
// @Param request body dto.CreateNginxClusterRequest true "Create cluster request"
// @Success 201 {object} dto.NginxClusterInfoResponse
// @Router /api/v1/nginx/cluster [post]
func (h *NginxClusterHandler) CreateCluster(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{
			Success: false,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	var req dto.CreateNginxClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.clusterService.CreateCluster(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("failed to create nginx cluster", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to create Nginx cluster",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "CREATED",
		Message: "Nginx cluster created successfully",
		Data:    result,
	})
}

// GetClusterInfo retrieves cluster information
// @Summary Get Nginx Cluster Info
// @Tags Nginx Cluster
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.NginxClusterInfoResponse
// @Router /api/v1/nginx/cluster/{id} [get]
func (h *NginxClusterHandler) GetClusterInfo(c *gin.Context) {
	clusterID := c.Param("id")

	result, err := h.clusterService.GetClusterInfo(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to get cluster info", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Code:    "NOT_FOUND",
			Message: "Cluster not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Cluster info retrieved successfully",
		Data:    result,
	})
}

// DeleteCluster removes the cluster
// @Summary Delete Nginx Cluster
// @Tags Nginx Cluster
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id} [delete]
func (h *NginxClusterHandler) DeleteCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.clusterService.DeleteCluster(c.Request.Context(), clusterID); err != nil {
		h.logger.Error("failed to delete cluster", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete cluster",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Cluster deleted successfully",
	})
}

// StartCluster starts all nodes
// @Summary Start Nginx Cluster
// @Tags Nginx Cluster
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id}/start [post]
func (h *NginxClusterHandler) StartCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.clusterService.StartCluster(c.Request.Context(), clusterID); err != nil {
		h.logger.Error("failed to start cluster", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to start cluster",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Cluster started successfully",
	})
}

// StopCluster stops all nodes
// @Summary Stop Nginx Cluster
// @Tags Nginx Cluster
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id}/stop [post]
func (h *NginxClusterHandler) StopCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.clusterService.StopCluster(c.Request.Context(), clusterID); err != nil {
		h.logger.Error("failed to stop cluster", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to stop cluster",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Cluster stopped successfully",
	})
}

// RestartCluster restarts all nodes
// @Summary Restart Nginx Cluster
// @Tags Nginx Cluster
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id}/restart [post]
func (h *NginxClusterHandler) RestartCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.clusterService.RestartCluster(c.Request.Context(), clusterID); err != nil {
		h.logger.Error("failed to restart cluster", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to restart cluster",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Cluster restarted successfully",
	})
}

// AddNode adds a new node to the cluster
// @Summary Add Node to Cluster
// @Tags Nginx Cluster
// @Accept json
// @Produce json
// @Param id path string true "Cluster ID"
// @Param request body dto.AddNginxNodeRequest true "Add node request"
// @Success 201 {object} dto.NginxNodeInfo
// @Router /api/v1/nginx/cluster/{id}/nodes [post]
func (h *NginxClusterHandler) AddNode(c *gin.Context) {
	clusterID := c.Param("id")

	var req dto.AddNginxNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.clusterService.AddNode(c.Request.Context(), clusterID, req)
	if err != nil {
		h.logger.Error("failed to add node", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to add node",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "CREATED",
		Message: "Node added successfully",
		Data:    result,
	})
}

// RemoveNode removes a node from the cluster
// @Summary Remove Node from Cluster
// @Tags Nginx Cluster
// @Param id path string true "Cluster ID"
// @Param nodeId path string true "Node ID"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id}/nodes/{nodeId} [delete]
func (h *NginxClusterHandler) RemoveNode(c *gin.Context) {
	clusterID := c.Param("id")
	nodeID := c.Param("nodeId")

	if err := h.clusterService.RemoveNode(c.Request.Context(), clusterID, nodeID); err != nil {
		h.logger.Error("failed to remove node", zap.String("cluster_id", clusterID), zap.String("node_id", nodeID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to remove node",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Node removed successfully",
	})
}

// UpdateClusterConfig updates nginx configuration
// @Summary Update Cluster Config
// @Tags Nginx Cluster
// @Accept json
// @Param id path string true "Cluster ID"
// @Param request body dto.UpdateNginxClusterConfigRequest true "Update config request"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id}/config [put]
func (h *NginxClusterHandler) UpdateClusterConfig(c *gin.Context) {
	clusterID := c.Param("id")

	var req dto.UpdateNginxClusterConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.clusterService.UpdateClusterConfig(c.Request.Context(), clusterID, req); err != nil {
		h.logger.Error("failed to update config", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update config",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Config updated successfully",
	})
}

// ListUpstreams lists all upstreams
// @Summary List Upstreams
// @Tags Nginx Cluster
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {array} dto.UpstreamInfo
// @Router /api/v1/nginx/cluster/{id}/upstreams [get]
func (h *NginxClusterHandler) ListUpstreams(c *gin.Context) {
	clusterID := c.Param("id")

	result, err := h.clusterService.ListUpstreams(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to list upstreams", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to list upstreams",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Data:    result,
	})
}

// AddUpstream adds an upstream
// @Summary Add Upstream
// @Tags Nginx Cluster
// @Accept json
// @Param id path string true "Cluster ID"
// @Param request body dto.AddNginxUpstreamRequest true "Add upstream request"
// @Success 201 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id}/upstreams [post]
func (h *NginxClusterHandler) AddUpstream(c *gin.Context) {
	clusterID := c.Param("id")

	var req dto.AddNginxUpstreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.clusterService.AddUpstream(c.Request.Context(), clusterID, req); err != nil {
		h.logger.Error("failed to add upstream", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to add upstream",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "CREATED",
		Message: "Upstream added successfully",
	})
}

// UpdateUpstream updates an upstream
// @Summary Update Upstream
// @Tags Nginx Cluster
// @Accept json
// @Param id path string true "Cluster ID"
// @Param upstreamId path string true "Upstream ID"
// @Param request body dto.UpdateNginxUpstreamRequest true "Update upstream request"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id}/upstreams/{upstreamId} [put]
func (h *NginxClusterHandler) UpdateUpstream(c *gin.Context) {
	clusterID := c.Param("id")
	upstreamID := c.Param("upstreamId")

	var req dto.UpdateNginxUpstreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.clusterService.UpdateUpstream(c.Request.Context(), clusterID, upstreamID, req); err != nil {
		h.logger.Error("failed to update upstream", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update upstream",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Upstream updated successfully",
	})
}

// DeleteUpstream deletes an upstream
// @Summary Delete Upstream
// @Tags Nginx Cluster
// @Param id path string true "Cluster ID"
// @Param upstreamId path string true "Upstream ID"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id}/upstreams/{upstreamId} [delete]
func (h *NginxClusterHandler) DeleteUpstream(c *gin.Context) {
	clusterID := c.Param("id")
	upstreamID := c.Param("upstreamId")

	if err := h.clusterService.DeleteUpstream(c.Request.Context(), clusterID, upstreamID); err != nil {
		h.logger.Error("failed to delete upstream", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete upstream",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Upstream deleted successfully",
	})
}

// ListServerBlocks lists all server blocks
// @Summary List Server Blocks
// @Tags Nginx Cluster
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {array} dto.ServerBlockInfo
// @Router /api/v1/nginx/cluster/{id}/server-blocks [get]
func (h *NginxClusterHandler) ListServerBlocks(c *gin.Context) {
	clusterID := c.Param("id")

	result, err := h.clusterService.ListServerBlocks(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to list server blocks", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to list server blocks",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Data:    result,
	})
}

// AddServerBlock adds a server block
// @Summary Add Server Block
// @Tags Nginx Cluster
// @Accept json
// @Param id path string true "Cluster ID"
// @Param request body dto.AddNginxServerBlockRequest true "Add server block request"
// @Success 201 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id}/server-blocks [post]
func (h *NginxClusterHandler) AddServerBlock(c *gin.Context) {
	clusterID := c.Param("id")

	var req dto.AddNginxServerBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.clusterService.AddServerBlock(c.Request.Context(), clusterID, req); err != nil {
		h.logger.Error("failed to add server block", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to add server block",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "CREATED",
		Message: "Server block added successfully",
	})
}

// DeleteServerBlock deletes a server block
// @Summary Delete Server Block
// @Tags Nginx Cluster
// @Param id path string true "Cluster ID"
// @Param blockId path string true "Block ID"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/nginx/cluster/{id}/server-blocks/{blockId} [delete]
func (h *NginxClusterHandler) DeleteServerBlock(c *gin.Context) {
	clusterID := c.Param("id")
	blockID := c.Param("blockId")

	if err := h.clusterService.DeleteServerBlock(c.Request.Context(), clusterID, blockID); err != nil {
		h.logger.Error("failed to delete server block", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete server block",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Server block deleted successfully",
	})
}

// GetClusterHealth returns cluster health status
// @Summary Get Cluster Health
// @Tags Nginx Cluster
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.NginxClusterHealthResponse
// @Router /api/v1/nginx/cluster/{id}/health [get]
func (h *NginxClusterHandler) GetClusterHealth(c *gin.Context) {
	clusterID := c.Param("id")

	result, err := h.clusterService.GetClusterHealth(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to get cluster health", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get cluster health",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Data:    result,
	})
}

// GetClusterMetrics returns cluster metrics
// @Summary Get Cluster Metrics
// @Tags Nginx Cluster
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.NginxClusterMetricsResponse
// @Router /api/v1/nginx/cluster/{id}/metrics [get]
func (h *NginxClusterHandler) GetClusterMetrics(c *gin.Context) {
	clusterID := c.Param("id")

	result, err := h.clusterService.GetClusterMetrics(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to get cluster metrics", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get cluster metrics",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Data:    result,
	})
}

// TriggerFailover manually triggers failover
// @Summary Trigger Failover
// @Tags Nginx Cluster
// @Accept json
// @Produce json
// @Param id path string true "Cluster ID"
// @Param request body dto.TriggerNginxFailoverRequest true "Failover request"
// @Success 200 {object} dto.NginxFailoverResponse
// @Router /api/v1/nginx/cluster/{id}/failover [post]
func (h *NginxClusterHandler) TriggerFailover(c *gin.Context) {
	clusterID := c.Param("id")

	var req dto.TriggerNginxFailoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.clusterService.TriggerFailover(c.Request.Context(), clusterID, req)
	if err != nil {
		h.logger.Error("failed to trigger failover", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to trigger failover",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Failover triggered successfully",
		Data:    result,
	})
}

// GetFailoverHistory returns failover history
// @Summary Get Failover History
// @Tags Nginx Cluster
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} dto.NginxFailoverHistoryResponse
// @Router /api/v1/nginx/cluster/{id}/failover-history [get]
func (h *NginxClusterHandler) GetFailoverHistory(c *gin.Context) {
	clusterID := c.Param("id")

	result, err := h.clusterService.GetFailoverHistory(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to get failover history", zap.String("cluster_id", clusterID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get failover history",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Data:    result,
	})
}
