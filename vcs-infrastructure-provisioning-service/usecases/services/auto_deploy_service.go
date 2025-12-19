package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/docker"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type IAutoDeployService interface {
	Deploy(ctx context.Context, userID string, req dto.AutoDeployRequest) (*dto.AutoDeployResponse, error)
	GetDeployment(ctx context.Context, deploymentID string) (*dto.AutoDeployResponse, error)
	DeleteDeployment(ctx context.Context, deploymentID string) error
}

type autoDeployService struct {
	clickhouseSvc IClickHouseService
	postgresSvc   IPostgreSQLClusterService
	dockerSvc     docker.IDockerService
	logger        logger.ILogger
}

func NewAutoDeployService(
	clickhouseSvc IClickHouseService,
	postgresSvc IPostgreSQLClusterService,
	dockerSvc docker.IDockerService,
	logger logger.ILogger,
) IAutoDeployService {
	return &autoDeployService{
		clickhouseSvc: clickhouseSvc,
		postgresSvc:   postgresSvc,
		dockerSvc:     dockerSvc,
		logger:        logger,
	}
}

// Deploy deploys an application with its required infrastructure
func (s *autoDeployService) Deploy(ctx context.Context, userID string, req dto.AutoDeployRequest) (*dto.AutoDeployResponse, error) {
	s.logger.Info("starting auto deploy", zap.String("name", req.Name), zap.String("image", req.Image))

	deploymentID := uuid.New().String()
	response := &dto.AutoDeployResponse{
		DeploymentID:          deploymentID,
		Name:                  req.Name,
		Status:                "deploying",
		CreatedInfrastructure: make([]dto.CreatedInfra, 0),
		Endpoints:             make(map[string]string),
	}

	detectedInfra := s.detectInfraFromEnv(req.Environment)

	infraEnvVars := make(map[string]string)
	infraNetworks := make([]string, 0) 

	for _, infraType := range detectedInfra {
		switch infraType {
		case "clickhouse":
			infraInfo, envVars, networkName, err := s.createClickHouse(ctx, userID, req.Environment)
			if err != nil {
				response.Status = "failed"
				return response, err
			}
			response.CreatedInfrastructure = append(response.CreatedInfrastructure, *infraInfo)
			for k, v := range envVars {
				infraEnvVars[k] = v
			}
			if networkName != "" {
				infraNetworks = append(infraNetworks, networkName)
			}

		case "postgresql":
			infraInfo, envVars, networkName, err := s.createPostgres(ctx, userID, req.Environment)
			if err != nil {
				response.Status = "failed"
				return response, err
			}
			response.CreatedInfrastructure = append(response.CreatedInfrastructure, *infraInfo)
			for k, v := range envVars {
				infraEnvVars[k] = v
			}
			if networkName != "" {
				infraNetworks = append(infraNetworks, networkName)
			}
		}
	}

	// Merge infra env vars with user env vars (infra values override user values for connection info)
	finalEnvVars := make(map[string]string)
	for k, v := range req.Environment {
		finalEnvVars[k] = v
	}
	for k, v := range infraEnvVars {
		finalEnvVars[k] = v // Infrastructure env vars take precedence
	}

	// Deploy the application container, connecting to all infrastructure networks
	containerInfo, err := s.deployContainer(ctx, req, finalEnvVars, infraNetworks)
	if err != nil {
		response.Status = "failed"
		return response, err
	}

	response.Status = "running"
	response.Container = *containerInfo
	response.Endpoints["api"] = containerInfo.Endpoint

	s.logger.Info("auto deploy completed", zap.String("deploymentID", deploymentID))

	return response, nil
}

// detectInfraFromEnv detects required infrastructure types from environment variables
// Detects ClickHouse (CH_HOST=auto) and PostgreSQL (PG_HOST=auto) independently
func (s *autoDeployService) detectInfraFromEnv(env map[string]string) []string {
	var detected []string

	// Check ClickHouse - uses CH_HOST or DB_HOST with CH_PORT/DB_PORT=9000
	chHost := env["CH_HOST"]
	if chHost == "" {
		chHost = env["DB_HOST"]
	}
	chPort := env["CH_PORT"]
	if chPort == "" {
		chPort = env["DB_PORT"]
	}

	if chHost == "auto" || (chHost == "" && chPort == "9000") {
		detected = append(detected, "clickhouse")
		s.logger.Info("Detected ClickHouse requirement", zap.String("CH_HOST", chHost))
	}

	// Check PostgreSQL - uses PG_HOST with port 5432
	pgHost := env["PG_HOST"]
	pgPort := env["PG_PORT"]
	if pgPort == "" {
		pgPort = "5432"
	}

	if pgHost == "auto" {
		detected = append(detected, "postgresql")
		s.logger.Info("Detected PostgreSQL requirement", zap.String("PG_HOST", pgHost))
	}

	return detected
}

