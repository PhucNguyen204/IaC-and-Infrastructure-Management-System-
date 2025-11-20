package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
)

type StackHandler struct {
	stackService services.IStackService
}

func NewStackHandler(stackService services.IStackService) *StackHandler {
	return &StackHandler{stackService: stackService}
}

func (h *StackHandler) RegisterRoutes(r *gin.RouterGroup) {
	stacks := r.Group("/stacks")
	{
		stacks.POST("", h.CreateStack)
		stacks.GET("", h.ListStacks)
		stacks.GET("/:id", h.GetStack)
		stacks.PUT("/:id", h.UpdateStack)
		stacks.DELETE("/:id", h.DeleteStack)
		stacks.POST("/:id/start", h.StartStack)
		stacks.POST("/:id/stop", h.StopStack)
		stacks.POST("/:id/restart", h.RestartStack)
		stacks.POST("/clone", h.CloneStack)
	}

	templates := r.Group("/stack-templates")
	{
		templates.POST("", h.CreateTemplate)
		templates.GET("", h.ListPublicTemplates)
		templates.GET("/:id", h.GetTemplate)
	}
}

func (h *StackHandler) CreateStack(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{
			Success: false,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	var req dto.CreateStackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	stack, err := h.stackService.CreateStack(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to create stack",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Stack created successfully",
		Data:    stack,
	})
}

func (h *StackHandler) GetStack(c *gin.Context) {
	stackID := c.Param("id")

	stack, err := h.stackService.GetStack(c.Request.Context(), stackID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Code:    "NOT_FOUND",
			Message: "Stack not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Stack retrieved successfully",
		Data:    stack,
	})
}

func (h *StackHandler) ListStacks(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{
			Success: false,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.stackService.ListStacks(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to list stacks",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Stacks retrieved successfully",
		Data:    result,
	})
}

func (h *StackHandler) UpdateStack(c *gin.Context) {
	stackID := c.Param("id")

	var req dto.UpdateStackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	stack, err := h.stackService.UpdateStack(c.Request.Context(), stackID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update stack",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Stack updated successfully",
		Data:    stack,
	})
}

func (h *StackHandler) DeleteStack(c *gin.Context) {
	stackID := c.Param("id")

	if err := h.stackService.DeleteStack(c.Request.Context(), stackID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete stack",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Stack deleted successfully",
	})
}

func (h *StackHandler) StartStack(c *gin.Context) {
	stackID := c.Param("id")

	if err := h.stackService.StartStack(c.Request.Context(), stackID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to start stack",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Stack started successfully",
	})
}

func (h *StackHandler) StopStack(c *gin.Context) {
	stackID := c.Param("id")

	if err := h.stackService.StopStack(c.Request.Context(), stackID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to stop stack",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Stack stopped successfully",
	})
}

func (h *StackHandler) RestartStack(c *gin.Context) {
	stackID := c.Param("id")

	if err := h.stackService.RestartStack(c.Request.Context(), stackID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to restart stack",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Stack restarted successfully",
	})
}

func (h *StackHandler) CloneStack(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{
			Success: false,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	var req dto.CloneStackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	stack, err := h.stackService.CloneStack(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to clone stack",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Stack cloned successfully",
		Data:    stack,
	})
}

func (h *StackHandler) CreateTemplate(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{
			Success: false,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	var req dto.CreateStackTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	template, err := h.stackService.CreateTemplate(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to create template",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Template created successfully",
		Data:    template,
	})
}

func (h *StackHandler) GetTemplate(c *gin.Context) {
	templateID := c.Param("id")

	template, err := h.stackService.GetTemplate(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Code:    "NOT_FOUND",
			Message: "Template not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Template retrieved successfully",
		Data:    template,
	})
}

func (h *StackHandler) ListPublicTemplates(c *gin.Context) {
	templates, err := h.stackService.ListPublicTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to list templates",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Templates retrieved successfully",
		Data:    templates,
	})
}
