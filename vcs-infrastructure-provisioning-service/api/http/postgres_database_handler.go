package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
)

type PostgresDatabaseHandler struct {
	dbService services.IPostgresDatabaseService
}

func NewPostgresDatabaseHandler(dbService services.IPostgresDatabaseService) *PostgresDatabaseHandler {
	return &PostgresDatabaseHandler{dbService: dbService}
}

func (h *PostgresDatabaseHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/postgres/:id/databases", h.CreateDatabase)
	r.GET("/postgres/:id/databases", h.ListDatabases)
	r.GET("/postgres/:id/overview", h.GetInstanceOverview)
	r.GET("/databases/:id", h.GetDatabase)
	r.PUT("/databases/:id/quota", h.UpdateQuota)
	r.GET("/databases/:id/metrics", h.GetMetrics)
	r.POST("/databases/:id/backup", h.BackupDatabase)
	r.POST("/databases/:id/restore", h.RestoreDatabase)
	r.POST("/databases/:id/lifecycle", h.ManageLifecycle)
}

func (h *PostgresDatabaseHandler) CreateDatabase(c *gin.Context) {
	instanceID := c.Param("id")
	var req dto.CreateDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}
	result, err := h.dbService.CreateDatabase(c.Request.Context(), instanceID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to create database",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "CREATED",
		Message: "Database created successfully",
		Data:    result,
	})
}

func (h *PostgresDatabaseHandler) GetDatabase(c *gin.Context) {
	databaseID := c.Param("id")
	result, err := h.dbService.GetDatabase(c.Request.Context(), databaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get database",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Database info retrieved successfully",
		Data:    result,
	})
}

func (h *PostgresDatabaseHandler) ListDatabases(c *gin.Context) {
	instanceID := c.Param("id")
	result, err := h.dbService.ListDatabases(c.Request.Context(), instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to list databases",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Databases retrieved successfully",
		Data:    result,
	})
}

func (h *PostgresDatabaseHandler) UpdateQuota(c *gin.Context) {
	databaseID := c.Param("id")
	var req dto.UpdateQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}
	if err := h.dbService.UpdateQuota(c.Request.Context(), databaseID, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update quota",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Quota updated successfully",
	})
}

func (h *PostgresDatabaseHandler) GetMetrics(c *gin.Context) {
	databaseID := c.Param("id")
	result, err := h.dbService.GetMetrics(c.Request.Context(), databaseID)
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
		Data:    result,
	})
}

func (h *PostgresDatabaseHandler) BackupDatabase(c *gin.Context) {
	databaseID := c.Param("id")
	var req dto.BackupDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}
	result, err := h.dbService.BackupDatabase(c.Request.Context(), databaseID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to backup database",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Backup started successfully",
		Data:    result,
	})
}

func (h *PostgresDatabaseHandler) RestoreDatabase(c *gin.Context) {
	databaseID := c.Param("id")
	var req dto.RestoreDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}
	if err := h.dbService.RestoreDatabase(c.Request.Context(), databaseID, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to restore database",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Database restored successfully",
	})
}

func (h *PostgresDatabaseHandler) ManageLifecycle(c *gin.Context) {
	databaseID := c.Param("id")
	var req dto.ManageLifecycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}
	if err := h.dbService.ManageLifecycle(c.Request.Context(), databaseID, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to manage lifecycle",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Lifecycle action completed successfully",
	})
}

func (h *PostgresDatabaseHandler) GetInstanceOverview(c *gin.Context) {
	instanceID := c.Param("id")
	result, err := h.dbService.GetInstanceOverview(c.Request.Context(), instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get instance overview",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Instance overview retrieved successfully",
		Data:    result,
	})
}