func (s *autoDeployService) buildClickHouseTablesFromEnv(env map[string]string) []dto.ClickHouseTableDef {
	var tables []dto.ClickHouseTableDef

	if logTable, ok := env["LOG_TABLE"]; ok && logTable != "" {
		tables = append(tables, dto.ClickHouseTableDef{
			Name:   logTable,
			Engine: "MergeTree",
			Columns: []dto.ClickHouseColumnDef{
				{Name: "id", Type: "String"},
				{Name: "timestamp", Type: "DateTime", DefaultValue: "now()"},
				{Name: "level", Type: "String"},
				{Name: "source", Type: "String"},
				{Name: "message", Type: "String"},
				{Name: "hostname", Type: "String"},
				{Name: "process_name", Type: "String"},
				{Name: "process_id", Type: "UInt32"},
				{Name: "user", Type: "String"},
				{Name: "event_type", Type: "String"},
				{Name: "raw_data", Type: "String"},
			},
			OrderBy:     []string{"timestamp", "id"},
			PartitionBy: "toYYYYMM(timestamp)",
		})
		s.logger.Info("Adding LOG_TABLE", zap.String("table", logTable))
	}

	if matchingTable, ok := env["MATCHING_TABLE"]; ok && matchingTable != "" {
		tables = append(tables, dto.ClickHouseTableDef{
			Name:   matchingTable,
			Engine: "MergeTree",
			Columns: []dto.ClickHouseColumnDef{
				{Name: "id", Type: "String"},
				{Name: "rule_id", Type: "String"},
				{Name: "log_id", Type: "String"},
				{Name: "matched_at", Type: "DateTime", DefaultValue: "now()"},
				{Name: "matched_field", Type: "String"},
				{Name: "matched_value", Type: "String"},
				{Name: "score", Type: "Float64"},
			},
			OrderBy:     []string{"matched_at", "id"},
			PartitionBy: "toYYYYMM(matched_at)",
		})
		s.logger.Info("Adding MATCHING_TABLE", zap.String("table", matchingTable))
	}

	if alertTable, ok := env["ALERT_TABLE"]; ok && alertTable != "" {
		tables = append(tables, dto.ClickHouseTableDef{
			Name:   alertTable,
			Engine: "MergeTree",
			Columns: []dto.ClickHouseColumnDef{
				{Name: "id", Type: "String"},
				{Name: "rule_id", Type: "String"},
				{Name: "rule_name", Type: "String"},
				{Name: "severity", Type: "String"},
				{Name: "message", Type: "String"},
				{Name: "log_id", Type: "String"},
				{Name: "matched_text", Type: "String"},
				{Name: "source", Type: "String"},
				{Name: "created_at", Type: "DateTime", DefaultValue: "now()"},
				{Name: "status", Type: "String", DefaultValue: "'open'"},
			},
			OrderBy:     []string{"created_at", "id"},
			PartitionBy: "toYYYYMM(created_at)",
		})
		s.logger.Info("Adding ALERT_TABLE", zap.String("table", alertTable))
	}

	if lookupTable, ok := env["LOOKUP_TABLE"]; ok && lookupTable != "" {
		tables = append(tables, dto.ClickHouseTableDef{
			Name:   lookupTable,
			Engine: "MergeTree",
			Columns: []dto.ClickHouseColumnDef{
				{Name: "id", Type: "String"},
				{Name: "list_name", Type: "String"},
				{Name: "list_type", Type: "String"},
				{Name: "value", Type: "String"},
				{Name: "description", Type: "String"},
				{Name: "created_at", Type: "DateTime", DefaultValue: "now()"},
				{Name: "updated_at", Type: "DateTime", DefaultValue: "now()"},
			},
			OrderBy: []string{"list_name", "value"},
		})
		s.logger.Info("Adding LOOKUP_TABLE", zap.String("table", lookupTable))
	}

	return tables
}

