package http

import (
	"net/http"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"

	"github.com/gin-gonic/gin"
)

type AutoDeployHandler struct {
	service services.IAutoDeployService
	logger  logger.ILogger
}

func NewAutoDeployHandler(service services.IAutoDeployService, logger logger.ILogger) *AutoDeployHandler {
	return &AutoDeployHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AutoDeployHandler) RegisterRoutes(r *gin.RouterGroup) {
	deploy := r.Group("/deploy")
	{
		deploy.POST("", h.Deploy)
		deploy.GET("", h.List)
		deploy.GET("/:id", h.Get)
		deploy.DELETE("/:id", h.Delete)
		deploy.GET("/:id/logs", h.GetLogs)
		deploy.GET("/:id/health", h.HealthCheck)
	}
}


func (h *AutoDeployHandler) Deploy(c *gin.Context) {
	var req dto.AutoDeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request: " + err.Error())
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	if req.Image == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Image is required",
		})
		return
	}

	// Set defaults
	if req.CPU == 0 {
		req.CPU = 1
	}
	if req.Memory == 0 {
		req.Memory = 512
	}
	if req.ExposedPort == 0 {
		req.ExposedPort = 8000
	}

	// Deploy
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default"
	}
	response, err := h.service.Deploy(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("Deploy failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "DEPLOY_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "DEPLOY_SUCCESS",
		Message: "Container deployed with auto-created infrastructure",
		Data:    response,
	})
}

// Get godoc
// @Summary Get deployment details
// @Description Get details of a specific deployment including infrastructure status
// @Tags Auto Deploy
// @Produce json
// @Param id path string true "Deployment ID"
// @Success 200 {object} dto.GetDeploymentResponse
// @Failure 404 {object} dto.APIResponse
// @Router /api/v1/deploy/{id} [get]
func (h *AutoDeployHandler) Get(c *gin.Context) {
	id := c.Param("id")

	response, err := h.service.GetDeployment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Code:    "NOT_FOUND",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Data:    response,
	})
}

// List godoc
// @Summary List all deployments
// @Description List all auto-deployed containers and their infrastructure
// @Tags Auto Deploy
// @Produce json
// @Success 200 {object} dto.ListDeploymentsResponse
// @Router /api/v1/deploy [get]
func (h *AutoDeployHandler) List(c *gin.Context) {
	// TODO: Implement list deployments
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "List feature coming soon",
	})
}

// Delete godoc
// @Summary Delete deployment and infrastructure
// @Description Delete a deployment and all its auto-created infrastructure
// @Tags Auto Deploy
// @Param id path string true "Deployment ID"
// @Success 200 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /api/v1/deploy/{id} [delete]
func (h *AutoDeployHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteDeployment(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Code:    "NOT_FOUND",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "DELETED",
		Message: "Deployment and infrastructure deleted successfully",
	})
}

// GetLogs godoc
// @Summary Get deployment logs
// @Description Get logs from the deployed container
// @Tags Auto Deploy
// @Param id path string true "Deployment ID"
// @Success 200 {object} dto.DeploymentLogsResponse
// @Router /api/v1/deploy/{id}/logs [get]
func (h *AutoDeployHandler) GetLogs(c *gin.Context) {
	// TODO: Implement logs retrieval
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Logs feature coming soon",
	})
}

// HealthCheck godoc
// @Summary Check deployment health
// @Description Check health status of deployment and its infrastructure
// @Tags Auto Deploy
// @Param id path string true "Deployment ID"
// @Success 200 {object} dto.DeploymentHealthResponse
// @Router /api/v1/deploy/{id}/health [get]
func (h *AutoDeployHandler) HealthCheck(c *gin.Context) {
	// TODO: Implement health check
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Health check feature coming soon",
	})
}
