package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
)

type NginxHandler struct {
	nginxService services.INginxService
}

func NewNginxHandler(nginxService services.INginxService) *NginxHandler {
	return &NginxHandler{nginxService: nginxService}
}

func (h *NginxHandler) RegisterRoutes(r *gin.RouterGroup) {
	nginx := r.Group("/nginx")
	{
		nginx.POST("", h.CreateNginx)
		nginx.GET("/:id", h.GetNginxInfo)
		nginx.POST("/:id/start", h.StartNginx)
		nginx.POST("/:id/stop", h.StopNginx)
		nginx.POST("/:id/restart", h.RestartNginx)
		nginx.DELETE("/:id", h.DeleteNginx)
		nginx.PUT("/:id/config", h.UpdateNginxConfig)

		nginx.POST("/:id/domains", h.AddDomain)
		nginx.DELETE("/:id/domains/:domain", h.DeleteDomain)
		nginx.POST("/:id/routes", h.AddRoute)
		nginx.PUT("/:id/routes/:route_id", h.UpdateRoute)
		nginx.DELETE("/:id/routes/:route_id", h.DeleteRoute)
		nginx.POST("/:id/certificate", h.UploadCertificate)
		nginx.GET("/:id/certificate", h.GetCertificate)
		nginx.PUT("/:id/upstreams", h.UpdateUpstreams)
		nginx.GET("/:id/upstreams", h.GetUpstreams)
		nginx.POST("/:id/security", h.SetSecurityPolicy)
		nginx.GET("/:id/security", h.GetSecurityPolicy)
		nginx.DELETE("/:id/security", h.DeleteSecurityPolicy)
		nginx.GET("/:id/logs", h.GetLogs)
		nginx.GET("/:id/metrics", h.GetMetrics)
		nginx.GET("/:id/stats", h.GetStats)
	}
}

func (h *NginxHandler) CreateNginx(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{
			Success: false,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	var req dto.CreateNginxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	resp, err := h.nginxService.CreateNginx(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to create Nginx instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "CREATED",
		Message: "Nginx instance created successfully",
		Data:    resp,
	})
}

func (h *NginxHandler) GetNginxInfo(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.nginxService.GetNginxInfo(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get Nginx info",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Nginx info retrieved successfully",
		Data:    resp,
	})
}

func (h *NginxHandler) StartNginx(c *gin.Context) {
	id := c.Param("id")

	if err := h.nginxService.StartNginx(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to start Nginx instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Nginx instance started successfully",
	})
}

func (h *NginxHandler) StopNginx(c *gin.Context) {
	id := c.Param("id")

	if err := h.nginxService.StopNginx(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to stop Nginx instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Nginx instance stopped successfully",
	})
}

func (h *NginxHandler) RestartNginx(c *gin.Context) {
	id := c.Param("id")

	if err := h.nginxService.RestartNginx(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to restart Nginx instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Nginx instance restarted successfully",
	})
}

func (h *NginxHandler) DeleteNginx(c *gin.Context) {
	id := c.Param("id")

	if err := h.nginxService.DeleteNginx(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete Nginx instance",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Nginx instance deleted successfully",
	})
}

func (h *NginxHandler) UpdateNginxConfig(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateNginxConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.nginxService.UpdateNginxConfig(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update Nginx config",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Nginx config updated successfully",
	})
}

func (h *NginxHandler) AddDomain(c *gin.Context) {
	id := c.Param("id")
	var req dto.AddDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.nginxService.AddDomain(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to add domain",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Domain added successfully",
	})
}

func (h *NginxHandler) DeleteDomain(c *gin.Context) {
	id := c.Param("id")
	domain := c.Param("domain")

	if err := h.nginxService.DeleteDomain(c.Request.Context(), id, domain); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete domain",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Domain deleted successfully",
	})
}

func (h *NginxHandler) AddRoute(c *gin.Context) {
	id := c.Param("id")
	var req dto.AddRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	route, err := h.nginxService.AddRoute(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to add route",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "CREATED",
		Message: "Route added successfully",
		Data:    route,
	})
}

func (h *NginxHandler) UpdateRoute(c *gin.Context) {
	id := c.Param("id")
	routeID := c.Param("route_id")
	var req dto.UpdateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.nginxService.UpdateRoute(c.Request.Context(), id, routeID, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update route",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Route updated successfully",
	})
}

func (h *NginxHandler) DeleteRoute(c *gin.Context) {
	id := c.Param("id")
	routeID := c.Param("route_id")

	if err := h.nginxService.DeleteRoute(c.Request.Context(), id, routeID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete route",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Route deleted successfully",
	})
}

func (h *NginxHandler) UploadCertificate(c *gin.Context) {
	id := c.Param("id")
	var req dto.UploadCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.nginxService.UploadCertificate(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to upload certificate",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Certificate uploaded successfully",
	})
}

func (h *NginxHandler) GetCertificate(c *gin.Context) {
	id := c.Param("id")

	cert, err := h.nginxService.GetCertificate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get certificate",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Certificate retrieved successfully",
		Data:    cert,
	})
}

func (h *NginxHandler) UpdateUpstreams(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUpstreamsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.nginxService.UpdateUpstreams(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update upstreams",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Upstreams updated successfully",
	})
}

func (h *NginxHandler) GetUpstreams(c *gin.Context) {
	id := c.Param("id")

	upstreams, err := h.nginxService.GetUpstreams(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get upstreams",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Upstreams retrieved successfully",
		Data:    upstreams,
	})
}

func (h *NginxHandler) SetSecurityPolicy(c *gin.Context) {
	id := c.Param("id")
	var req dto.SecurityPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.nginxService.SetSecurityPolicy(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to set security policy",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Security policy set successfully",
	})
}

func (h *NginxHandler) GetSecurityPolicy(c *gin.Context) {
	id := c.Param("id")

	policy, err := h.nginxService.GetSecurityPolicy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get security policy",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Security policy retrieved successfully",
		Data:    policy,
	})
}

func (h *NginxHandler) DeleteSecurityPolicy(c *gin.Context) {
	id := c.Param("id")

	if err := h.nginxService.DeleteSecurityPolicy(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete security policy",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Security policy deleted successfully",
	})
}

func (h *NginxHandler) GetLogs(c *gin.Context) {
	id := c.Param("id")
	tail := 100
	if t := c.Query("tail"); t != "" {
		if parsed := c.GetInt("tail"); parsed != 0 {
			tail = parsed
		}
	}

	logs, err := h.nginxService.GetLogs(c.Request.Context(), id, tail)
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

func (h *NginxHandler) GetMetrics(c *gin.Context) {
	id := c.Param("id")

	metrics, err := h.nginxService.GetMetrics(c.Request.Context(), id)
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

func (h *NginxHandler) GetStats(c *gin.Context) {
	id := c.Param("id")

	stats, err := h.nginxService.GetStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to get stats",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SUCCESS",
		Message: "Stats retrieved successfully",
		Data:    stats,
	})
}