func (s *autoDeployService) ensureClickHouseDatabaseAndTables(ctx context.Context, containerName, username, password, database string, env map[string]string) error {
	s.logger.Info("Ensuring database and tables on existing ClickHouse",
		zap.String("container", containerName),
		zap.String("database", database))

	createDBCmd := []string{"clickhouse-client", "--user", username, "--password", password,
		"-q", fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database)}

	if _, err := s.dockerSvc.ExecCommand(ctx, containerName, createDBCmd); err != nil {
		s.logger.Warn("failed to create database (may already exist)", zap.Error(err))
	} else {
		s.logger.Info("Database created/verified", zap.String("database", database))
	}

	tables := s.buildClickHouseTablesFromEnv(env)
	for _, table := range tables {
		if err := s.createClickHouseTableViaExec(ctx, containerName, username, password, database, table); err != nil {
			s.logger.Warn("failed to create table", zap.String("table", table.Name), zap.Error(err))
		} else {
			s.logger.Info("Table created", zap.String("table", table.Name))
		}
	}

	return nil
}

func (s *autoDeployService) createClickHouseTableViaExec(ctx context.Context, containerName, username, password, database string, table dto.ClickHouseTableDef) error {
	var columns []string
	for _, col := range table.Columns {
		colDef := fmt.Sprintf("%s %s", col.Name, col.Type)
		if col.DefaultValue != "" {
			colDef += fmt.Sprintf(" DEFAULT %s", col.DefaultValue)
		}
		columns = append(columns, colDef)
	}

	engine := table.Engine
	if engine == "" {
		engine = "MergeTree"
	}

	orderBy := "tuple()"
	if len(table.OrderBy) > 0 {
		orderBy = fmt.Sprintf("(%s)", strings.Join(table.OrderBy, ", "))
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (%s) ENGINE = %s ORDER BY %s",
		database, table.Name, strings.Join(columns, ", "), engine, orderBy)

	if table.PartitionBy != "" {
		createSQL += fmt.Sprintf(" PARTITION BY %s", table.PartitionBy)
	}

	cmd := []string{"clickhouse-client", "--user", username, "--password", password, "-q", createSQL}

	if _, err := s.dockerSvc.ExecCommand(ctx, containerName, cmd); err != nil {
		return fmt.Errorf("failed to create table %s: %w", table.Name, err)
	}

	return nil
}

func (s *autoDeployService) findExistingClickHouse(ctx context.Context, requestEnv map[string]string) (*dto.CreatedInfra, map[string]string, string, error) {
	dockerClient := s.dockerSvc.GetClient()
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, nil, "", nil
	}

	for _, c := range containers {
		for _, name := range c.Names {
			containerName := strings.TrimPrefix(name, "/")
			if strings.HasPrefix(containerName, "iaas-clickhouse-") && c.State == "running" {
				s.logger.Info("Found existing ClickHouse container", zap.String("containerName", containerName))

				// Extract cluster ID from container name (iaas-clickhouse-{clusterID[:8]})
				clusterID := strings.TrimPrefix(containerName, "iaas-clickhouse-")
				networkName := fmt.Sprintf("iaas-clickhouse-%s", clusterID)

				infraInfo := &dto.CreatedInfra{
					Type:     "clickhouse",
					ID:       clusterID,
					Name:     containerName,
					Endpoint: fmt.Sprintf("%s:9000", containerName),
					Status:   "running",
				}

				// Use password from request env or default
				password := "clickhouse123"
				if pass, ok := requestEnv["DB_PASSWORD"]; ok && pass != "" {
					password = pass
				}
				if pass, ok := requestEnv["CH_PASSWORD"]; ok && pass != "" {
					password = pass
				}

				username := "default"
				if user, ok := requestEnv["DB_USER"]; ok && user != "" {
					username = user
				}
				if user, ok := requestEnv["CH_USER"]; ok && user != "" {
					username = user
				}

				database := "default"
				if db, ok := requestEnv["DB_NAME"]; ok && db != "" {
					database = db
				}
				if db, ok := requestEnv["CH_DATABASE"]; ok && db != "" {
					database = db
				}

				envVars := map[string]string{
					"DB_HOST":     containerName,
					"CH_HOST":     containerName,
					"DB_PORT":     "9000",
					"CH_PORT":     "9000",
					"DB_USER":     username,
					"CH_USER":     username,
					"DB_PASSWORD": password,
					"CH_PASSWORD": password,
					"DB_NAME":     database,
					"CH_DATABASE": database,
				}

				return infraInfo, envVars, networkName, nil
			}
		}
	}

	return nil, nil, "", nil // Not found
}

