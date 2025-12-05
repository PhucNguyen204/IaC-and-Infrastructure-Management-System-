package http

import (
	"net/http"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
	"github.com/gin-gonic/gin"
)

type ClickHouseHandler struct {
	service services.IClickHouseService
}

func NewClickHouseHandler(service services.IClickHouseService) *ClickHouseHandler {
	return &ClickHouseHandler{service: service}
}

func (h *ClickHouseHandler) RegisterRoutes(rg *gin.RouterGroup) {
	clickhouse := rg.Group("/clickhouse")
	{
		clickhouse.POST("", h.Create)
		clickhouse.GET("", h.List)
		clickhouse.GET("/:id", h.Get)
		clickhouse.DELETE("/:id", h.Delete)

		// Query operations
		clickhouse.POST("/:id/query", h.ExecuteQuery)
		clickhouse.POST("/:id/insert", h.InsertData)

		// Table management
		clickhouse.POST("/:id/tables", h.CreateTable)
		clickhouse.GET("/:id/tables", h.ListTables)
	}
}

// Create godoc
// @Summary Create ClickHouse instance
// @Description Create a new ClickHouse database instance
// @Tags ClickHouse
// @Accept json
// @Produce json
// @Param request body dto.CreateClickHouseRequest true "Create request"
// @Success 201 {object} dto.ClickHouseResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/clickhouse [post]
func (h *ClickHouseHandler) Create(c *gin.Context) {
	var req dto.CreateClickHouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system"
	}

	resp, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Get godoc
// @Summary Get ClickHouse instance
// @Description Get ClickHouse instance details by ID
// @Tags ClickHouse
// @Produce json
// @Param id path string true "Instance ID"
// @Success 200 {object} dto.ClickHouseResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/clickhouse/{id} [get]
func (h *ClickHouseHandler) Get(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// List godoc
// @Summary List ClickHouse instances
// @Description Get all ClickHouse instances
// @Tags ClickHouse
// @Produce json
// @Success 200 {array} dto.ClickHouseResponse
// @Router /api/v1/clickhouse [get]
func (h *ClickHouseHandler) List(c *gin.Context) {
	instances, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instances)
}

// Delete godoc
// @Summary Delete ClickHouse instance
// @Description Delete a ClickHouse instance
// @Tags ClickHouse
// @Param id path string true "Instance ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Router /api/v1/clickhouse/{id} [delete]
func (h *ClickHouseHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ExecuteQuery godoc
// @Summary Execute SQL query
// @Description Execute a SQL query on ClickHouse
// @Tags ClickHouse
// @Accept json
// @Produce json
// @Param id path string true "Instance ID"
// @Param request body dto.ClickHouseQueryRequest true "Query request"
// @Success 200 {object} dto.ClickHouseQueryResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/clickhouse/{id}/query [post]
func (h *ClickHouseHandler) ExecuteQuery(c *gin.Context) {
	id := c.Param("id")

	var req dto.ClickHouseQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.ExecuteQuery(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// InsertData godoc
// @Summary Insert data
// @Description Insert data into a ClickHouse table
// @Tags ClickHouse
// @Accept json
// @Produce json
// @Param id path string true "Instance ID"
// @Param request body dto.ClickHouseInsertRequest true "Insert request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/clickhouse/{id}/insert [post]
func (h *ClickHouseHandler) InsertData(c *gin.Context) {
	id := c.Param("id")

	var req dto.ClickHouseInsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.InsertData(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data inserted successfully"})
}

// CreateTable godoc
// @Summary Create table
// @Description Create a new table in ClickHouse
// @Tags ClickHouse
// @Accept json
// @Produce json
// @Param id path string true "Instance ID"
// @Param request body dto.ClickHouseTableDef true "Table definition"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/clickhouse/{id}/tables [post]
func (h *ClickHouseHandler) CreateTable(c *gin.Context) {
	id := c.Param("id")

	var req dto.ClickHouseTableDef
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateTable(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Table created successfully"})
}

// ListTables godoc
// @Summary List tables
// @Description List all tables in ClickHouse database
// @Tags ClickHouse
// @Produce json
// @Param id path string true "Instance ID"
// @Success 200 {array} string
// @Failure 404 {object} map[string]string
// @Router /api/v1/clickhouse/{id}/tables [get]
func (h *ClickHouseHandler) ListTables(c *gin.Context) {
	id := c.Param("id")

	tables, err := h.service.ListTables(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tables": tables})
}
