package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/docker"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type IClickHouseService interface {
	Create(ctx context.Context, userID string, req dto.CreateClickHouseRequest) (*dto.ClickHouseResponse, error)
	Get(ctx context.Context, id string) (*dto.ClickHouseResponse, error)
	GetByName(ctx context.Context, name string) (*dto.ClickHouseResponse, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]dto.ClickHouseResponse, error)

	// Query operations
	ExecuteQuery(ctx context.Context, id string, req dto.ClickHouseQueryRequest) (*dto.ClickHouseQueryResponse, error)
	InsertData(ctx context.Context, id string, req dto.ClickHouseInsertRequest) error

	// Table management
	CreateTable(ctx context.Context, id string, tableDef dto.ClickHouseTableDef) error
	ListTables(ctx context.Context, id string) ([]string, error)
}

type clickHouseService struct {
	infraRepo      repositories.IInfrastructureRepository
	clickhouseRepo *repositories.ClickHouseRepository
	dockerSvc      docker.IDockerService
	logger         logger.ILogger
}

func NewClickHouseService(
	infraRepo repositories.IInfrastructureRepository,
	clickhouseRepo *repositories.ClickHouseRepository,
	dockerSvc docker.IDockerService,
	logger logger.ILogger,
) IClickHouseService {
	return &clickHouseService{
		infraRepo:      infraRepo,
		clickhouseRepo: clickhouseRepo,
		dockerSvc:      dockerSvc,
		logger:         logger,
	}
}