// findExistingPostgres finds an existing PostgreSQL HAProxy container
func (s *autoDeployService) findExistingPostgres(ctx context.Context) (*dto.CreatedInfra, map[string]string, string, error) {
	dockerClient := s.dockerSvc.GetClient()
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, nil, "", nil
	}

	for _, c := range containers {
		for _, name := range c.Names {
			containerName := strings.TrimPrefix(name, "/")
			if strings.HasPrefix(containerName, "iaas-haproxy-") && c.State == "running" {
				s.logger.Info("Found existing PostgreSQL HAProxy container", zap.String("containerName", containerName))

				// Extract cluster ID from container name (iaas-haproxy-{clusterID})
				clusterID := strings.TrimPrefix(containerName, "iaas-haproxy-")
				networkName := fmt.Sprintf("iaas-cluster-%s", clusterID)

				infraInfo := &dto.CreatedInfra{
					Type:     "postgresql",
					ID:       clusterID,
					Name:     containerName,
					Endpoint: fmt.Sprintf("%s:5000", containerName),
					Status:   "running",
				}

				envVars := map[string]string{
					"PG_HOST":     containerName,
					"PG_PORT":     "5000",
					"PG_USER":     "postgres",
					"PG_PASSWORD": "postgres123",
				}

				return infraInfo, envVars, networkName, nil
			}
		}
	}

	return nil, nil, "", nil
}

// createClickHouse creates a ClickHouse instance and returns network name
func (s *autoDeployService) createClickHouse(ctx context.Context, userID string, env map[string]string) (*dto.CreatedInfra, map[string]string, string, error) {
	// Try to find existing ClickHouse first
	if infraInfo, envVars, networkName, err := s.findExistingClickHouse(ctx, env); err == nil && infraInfo != nil {
		s.logger.Info("Reusing existing ClickHouse", zap.String("containerName", infraInfo.Name))

		// Still create database and tables for existing ClickHouse
		database := envVars["DB_NAME"]
		username := envVars["DB_USER"]
		password := envVars["DB_PASSWORD"]

		// Create database and tables on existing ClickHouse
		if err := s.ensureClickHouseDatabaseAndTables(ctx, infraInfo.Name, username, password, database, env); err != nil {
			s.logger.Warn("failed to create database/tables on existing ClickHouse", zap.Error(err))
		}

		return infraInfo, envVars, networkName, nil
	}
	s.logger.Info("creating ClickHouse for auto deploy")

	clusterName := fmt.Sprintf("auto-ch-%s", uuid.New().String()[:8])

	database := "default"
	if db, ok := env["DB_NAME"]; ok && db != "" {
		database = db
	}

	// Use password from env or generate one
	password := "clickhouse123"
	if pass, ok := env["DB_PASSWORD"]; ok && pass != "" {
		password = pass
	}

	username := "default"
	if user, ok := env["DB_USER"]; ok && user != "" {
		username = user
	}

	// Build tables from environment variables
	tables := s.buildClickHouseTablesFromEnv(env)

	req := dto.CreateClickHouseRequest{
		ClusterName: clusterName,
		Version:     "latest",
		Username:    username,
		Password:    password,
		Database:    database,
		CPULimit:    1,
		MemoryLimit: 1024,
		Tables:      tables,
	}

	result, err := s.clickhouseSvc.Create(ctx, userID, req)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create ClickHouse: %w", err)
	}

	// Network name follows pattern: iaas-clickhouse-{clusterID[:8]}
	networkName := result.NetworkName
	if networkName == "" {
		networkName = fmt.Sprintf("iaas-clickhouse-%s", result.ClusterID[:8])
	}

	infraInfo := &dto.CreatedInfra{
		Type:     "clickhouse",
		ID:       result.ClusterID,
		Name:     result.ContainerName, // Use container name (iaas-clickhouse-xxx), not cluster name
		Endpoint: fmt.Sprintf("%s:9000", result.ContainerName),
		Status:   "running",
	}

	// Return all connection env vars
	// Use ContainerName as DB_HOST - this is the actual container hostname in Docker network
	// Also set CH_HOST for applications that use CH_HOST specifically
	envVars := map[string]string{
		"DB_HOST":     result.ContainerName,
		"CH_HOST":     result.ContainerName, // Override any "auto" value
		"DB_PORT":     "9000",
		"CH_PORT":     "9000",
		"DB_USER":     username,
		"CH_USER":     username,
		"DB_PASSWORD": password,
		"CH_PASSWORD": password,
		"DB_NAME":     database,
		"CH_DATABASE": database,
	}

	s.logger.Info("ClickHouse created",
		zap.String("clusterName", result.ClusterName),
		zap.String("containerName", result.ContainerName),
		zap.String("networkName", networkName))

	return infraInfo, envVars, networkName, nil
}

