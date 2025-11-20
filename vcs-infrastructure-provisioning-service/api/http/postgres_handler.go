package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
)

type PostgreSQLHandler struct {
	pgService services.IPostgreSQLService
}

func NewPostgreSQLHandler(pgService services.IPostgreSQLService) *PostgreSQLHandler {
	return &PostgreSQLHandler{pgService: pgService}
}

func (h *PostgreSQLHandler) RegisterRoutes(r *gin.RouterGroup) {
	postgres := r.Group("/postgres")
	{
		single := postgres.Group("/single")
		{
			single.POST("", h.CreatePostgreSQL)
			single.GET("/:id", h.GetPostgreSQLInfo)
			single.GET("/:id/logs", h.GetPostgreSQLLogs)
			single.GET("/:id/stats", h.GetPostgreSQLStats)
			single.POST("/:id/start", h.StartPostgreSQL)
			single.POST("/:id/stop", h.StopPostgreSQL)
			single.POST("/:id/restart", h.RestartPostgreSQL)
			single.DELETE("/:id", h.DeletePostgreSQL)
			single.POST("/:id/backup", h.BackupPostgreSQL)
			single.POST("/:id/restore", h.RestorePostgreSQL)
		}
	}
}

func (h *PostgreSQLHandler) CreatePostgreSQL(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{
			Success: false,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	var req dto.CreatePostgreSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	resp, err := h.pgService.CreatePostgreSQL(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to create PostgreSQL instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "CREATED",
		Message: "PostgreSQL instance created successfully",
		Data:    resp,
	})
}

func (h *PostgreSQLHandler) GetPostgreSQLInfo(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.pgService.GetPostgreSQLInfo(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get PostgreSQL info",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "PostgreSQL info retrieved successfully",
		Data:    resp,
	})
}

func (h *PostgreSQLHandler) StartPostgreSQL(c *gin.Context) {
	id := c.Param("id")

	if err := h.pgService.StartPostgreSQL(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to start PostgreSQL instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "PostgreSQL instance started successfully",
	})
}

func (h *PostgreSQLHandler) StopPostgreSQL(c *gin.Context) {
	id := c.Param("id")

	if err := h.pgService.StopPostgreSQL(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to stop PostgreSQL instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "PostgreSQL instance stopped successfully",
	})
}

func (h *PostgreSQLHandler) RestartPostgreSQL(c *gin.Context) {
	id := c.Param("id")

	if err := h.pgService.RestartPostgreSQL(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to restart PostgreSQL instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "PostgreSQL instance restarted successfully",
	})
}

func (h *PostgreSQLHandler) DeletePostgreSQL(c *gin.Context) {
	id := c.Param("id")

	if err := h.pgService.DeletePostgreSQL(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete PostgreSQL instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "PostgreSQL instance deleted successfully",
	})
}

func (h *PostgreSQLHandler) BackupPostgreSQL(c *gin.Context) {
	id := c.Param("id")

	var req dto.BackupPostgreSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	resp, err := h.pgService.BackupPostgreSQL(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to backup PostgreSQL instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "PostgreSQL backup created successfully",
		Data:    resp,
	})
}

func (h *PostgreSQLHandler) GetPostgreSQLLogs(c *gin.Context) {
	id := c.Param("id")
	tail := c.DefaultQuery("tail", "100")

	logs, err := h.pgService.GetPostgreSQLLogs(c.Request.Context(), id, tail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get PostgreSQL logs",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "PostgreSQL logs retrieved successfully",
		Data: map[string]interface{}{
			"logs": logs,
		},
	})
}

func (h *PostgreSQLHandler) GetPostgreSQLStats(c *gin.Context) {
	id := c.Param("id")

	stats, err := h.pgService.GetPostgreSQLStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get PostgreSQL stats",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "PostgreSQL stats retrieved successfully",
		Data:    stats,
	})
}

func (h *PostgreSQLHandler) RestorePostgreSQL(c *gin.Context) {
	id := c.Param("id")

	var req dto.RestorePostgreSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.pgService.RestorePostgreSQL(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to restore PostgreSQL instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "PostgreSQL restored successfully",
	})
}