func (s *clickHouseService) Create(ctx context.Context, userID string, req dto.CreateClickHouseRequest) (*dto.ClickHouseResponse, error) {
	s.logger.Info("creating ClickHouse instance", zap.String("name", req.ClusterName))

	// Set defaults
	if req.Version == "" {
		req.Version = "latest"
	}
	if req.Username == "" {
		req.Username = "default"
	}
	if req.CPULimit == 0 {
		req.CPULimit = 2
	}
	if req.MemoryLimit == 0 {
		req.MemoryLimit = 2048 // 2GB
	}

	// Create infrastructure record
	infraID := uuid.New().String()
	infra := &entities.Infrastructure{
		ID:     infraID,
		Name:   req.ClusterName,
		Type:   entities.TypeClickHouse,
		Status: entities.StatusCreating,
		UserID: userID,
	}
	if err := s.infraRepo.Create(infra); err != nil {
		return nil, fmt.Errorf("failed to create infrastructure: %w", err)
	}

	// Use fixed ports with offset based on name hash
	httpPort := 8123
	nativePort := 9000

	// Create cluster record
	clusterID := uuid.New().String()
	cluster := &entities.ClickHouseCluster{
		ID:               clusterID,
		InfrastructureID: infraID,
		ClusterName:      req.ClusterName,
		Version:          req.Version,
		NodeCount:        1,
		Username:         req.Username,
		Password:         req.Password,
		DatabaseName:     req.Database,
		HTTPPort:         httpPort,
		NativePort:       nativePort,
		CPULimit:         req.CPULimit,
		MemoryLimit:      req.MemoryLimit,
		StorageSize:      req.StorageSize,
	}

	// Create Docker network
	networkName := fmt.Sprintf("iaas-clickhouse-%s", clusterID[:8])
	networkID, err := s.dockerSvc.CreateNetwork(ctx, networkName)
	if err != nil {
		s.logger.Warn("failed to create network, using default", zap.Error(err))
		networkName = "bridge"
	} else {
		cluster.NetworkID = networkID
	}

	// Create ClickHouse container
	containerName := fmt.Sprintf("iaas-clickhouse-%s", clusterID[:8])
	imageName := fmt.Sprintf("clickhouse/clickhouse-server:%s", req.Version)

	// Container config using existing interface
	containerConfig := docker.ContainerConfig{
		Name:    containerName,
		Image:   imageName,
		Network: networkName,
		Env: []string{
			fmt.Sprintf("CLICKHOUSE_USER=%s", req.Username),
			fmt.Sprintf("CLICKHOUSE_PASSWORD=%s", req.Password),
			fmt.Sprintf("CLICKHOUSE_DB=%s", req.Database),
			"CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1",
		},
		Ports: map[string]string{
			"8123": fmt.Sprintf("%d", httpPort),
			"9000": fmt.Sprintf("%d", nativePort),
		},
		Resources: docker.ResourceConfig{
			CPULimit:    req.CPULimit,
			MemoryLimit: req.MemoryLimit * 1024 * 1024, // MB to bytes
		},
	}

	// Create container
	containerID, err := s.dockerSvc.CreateContainer(ctx, containerConfig)
	if err != nil {
		infra.Status = entities.StatusFailed
		s.infraRepo.Update(infra)
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := s.dockerSvc.StartContainer(ctx, containerID); err != nil {
		infra.Status = entities.StatusFailed
		s.infraRepo.Update(infra)
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Create node record
	node := &entities.ClickHouseNode{
		ID:          uuid.New().String(),
		ClusterID:   clusterID,
		NodeName:    containerName,
		ContainerID: containerID,
		HTTPPort:    httpPort,
		NativePort:  nativePort,
		IsHealthy:   true,
	}

	// Save to database
	if err := s.clickhouseRepo.CreateCluster(cluster); err != nil {
		return nil, fmt.Errorf("failed to save cluster: %w", err)
	}
	if err := s.clickhouseRepo.CreateNode(node); err != nil {
		return nil, fmt.Errorf("failed to save node: %w", err)
	}

	// Wait for ClickHouse to be ready
	s.waitForClickHouse(ctx, containerName, httpPort, req.Username, req.Password, 60*time.Second)

	// Create initial tables if specified
	for _, tableDef := range req.Tables {
		if err := s.createTableInternal(ctx, containerName, httpPort, req.Username, req.Password, req.Database, tableDef); err != nil {
			s.logger.Warn("failed to create table", zap.String("table", tableDef.Name), zap.Error(err))
		}
	}

	// Update status to running
	infra.Status = entities.StatusRunning
	s.infraRepo.Update(infra)

	return s.buildResponse(cluster, node, containerName, networkName), nil
}

func (s *clickHouseService) Get(ctx context.Context, id string) (*dto.ClickHouseResponse, error) {
	cluster, err := s.clickhouseRepo.GetClusterByID(id)
	if err != nil {
		return nil, fmt.Errorf("clickhouse not found: %w", err)
	}

	nodes, err := s.clickhouseRepo.GetNodesByClusterID(id)
	if err != nil || len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	networkName := ""
	if cluster.NetworkID != "" {
		networkName = fmt.Sprintf("iaas-clickhouse-%s", id[:8])
	}

	return s.buildResponse(cluster, &nodes[0], nodes[0].NodeName, networkName), nil
}

func (s *clickHouseService) GetByName(ctx context.Context, name string) (*dto.ClickHouseResponse, error) {
	cluster, err := s.clickhouseRepo.GetClusterByName(name)
	if err != nil {
		return nil, fmt.Errorf("clickhouse not found: %w", err)
	}

	nodes, err := s.clickhouseRepo.GetNodesByClusterID(cluster.ID)
	if err != nil || len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	networkName := ""
	if cluster.NetworkID != "" {
		networkName = fmt.Sprintf("iaas-clickhouse-%s", cluster.ID[:8])
	}

	return s.buildResponse(cluster, &nodes[0], nodes[0].NodeName, networkName), nil
}

func (s *clickHouseService) Delete(ctx context.Context, id string) error {
	cluster, err := s.clickhouseRepo.GetClusterByID(id)
	if err != nil {
		return fmt.Errorf("clickhouse not found: %w", err)
	}

	// Get nodes
	nodes, _ := s.clickhouseRepo.GetNodesByClusterID(id)

	// Stop and remove containers
	for _, node := range nodes {
		if node.ContainerID != "" {
			s.dockerSvc.StopContainer(ctx, node.ContainerID)
			s.dockerSvc.RemoveContainer(ctx, node.ContainerID)
		}
	}

	// Remove network
	if cluster.NetworkID != "" {
		s.dockerSvc.RemoveNetwork(ctx, cluster.NetworkID)
	}

	// Delete from database
	s.clickhouseRepo.DeleteNodesByClusterID(id)
	s.clickhouseRepo.DeleteCluster(id)
	s.infraRepo.Delete(cluster.InfrastructureID)

	return nil
}

func (s *clickHouseService) List(ctx context.Context) ([]dto.ClickHouseResponse, error) {
	clusters, err := s.clickhouseRepo.ListClusters()
	if err != nil {
		return nil, err
	}

	var responses []dto.ClickHouseResponse
	for _, cluster := range clusters {
		nodes, _ := s.clickhouseRepo.GetNodesByClusterID(cluster.ID)
		if len(nodes) > 0 {
			networkName := ""
			if cluster.NetworkID != "" {
				networkName = fmt.Sprintf("iaas-clickhouse-%s", cluster.ID[:8])
			}
			responses = append(responses, *s.buildResponse(&cluster, &nodes[0], nodes[0].NodeName, networkName))
		}
	}

	return responses, nil
}

func (s *clickHouseService) ExecuteQuery(ctx context.Context, id string, req dto.ClickHouseQueryRequest) (*dto.ClickHouseQueryResponse, error) {
	cluster, err := s.clickhouseRepo.GetClusterByID(id)
	if err != nil {
		return nil, fmt.Errorf("clickhouse not found: %w", err)
	}

	nodes, err := s.clickhouseRepo.GetNodesByClusterID(id)
	if err != nil || len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	database := req.Database
	if database == "" {
		database = cluster.DatabaseName
	}

	start := time.Now()
	result, err := s.executeHTTPQuery(ctx, nodes[0].NodeName, nodes[0].HTTPPort, cluster.Username, cluster.Password, database, req.Query)
	elapsed := time.Since(start).Seconds() * 1000

	if err != nil {
		return &dto.ClickHouseQueryResponse{
			Success: false,
			Error:   err.Error(),
			Elapsed: elapsed,
		}, nil
	}

	// Parse result
	var data interface{}
	rowCount := 0
	if result != "" {
		lines := strings.Split(strings.TrimSpace(result), "\n")
		rowCount = len(lines)
		if rowCount > 0 {
			if strings.HasPrefix(result, "[") || strings.HasPrefix(result, "{") {
				json.Unmarshal([]byte(result), &data)
			} else {
				data = result
			}
		}
	}

	return &dto.ClickHouseQueryResponse{
		Success:  true,
		Data:     data,
		RowCount: rowCount,
		Elapsed:  elapsed,
	}, nil
}

func (s *clickHouseService) InsertData(ctx context.Context, id string, req dto.ClickHouseInsertRequest) error {
	cluster, err := s.clickhouseRepo.GetClusterByID(id)
	if err != nil {
		return fmt.Errorf("clickhouse not found: %w", err)
	}

	nodes, err := s.clickhouseRepo.GetNodesByClusterID(id)
	if err != nil || len(nodes) == 0 {
		return fmt.Errorf("no nodes found")
	}

	if len(req.Data) == 0 {
		return fmt.Errorf("no data to insert")
	}

	// Get columns from first row
	var columns []string
	for col := range req.Data[0] {
		columns = append(columns, col)
	}

	// Build values
	var valueRows []string
	for _, row := range req.Data {
		var values []string
		for _, col := range columns {
			val := row[col]
			switch v := val.(type) {
			case string:
				values = append(values, fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''")))
			case nil:
				values = append(values, "NULL")
			default:
				values = append(values, fmt.Sprintf("%v", v))
			}
		}
		valueRows = append(valueRows, fmt.Sprintf("(%s)", strings.Join(values, ", ")))
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		req.Table,
		strings.Join(columns, ", "),
		strings.Join(valueRows, ", "))

	_, err = s.executeHTTPQuery(ctx, nodes[0].NodeName, nodes[0].HTTPPort, cluster.Username, cluster.Password, cluster.DatabaseName, query)
	return err
}

func (s *clickHouseService) CreateTable(ctx context.Context, id string, tableDef dto.ClickHouseTableDef) error {
	cluster, err := s.clickhouseRepo.GetClusterByID(id)
	if err != nil {
		return fmt.Errorf("clickhouse not found: %w", err)
	}

	nodes, err := s.clickhouseRepo.GetNodesByClusterID(id)
	if err != nil || len(nodes) == 0 {
		return fmt.Errorf("no nodes found")
	}

	return s.createTableInternal(ctx, nodes[0].NodeName, nodes[0].HTTPPort, cluster.Username, cluster.Password, cluster.DatabaseName, tableDef)
}

func (s *clickHouseService) ListTables(ctx context.Context, id string) ([]string, error) {
	cluster, err := s.clickhouseRepo.GetClusterByID(id)
	if err != nil {
		return nil, fmt.Errorf("clickhouse not found: %w", err)
	}

	nodes, err := s.clickhouseRepo.GetNodesByClusterID(id)
	if err != nil || len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	result, err := s.executeHTTPQuery(ctx, nodes[0].NodeName, nodes[0].HTTPPort, cluster.Username, cluster.Password, cluster.DatabaseName, "SHOW TABLES")
	if err != nil {
		return nil, err
	}

	var tables []string
	for _, line := range strings.Split(strings.TrimSpace(result), "\n") {
		if line != "" {
			tables = append(tables, line)
		}
	}

	return tables, nil
}

// Helper functions
func (s *clickHouseService) buildResponse(cluster *entities.ClickHouseCluster, node *entities.ClickHouseNode, containerName, networkName string) *dto.ClickHouseResponse {
	return &dto.ClickHouseResponse{
		ClusterID:        cluster.ID,
		InfrastructureID: cluster.InfrastructureID,
		ClusterName:      cluster.ClusterName,
		Version:          cluster.Version,
		Status:           string(cluster.Infrastructure.Status),
		Database:         cluster.DatabaseName,
		Username:         cluster.Username,
		HTTPEndpoint: dto.ClickHouseEndpoint{
			Host:         "localhost",
			Port:         node.HTTPPort,
			InternalHost: containerName,
			InternalPort: 8123,
		},
		NativeEndpoint: dto.ClickHouseEndpoint{
			Host:         "localhost",
			Port:         node.NativePort,
			InternalHost: containerName,
			InternalPort: 9000,
		},
		ContainerID:   node.ContainerID,
		ContainerName: containerName,
		NetworkName:   networkName,
		CreatedAt:     cluster.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     cluster.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *clickHouseService) waitForClickHouse(ctx context.Context, host string, port int, username, password string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := s.executeHTTPQuery(ctx, host, port, username, password, "default", "SELECT 1")
		if err == nil {
			s.logger.Info("ClickHouse is ready", zap.Int("port", port))
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timeout waiting for ClickHouse")
}

func (s *clickHouseService) executeHTTPQuery(ctx context.Context, host string, port int, username, password, database, query string) (string, error) {
	// When running in Docker, use host.docker.internal to access containers via host ports
	// This is needed because provisioning-service is in a different Docker network
	queryHost := "host.docker.internal"
	url := fmt.Sprintf("http://%s:%d/?database=%s", queryHost, port, database)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(query))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("query failed: %s", string(body))
	}

	return string(body), nil
}

func (s *clickHouseService) createTableInternal(ctx context.Context, host string, port int, username, password, database string, tableDef dto.ClickHouseTableDef) error {
	var columnDefs []string
	for _, col := range tableDef.Columns {
		colDef := fmt.Sprintf("`%s` %s", col.Name, col.Type)
		if col.Nullable {
			colDef = fmt.Sprintf("`%s` Nullable(%s)", col.Name, col.Type)
		}
		if col.DefaultValue != "" {
			colDef += fmt.Sprintf(" DEFAULT %s", col.DefaultValue)
		}
		columnDefs = append(columnDefs, colDef)
	}

	engine := tableDef.Engine
	if engine == "" {
		engine = "MergeTree()"
	} else if !strings.Contains(engine, "(") {
		engine += "()"
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n) ENGINE = %s",
		tableDef.Name,
		strings.Join(columnDefs, ",\n"),
		engine)

	if len(tableDef.OrderBy) > 0 {
		query += fmt.Sprintf("\nORDER BY (%s)", strings.Join(tableDef.OrderBy, ", "))
	} else {
		query += "\nORDER BY tuple()"
	}

	if tableDef.PartitionBy != "" {
		query += fmt.Sprintf("\nPARTITION BY %s", tableDef.PartitionBy)
	}

	if tableDef.TTL != "" {
		query += fmt.Sprintf("\nTTL %s", tableDef.TTL)
	}

	s.logger.Info("creating ClickHouse table", zap.String("query", query))
	_, err := s.executeHTTPQuery(ctx, host, port, username, password, database, query)
	return err
}