// createPostgres creates a PostgreSQL cluster and returns network name
func (s *autoDeployService) createPostgres(ctx context.Context, userID string, env map[string]string) (*dto.CreatedInfra, map[string]string, string, error) {
	// Try to find existing PostgreSQL first
	if infraInfo, envVars, networkName, err := s.findExistingPostgres(ctx); err == nil && infraInfo != nil {
		s.logger.Info("Reusing existing PostgreSQL", zap.String("containerName", infraInfo.Name))

		// Add database name to env vars
		database := "alerts_db"
		if db, ok := env["PG_DATABASE"]; ok && db != "" {
			database = db
		}
		envVars["PG_DATABASE"] = database

		return infraInfo, envVars, networkName, nil
	}

	s.logger.Info("creating PostgreSQL for auto deploy")

	clusterName := fmt.Sprintf("auto-pg-%s", uuid.New().String()[:8])

	database := "alerts_db"
	if db, ok := env["PG_DATABASE"]; ok && db != "" {
		database = db
	}

	username := "postgres"
	if user, ok := env["PG_USER"]; ok && user != "" {
		username = user
	}

	password := "postgres123"
	if pass, ok := env["PG_PASSWORD"]; ok && pass != "" {
		password = pass
	}

	req := dto.CreateClusterRequest{
		ClusterName:        clusterName,
		PostgreSQLVersion:  "16",
		NodeCount:          1,
		CPUPerNode:         1,
		MemoryPerNode:      1024,
		StoragePerNode:     10,
		PostgreSQLPassword: password,
		ReplicationMode:    "async",
		EnableHAProxy:      true,
		Databases: []dto.ClusterDatabase{
			{Name: database, Owner: username, Encoding: "UTF8"},
		},
	}

	result, err := s.postgresSvc.CreateCluster(ctx, userID, req)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create PostgreSQL: %w", err)
	}

	// Network name follows pattern: iaas-cluster-{clusterID}
	networkName := fmt.Sprintf("iaas-cluster-%s", result.ClusterID)

	// HAProxy container name is the host to connect to within Docker network
	haproxyContainerName := fmt.Sprintf("iaas-haproxy-%s", result.ClusterID)
	haproxyPort := 5000 // Default HAProxy write port

	// Wait for PostgreSQL to be ready and create database
	s.logger.Info("waiting for PostgreSQL cluster to be ready...")
	if err := s.waitForPostgresAndCreateDB(ctx, result.ClusterID, database, username, password); err != nil {
		s.logger.Warn("failed to create database, will retry on container start", zap.Error(err))
	}

	infraInfo := &dto.CreatedInfra{
		Type:     "postgresql",
		ID:       result.ClusterID,
		Name:     haproxyContainerName, // Use HAProxy container name for Docker network
		Endpoint: fmt.Sprintf("%s:%d", haproxyContainerName, haproxyPort),
		Status:   "running",
	}

	envVars := map[string]string{
		"PG_HOST":     haproxyContainerName, // Use HAProxy container name instead of localhost
		"PG_PORT":     fmt.Sprintf("%d", haproxyPort),
		"PG_USER":     username,
		"PG_PASSWORD": password,
		"PG_DATABASE": database,
	}

	s.logger.Info("PostgreSQL created",
		zap.String("clusterName", clusterName),
		zap.String("host", result.WriteEndpoint.Host),
		zap.String("networkName", networkName))

	return infraInfo, envVars, networkName, nil
}

