package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
)

type DockerServiceHandler struct {
	dockerService services.IDockerServiceService
}

func NewDockerServiceHandler(dockerService services.IDockerServiceService) *DockerServiceHandler {
	return &DockerServiceHandler{dockerService: dockerService}
}

func (h *DockerServiceHandler) RegisterRoutes(r *gin.RouterGroup) {
	docker := r.Group("/docker")
	{
		docker.POST("", h.CreateDockerService)
		docker.GET("/:id", h.GetDockerService)
		docker.POST("/:id/start", h.StartDockerService)
		docker.POST("/:id/stop", h.StopDockerService)
		docker.POST("/:id/restart", h.RestartDockerService)
		docker.DELETE("/:id", h.DeleteDockerService)
		docker.PUT("/:id/env", h.UpdateEnvVars)
		docker.GET("/:id/logs", h.GetServiceLogs)
	}
}

func (h *DockerServiceHandler) CreateDockerService(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{
			Success: false,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	var req dto.CreateDockerServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	service, err := h.dockerService.CreateDockerService(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to create Docker service",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Docker service created successfully",
		Data:    service,
	})
}

func (h *DockerServiceHandler) GetDockerService(c *gin.Context) {
	serviceID := c.Param("id")

	service, err := h.dockerService.GetDockerService(c.Request.Context(), serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get Docker service",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Docker service retrieved successfully",
		Data:    service,
	})
}

func (h *DockerServiceHandler) StartDockerService(c *gin.Context) {
	serviceID := c.Param("id")

	if err := h.dockerService.StartDockerService(c.Request.Context(), serviceID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to start Docker service",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Docker service started successfully",
	})
}

func (h *DockerServiceHandler) StopDockerService(c *gin.Context) {
	serviceID := c.Param("id")

	if err := h.dockerService.StopDockerService(c.Request.Context(), serviceID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to stop Docker service",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Docker service stopped successfully",
	})
}

func (h *DockerServiceHandler) RestartDockerService(c *gin.Context) {
	serviceID := c.Param("id")

	if err := h.dockerService.RestartDockerService(c.Request.Context(), serviceID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to restart Docker service",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Docker service restarted successfully",
	})
}

func (h *DockerServiceHandler) DeleteDockerService(c *gin.Context) {
	serviceID := c.Param("id")

	if err := h.dockerService.DeleteDockerService(c.Request.Context(), serviceID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete Docker service",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Docker service deleted successfully",
	})
}

func (h *DockerServiceHandler) UpdateEnvVars(c *gin.Context) {
	serviceID := c.Param("id")

	var req dto.UpdateDockerEnvRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	if err := h.dockerService.UpdateEnvVars(c.Request.Context(), serviceID, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update environment variables",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Environment variables updated successfully",
	})
}

func (h *DockerServiceHandler) GetServiceLogs(c *gin.Context) {
	serviceID := c.Param("id")
	tailStr := c.DefaultQuery("tail", "100")

	tail, err := strconv.Atoi(tailStr)
	if err != nil {
		tail = 100
	}

	logs, err := h.dockerService.GetServiceLogs(c.Request.Context(), serviceID, tail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get service logs",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Service logs retrieved successfully",
		Data: dto.DockerServiceLogsResponse{
			Logs: logs,
		},
	})
}
