package http

import (
	"context"
	"net/http"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// K8sClusterHandler handles HTTP requests for K8s clusters
type K8sClusterHandler struct {
	k8sService services.IK8sClusterService
	logger     logger.ILogger
}

// NewK8sClusterHandler creates a new K8s cluster handler
func NewK8sClusterHandler(k8sService services.IK8sClusterService, logger logger.ILogger) *K8sClusterHandler {
	return &K8sClusterHandler{
		k8sService: k8sService,
		logger:     logger,
	}
}

// RegisterRoutes registers K8s cluster routes
func (h *K8sClusterHandler) RegisterRoutes(router *gin.RouterGroup) {
	k8s := router.Group("/k8s")
	{
		// Cluster operations
		k8s.POST("/cluster", h.CreateCluster)
		k8s.GET("/cluster/:id", h.GetClusterInfo)
		k8s.DELETE("/cluster/:id", h.DeleteCluster)
		k8s.POST("/cluster/:id/start", h.StartCluster)
		k8s.POST("/cluster/:id/stop", h.StopCluster)
		k8s.POST("/cluster/:id/restart", h.RestartCluster)
		k8s.POST("/cluster/:id/scale", h.ScaleCluster)
		
		// Cluster info
		k8s.GET("/cluster/:id/health", h.GetClusterHealth)
		k8s.GET("/cluster/:id/metrics", h.GetClusterMetrics)
		k8s.GET("/cluster/:id/connection-info", h.GetConnectionInfo)
		k8s.GET("/cluster/:id/kubeconfig", h.GetKubeconfig)
	}
}

// CreateCluster creates a new K8s cluster
func (h *K8sClusterHandler) CreateCluster(c *gin.Context) {
	var req dto.CreateK8sClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	h.logger.Info("creating k8s cluster", zap.String("cluster_name", req.ClusterName))

	cluster, err := h.k8sService.CreateCluster(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to create k8s cluster", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    "CLUSTER_CREATE_FAILED",
			"message": "Failed to create cluster",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"code":    "CLUSTER_CREATED",
		"message": "Cluster creation initiated",
		"data":    cluster,
	})
}

// GetClusterInfo retrieves cluster information
func (h *K8sClusterHandler) GetClusterInfo(c *gin.Context) {
	clusterID := c.Param("id")

	cluster, err := h.k8sService.GetClusterInfo(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    "CLUSTER_NOT_FOUND",
			"message": "Cluster not found",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    "SUCCESS",
		"message": "Cluster info retrieved successfully",
		"data":    cluster,
	})
}

// DeleteCluster deletes a cluster
func (h *K8sClusterHandler) DeleteCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.k8sService.DeleteCluster(c.Request.Context(), clusterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    "CLUSTER_DELETE_FAILED",
			"message": "Failed to delete cluster",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    "CLUSTER_DELETED",
		"message": "Cluster deleted successfully",
	})
}

// StartCluster starts a stopped cluster
func (h *K8sClusterHandler) StartCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.k8sService.StartCluster(c.Request.Context(), clusterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    "CLUSTER_START_FAILED",
			"message": "Failed to start cluster",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    "CLUSTER_STARTED",
		"message": "Cluster started successfully",
	})
}

// StopCluster stops a running cluster
func (h *K8sClusterHandler) StopCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.k8sService.StopCluster(c.Request.Context(), clusterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    "CLUSTER_STOP_FAILED",
			"message": "Failed to stop cluster",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    "CLUSTER_STOPPED",
		"message": "Cluster stopped successfully",
	})
}

// RestartCluster restarts a cluster
func (h *K8sClusterHandler) RestartCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := h.k8sService.RestartCluster(c.Request.Context(), clusterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    "CLUSTER_RESTART_FAILED",
			"message": "Failed to restart cluster",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    "CLUSTER_RESTARTED",
		"message": "Cluster restarted successfully",
	})
}

// ScaleCluster scales the cluster
func (h *K8sClusterHandler) ScaleCluster(c *gin.Context) {
	clusterID := c.Param("id")
	
	var req dto.ScaleK8sClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if err := h.k8sService.ScaleCluster(c.Request.Context(), clusterID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    "CLUSTER_SCALE_FAILED",
			"message": "Failed to scale cluster",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    "CLUSTER_SCALED",
		"message": "Cluster scaled successfully",
	})
}

// GetClusterHealth retrieves cluster health
func (h *K8sClusterHandler) GetClusterHealth(c *gin.Context) {
	clusterID := c.Param("id")

	health, err := h.k8sService.GetClusterHealth(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    "HEALTH_CHECK_FAILED",
			"message": "Failed to get cluster health",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    "SUCCESS",
		"message": "",
		"data":    health,
	})
}

// GetClusterMetrics retrieves cluster metrics
func (h *K8sClusterHandler) GetClusterMetrics(c *gin.Context) {
	clusterID := c.Param("id")

	metrics, err := h.k8sService.GetClusterMetrics(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    "METRICS_FAILED",
			"message": "Failed to get cluster metrics",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    "SUCCESS",
		"message": "",
		"data":    metrics,
	})
}

// GetConnectionInfo retrieves connection information
func (h *K8sClusterHandler) GetConnectionInfo(c *gin.Context) {
	clusterID := c.Param("id")

	info, err := h.k8sService.GetConnectionInfo(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    "CONNECTION_INFO_FAILED",
			"message": "Failed to get connection info",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    "SUCCESS",
		"message": "Connection info retrieved successfully",
		"data":    info,
	})
}

// GetKubeconfig retrieves the kubeconfig file
func (h *K8sClusterHandler) GetKubeconfig(c *gin.Context) {
	clusterID := c.Param("id")

	ctx := context.WithValue(c.Request.Context(), "include_kubeconfig", true)
	cluster, err := h.k8sService.GetClusterInfo(ctx, clusterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    "CLUSTER_NOT_FOUND",
			"message": "Cluster not found",
			"error":   err.Error(),
		})
		return
	}

	c.Header("Content-Type", "application/x-yaml")
	c.Header("Content-Disposition", "attachment; filename=kubeconfig-"+cluster.ClusterName+".yaml")
	c.String(http.StatusOK, cluster.Kubeconfig)
}