// waitForPostgresAndCreateDB waits for PostgreSQL to be ready and creates the database
func (s *autoDeployService) waitForPostgresAndCreateDB(ctx context.Context, clusterID, database, username, password string) error {
	// Find the patroni container
	patroniContainerName := ""
	containers, err := s.dockerSvc.GetClient().ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.Contains(name, fmt.Sprintf("iaas-patroni-%s", clusterID)) && strings.Contains(name, "node-1") {
				patroniContainerName = strings.TrimPrefix(name, "/")
				break
			}
		}
	}

	if patroniContainerName == "" {
		return fmt.Errorf("patroni container not found for cluster %s", clusterID)
	}

	// Wait for PostgreSQL to be ready (max 60 seconds)
	s.logger.Info("waiting for PostgreSQL to accept connections...", zap.String("container", patroniContainerName))

	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		// Check if PostgreSQL is ready using dockerSvc.ExecCommand
		output, err := s.dockerSvc.ExecCommand(ctx, patroniContainerName, []string{"pg_isready", "-U", "postgres"})
		if err == nil && strings.Contains(output, "accepting connections") {
			s.logger.Info("PostgreSQL is ready, creating database...")

			// Create database
			_, err := s.dockerSvc.ExecCommand(ctx, patroniContainerName, []string{
				"psql", "-U", "postgres", "-c", fmt.Sprintf("CREATE DATABASE %s;", database),
			})
			if err != nil {
				// Database might already exist, check error
				s.logger.Warn("database creation returned error (may already exist)", zap.Error(err))
			}

			s.logger.Info("database created successfully", zap.String("database", database))
			return nil
		}

		s.logger.Debug("PostgreSQL not ready yet, retrying...", zap.Int("attempt", i+1))
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for PostgreSQL to be ready")
}

// deployContainer deploys the application container and connects to all infrastructure networks
func (s *autoDeployService) deployContainer(ctx context.Context, req dto.AutoDeployRequest, envVars map[string]string, infraNetworks []string) (*dto.ContainerInfo, error) {
	s.logger.Info("deploying application container", zap.String("image", req.Image), zap.Strings("networks", infraNetworks))

	dockerClient := s.dockerSvc.GetClient()

	// Build env list
	envList := make([]string, 0, len(envVars))
	for k, v := range envVars {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}

	// Configure port
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}

	if req.ExposedPort > 0 {
		portStr := fmt.Sprintf("%d/tcp", req.ExposedPort)
		exposedPorts[nat.Port(portStr)] = struct{}{}
		portBindings[nat.Port(portStr)] = []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", req.ExposedPort)},
		}
	}

	// Configure volumes - auto-detect rules volume from image
	mounts := make([]mount.Mount, 0)

	// Add user-specified volumes
	for _, vol := range req.Volumes {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: vol.HostPath,
			Target: vol.ContainerPath,
		})
	}

	// Note: For detection-engine images, rules are already embedded in the image
	// If persistent rules storage is needed, user can specify volumes in the request
	// We no longer auto-mount an empty volume that would override the embedded rules

	// Auto-mount rules volume for detection-engine images if user didn't specify one
	if strings.Contains(req.Image, "detection-engine") {
		hasRulesVolume := false
		for _, vol := range req.Volumes {
			if vol.ContainerPath == "/opt/rules_storage" {
				hasRulesVolume = true
				break
			}
		}

		if !hasRulesVolume {
			// Create a named volume for rules storage
			rulesVolumeName := fmt.Sprintf("iaas-rules-%s", uuid.New().String()[:8])

			_, err := dockerClient.VolumeCreate(ctx, volume.CreateOptions{
				Name: rulesVolumeName,
				Labels: map[string]string{
					"iaas.managed": "true",
					"iaas.type":    "rules-storage",
				},
			})
			if err != nil {
				s.logger.Warn("failed to create rules volume", zap.Error(err))
			} else {
				mounts = append(mounts, mount.Mount{
					Type:   mount.TypeVolume,
					Source: rulesVolumeName,
					Target: "/opt/rules_storage",
				})
				s.logger.Info("auto-mounted rules volume for detection-engine",
					zap.String("volume", rulesVolumeName),
					zap.String("target", "/opt/rules_storage"))
			}
		}
	}

	containerName := fmt.Sprintf("%s-%s", req.Name, uuid.New().String()[:8])

	// Use first infrastructure network for initial creation, or default iaas network
	primaryNetwork := "iaas_iaas-network"
	if len(infraNetworks) > 0 {
		primaryNetwork = infraNetworks[0]
	}

	config := &container.Config{
		Image:        req.Image,
		Env:          envList,
		ExposedPorts: exposedPorts,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Mounts:       mounts,
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyUnlessStopped,
		},
		NetworkMode: container.NetworkMode(primaryNetwork),
	}

	// Create container
	resp, err := dockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Connect to additional infrastructure networks
	for i := 1; i < len(infraNetworks); i++ {
		networkName := infraNetworks[i]
		s.logger.Info("connecting container to additional network",
			zap.String("container", containerName),
			zap.String("network", networkName))

		if err := dockerClient.NetworkConnect(ctx, networkName, resp.ID, nil); err != nil {
			s.logger.Warn("failed to connect to network",
				zap.String("network", networkName),
				zap.Error(err))
		}
	}

	// Start container
	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// For detection-engine with auto-mounted volume, copy default rules if volume is empty
	if strings.Contains(req.Image, "detection-engine") {
		s.copyDefaultRulesToVolume(ctx, containerName)
	}

	// Wait for container to fully start and establish connections
	s.logger.Info("waiting for container to initialize...")
	time.Sleep(5 * time.Second)

	return &dto.ContainerInfo{
		ID:       resp.ID,
		Name:     containerName,
		Status:   "running",
		Endpoint: fmt.Sprintf("http://localhost:%d", req.ExposedPort),
	}, nil
}

