package http

import (
	"net/http"

	"github.com/PhucNguyen204/vcs-healthcheck-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	logger logger.ILogger
}

func NewHealthHandler(logger logger.ILogger) *HealthHandler {
	return &HealthHandler{logger: logger}
}

func (h *HealthHandler) RegisterRoutes(r *gin.RouterGroup) {
	health := r.Group("/health")
	{
		health.GET("", h.HealthCheck)
		health.GET("/ready", h.ReadinessCheck)
		health.GET("/live", h.LivenessCheck)
	}
}

// HealthCheck returns overall service health
// @Summary Health check
// @Description Returns the health status of the service
// @Tags health
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "healthcheck-service",
	})
}

// ReadinessCheck returns whether the service is ready to accept traffic
// @Summary Readiness check
// @Description Returns whether the service is ready
// @Tags health
// @Success 200 {object} map[string]interface{}
// @Router /health/ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// TODO: Check Kafka, ES connections
	c.JSON(http.StatusOK, gin.H{
		"ready": true,
	})
}

// LivenessCheck returns whether the service is alive
// @Summary Liveness check
// @Description Returns whether the service is alive
// @Tags health
// @Success 200 {object} map[string]interface{}
// @Router /health/live [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"alive": true,
	})
}
