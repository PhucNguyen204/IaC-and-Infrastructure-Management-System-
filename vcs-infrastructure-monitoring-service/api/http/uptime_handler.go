package http

import (
	"net/http"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/infrastructures/elasticsearch"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/usecases/services"
	"github.com/gin-gonic/gin"
)

type UptimeHandler struct {
	uptimeService services.IUptimeService
}

func NewUptimeHandler(uptimeService services.IUptimeService) *UptimeHandler {
	return &UptimeHandler{
		uptimeService: uptimeService,
	}
}

func (h *UptimeHandler) RegisterRoutes(r *gin.RouterGroup) {
	uptime := r.Group("/uptime")
	{
		// Get uptime for specific infrastructure
		uptime.GET("/:infrastructure_id", h.GetInfrastructureUptime)

		// Get uptime summary for user
		uptime.GET("/user/:user_id", h.GetUserUptimeSummary)

		// Get uptime by infrastructure type
		uptime.GET("/type/:type", h.GetUptimeByType)

		// Get overall uptime summary
		uptime.GET("/summary", h.GetOverallSummary)

		// Record a status change event (for internal use or webhook)
		uptime.POST("/event", h.RecordStatusEvent)
	}
}

// GetInfrastructureUptime returns uptime data for a specific infrastructure
// @Summary Get infrastructure uptime
// @Description Get uptime statistics for a specific infrastructure instance
// @Tags uptime
// @Param infrastructure_id path string true "Infrastructure ID"
// @Param from query string false "Start time (RFC3339)"
// @Param to query string false "End time (RFC3339)"
// @Param period query string false "Period (1h, 24h, 7d, 30d)" default(24h)
// @Success 200 {object} dto.APIResponse{data=dto.UptimeResponse}
// @Router /api/v1/uptime/{infrastructure_id} [get]
func (h *UptimeHandler) GetInfrastructureUptime(c *gin.Context) {
	infraID := c.Param("infrastructure_id")
	from, to := h.parseTimeRange(c)

	uptime, err := h.uptimeService.GetInfrastructureUptime(c.Request.Context(), infraID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get infrastructure uptime",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Infrastructure uptime retrieved successfully",
		Data:    uptime,
	})
}

// GetUserUptimeSummary returns uptime summary for all infrastructures of a user
// @Summary Get user uptime summary
// @Description Get aggregated uptime statistics for all infrastructures owned by a user
// @Tags uptime
// @Param user_id path string true "User ID"
// @Param from query string false "Start time (RFC3339)"
// @Param to query string false "End time (RFC3339)"
// @Param period query string false "Period (1h, 24h, 7d, 30d)" default(24h)
// @Success 200 {object} dto.APIResponse{data=dto.UptimeSummaryResponse}
// @Router /api/v1/uptime/user/{user_id} [get]
func (h *UptimeHandler) GetUserUptimeSummary(c *gin.Context) {
	userID := c.Param("user_id")
	from, to := h.parseTimeRange(c)

	summary, err := h.uptimeService.GetUserUptimeSummary(c.Request.Context(), userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get user uptime summary",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "User uptime summary retrieved successfully",
		Data:    summary,
	})
}

// GetUptimeByType returns uptime summary for infrastructures of a specific type
// @Summary Get uptime by infrastructure type
// @Description Get aggregated uptime statistics for all infrastructures of a specific type
// @Tags uptime
// @Param type path string true "Infrastructure type (postgres_cluster, nginx_cluster, dind, clickhouse)"
// @Param from query string false "Start time (RFC3339)"
// @Param to query string false "End time (RFC3339)"
// @Param period query string false "Period (1h, 24h, 7d, 30d)" default(24h)
// @Success 200 {object} dto.APIResponse{data=dto.UptimeSummaryResponse}
// @Router /api/v1/uptime/type/{type} [get]
func (h *UptimeHandler) GetUptimeByType(c *gin.Context) {
	infraType := c.Param("type")
	from, to := h.parseTimeRange(c)

	summary, err := h.uptimeService.GetUptimeByType(c.Request.Context(), infraType, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get uptime by type",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Uptime by type retrieved successfully",
		Data:    summary,
	})
}

// GetOverallSummary returns overall uptime summary for all infrastructures
// @Summary Get overall uptime summary
// @Description Get aggregated uptime statistics for all infrastructures
// @Tags uptime
// @Param from query string false "Start time (RFC3339)"
// @Param to query string false "End time (RFC3339)"
// @Param period query string false "Period (1h, 24h, 7d, 30d)" default(24h)
// @Success 200 {object} dto.APIResponse{data=dto.UptimeSummaryResponse}
// @Router /api/v1/uptime/summary [get]
func (h *UptimeHandler) GetOverallSummary(c *gin.Context) {
	from, to := h.parseTimeRange(c)

	summary, err := h.uptimeService.GetOverallUptimeSummary(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get overall uptime summary",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Overall uptime summary retrieved successfully",
		Data:    summary,
	})
}

// RecordStatusEvent records a status change event
// @Summary Record status change event
// @Description Record a status change event for uptime tracking
// @Tags uptime
// @Accept json
// @Param event body RecordEventRequest true "Status event"
// @Success 200 {object} dto.APIResponse
// @Router /api/v1/uptime/event [post]
func (h *UptimeHandler) RecordStatusEvent(c *gin.Context) {
	var req RecordEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	event := elasticsearch.UptimeEvent{
		InstanceID:     req.InstanceID,
		InstanceName:   req.InstanceName,
		UserID:         req.UserID,
		Type:           req.Type,
		Action:         req.Action,
		Status:         req.Status,
		PreviousStatus: req.PreviousStatus,
		Message:        req.Message,
		Metadata:       req.Metadata,
	}

	if req.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			event.Timestamp = t
		}
	}

	if err := h.uptimeService.RecordStatusChange(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to record status event",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Status event recorded successfully",
	})
}

// RecordEventRequest represents the request body for recording status events
type RecordEventRequest struct {
	InstanceID     string                 `json:"instance_id" binding:"required"`
	InstanceName   string                 `json:"instance_name"`
	UserID         string                 `json:"user_id" binding:"required"`
	Type           string                 `json:"type" binding:"required"`
	Action         string                 `json:"action" binding:"required"`
	Status         string                 `json:"status"`
	PreviousStatus string                 `json:"previous_status"`
	Timestamp      string                 `json:"timestamp"`
	Message        string                 `json:"message"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// parseTimeRange parses time range from query parameters
func (h *UptimeHandler) parseTimeRange(c *gin.Context) (time.Time, time.Time) {
	now := time.Now()
	to := now
	from := now.Add(-24 * time.Hour) // Default: last 24 hours

	// Parse 'period' parameter
	period := c.DefaultQuery("period", "24h")
	switch period {
	case "1h":
		from = now.Add(-1 * time.Hour)
	case "24h":
		from = now.Add(-24 * time.Hour)
	case "7d":
		from = now.Add(-7 * 24 * time.Hour)
	case "30d":
		from = now.Add(-30 * 24 * time.Hour)
	case "90d":
		from = now.Add(-90 * 24 * time.Hour)
	}

	// Override with explicit from/to if provided
	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	return from, to
}