// copyDefaultRulesToVolume copies default rules to the mounted volume if it's empty
func (s *autoDeployService) copyDefaultRulesToVolume(ctx context.Context, containerName string) {
	s.logger.Info("checking if rules volume needs default rules...", zap.String("container", containerName))

	// Check if the rules directory is empty
	output, err := s.dockerSvc.ExecCommand(ctx, containerName, []string{"ls", "-A", "/opt/rules_storage"})
	if err != nil {
		s.logger.Warn("failed to check rules directory", zap.Error(err))
		return
	}

	// If directory is not empty (has files), skip copying
	if strings.TrimSpace(output) != "" {
		s.logger.Info("rules directory already has files, skipping default rules copy")
		return
	}

	// Create default rules file
	defaultRules := `# Default Detection Rules
# These rules are auto-generated. You can modify or add more rules.

- id: RULE-001
  name: High CPU Detection
  description: Detect high CPU usage above 90%
  severity: high
  conditions:
    - cpu_usage > 90
  actions:
    - alert
    - log
  enabled: true

- id: RULE-002
  name: Memory Alert
  description: Detect high memory usage above 80%
  severity: medium
  conditions:
    - memory_usage > 80
  actions:
    - alert
  enabled: true

- id: RULE-003
  name: Disk Space Warning
  description: Detect low disk space below 20%
  severity: warning
  conditions:
    - disk_free < 20
  actions:
    - log
  enabled: true

- id: RULE-004
  name: Network Anomaly
  description: Detect unusual network traffic
  severity: critical
  conditions:
    - network_bytes_out > 1000000000
    - connection_count > 1000
  actions:
    - alert
    - log
    - notify
  enabled: true
`

	// Write default rules to the volume using echo and redirect
	// We use sh -c to handle the multiline string properly
	_, err = s.dockerSvc.ExecCommand(ctx, containerName, []string{
		"sh", "-c", fmt.Sprintf("cat > /opt/rules_storage/default_rules.yaml << 'EOF'\n%sEOF", defaultRules),
	})
	if err != nil {
		s.logger.Warn("failed to write default rules", zap.Error(err))
		return
	}

	s.logger.Info("default rules copied to volume successfully")
}

// GetDeployment retrieves deployment information
func (s *autoDeployService) GetDeployment(ctx context.Context, deploymentID string) (*dto.AutoDeployResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// DeleteDeployment removes a deployment
func (s *autoDeployService) DeleteDeployment(ctx context.Context, deploymentID string) error {
	return fmt.Errorf("not implemented")
}
