package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/docker"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/kafka"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/repositories"
	"go.uber.org/zap"
)

type IPostgreSQLClusterService interface {
	CreateCluster(ctx context.Context, userID string, req dto.CreateClusterRequest) (*dto.ClusterInfoResponse, error)
	StartCluster(ctx context.Context, clusterID string) error
	StopCluster(ctx context.Context, clusterID string) error
	RestartCluster(ctx context.Context, clusterID string) error
	DeleteCluster(ctx context.Context, clusterID string) error
	GetClusterInfo(ctx context.Context, clusterID string) (*dto.ClusterInfoResponse, error)
	ScaleCluster(ctx context.Context, clusterID string, req dto.ScaleClusterRequest) error
	GetClusterStats(ctx context.Context, clusterID string) (*dto.ClusterStatsResponse, error)
	GetClusterLogs(ctx context.Context, clusterID string, tail string) (*dto.ClusterLogsResponse, error)
	PromoteReplica(ctx context.Context, clusterID, nodeID string) error
	GetReplicationStatus(ctx context.Context, clusterID string) (*dto.ReplicationStatusResponse, error)
}

type postgreSQLClusterService struct {
	infraRepo     repositories.IInfrastructureRepository
	clusterRepo   repositories.IPostgreSQLClusterRepository
	dockerSvc     docker.IDockerService
	kafkaProducer kafka.IKafkaProducer
	logger        logger.ILogger
}

func NewPostgreSQLClusterService(
	infraRepo repositories.IInfrastructureRepository,
	clusterRepo repositories.IPostgreSQLClusterRepository,
	dockerSvc docker.IDockerService,
	kafkaProducer kafka.IKafkaProducer,
	logger logger.ILogger,
) IPostgreSQLClusterService {
	return &postgreSQLClusterService{
		infraRepo:     infraRepo,
		clusterRepo:   clusterRepo,
		dockerSvc:     dockerSvc,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

// CreateCluster creates a PostgreSQL cluster with streaming replication
func (s *postgreSQLClusterService) CreateCluster(ctx context.Context, userID string, req dto.CreateClusterRequest) (*dto.ClusterInfoResponse, error) {
	s.logger.Info("creating PostgreSQL cluster", zap.String("name", req.ClusterName), zap.Int("nodes", req.NodeCount))

	if req.NodeCount < 1 {
		return nil, fmt.Errorf("node_count must be at least 1")
	}

	// Create infrastructure record
	infraID := uuid.New().String()
	infra := &entities.Infrastructure{
		ID:     infraID,
		Name:   req.ClusterName,
		Type:   entities.TypePostgreSQLCluster,
		Status: entities.StatusCreating,
		UserID: userID,
	}
	if err := s.infraRepo.Create(infra); err != nil {
		return nil, fmt.Errorf("failed to create infrastructure: %w", err)
	}

	// Create cluster record
	clusterID := uuid.New().String()
	cluster := &entities.PostgreSQLCluster{
		ID:               clusterID,
		InfrastructureID: infraID,
		NodeCount:        req.NodeCount,
		Version:          req.PostgreSQLVersion,
		DatabaseName:     "postgres",
		Username:         "postgres",
		Password:         req.PostgreSQLPassword,
		CPULimit:         req.CPUPerNode,
		MemoryLimit:      req.MemoryPerNode,
	}
	if err := s.clusterRepo.Create(cluster); err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	// Create dedicated network
	networkName := fmt.Sprintf("iaas-cluster-%s", clusterID)
	networkID, err := s.dockerSvc.CreateNetwork(ctx, networkName)
	if err != nil {
		s.updateInfraStatus(infraID, entities.StatusFailed)
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	cluster.NetworkID = networkID
	s.clusterRepo.Update(cluster)

	// Create primary node
	primaryNode, err := s.createPrimaryNode(ctx, cluster, req, networkName)
	if err != nil {
		s.cleanup(ctx, cluster, networkID)
		return nil, fmt.Errorf("failed to create primary: %w", err)
	}

	cluster.PrimaryNodeID = primaryNode.ID
	s.clusterRepo.Update(cluster)

	// Wait for primary to be ready
	time.Sleep(10 * time.Second)

	// Create replicas
	for i := 1; i < req.NodeCount; i++ {
		if _, err := s.createReplicaNode(ctx, cluster, req, i, networkName, primaryNode); err != nil {
			s.logger.Warn("failed to create replica", zap.Int("index", i), zap.Error(err))
		}
	}

	// Update status
	infra.Status = entities.StatusRunning
	s.infraRepo.Update(infra)

	s.publishEvent(ctx, "cluster.created", infraID, clusterID, string(entities.StatusRunning))
	s.logger.Info("cluster created", zap.String("cluster_id", clusterID))

	return s.GetClusterInfo(ctx, clusterID)
}

func (s *postgreSQLClusterService) createPrimaryNode(ctx context.Context, cluster *entities.PostgreSQLCluster, req dto.CreateClusterRequest, networkName string) (*entities.ClusterNode, error) {
	nodeID := uuid.New().String()
	containerName := fmt.Sprintf("iaas-pgcluster-%s-primary", cluster.ID)
	volumeName := fmt.Sprintf("iaas-pgdata-%s-primary", cluster.ID)

	if err := s.dockerSvc.CreateVolume(ctx, volumeName); err != nil {
		return nil, err
	}

	config := docker.ContainerConfig{
		Name:  containerName,
		Image: fmt.Sprintf("postgres:%s", req.PostgreSQLVersion),
		Env: []string{
			"POSTGRES_PASSWORD=" + req.PostgreSQLPassword,
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=postgres",
		},
		Ports:        map[string]string{"5432": "0"},
		Volumes:      map[string]string{volumeName: "/var/lib/postgresql/data"},
		Network:      networkName,
		NetworkAlias: "primary",
		Resources: docker.ResourceConfig{
			CPULimit:    req.CPUPerNode,
			MemoryLimit: req.MemoryPerNode,
		},
	}

	containerID, err := s.dockerSvc.CreateContainer(ctx, config)
	if err != nil {
		s.dockerSvc.RemoveVolume(ctx, volumeName)
		return nil, err
	}

	if err := s.dockerSvc.StartContainer(ctx, containerID); err != nil {
		s.dockerSvc.RemoveContainer(ctx, containerID)
		s.dockerSvc.RemoveVolume(ctx, volumeName)
		return nil, err
	}

	// Setup replication
	time.Sleep(5 * time.Second)
	setupCmds := []string{
		"psql", "-U", "postgres", "-c",
		"CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD 'replicator_pass';",
	}
	s.dockerSvc.ExecCommand(ctx, containerID, setupCmds)

	// Get assigned port
	inspect, _ := s.dockerSvc.InspectContainer(ctx, containerID)
	port := 5432
	if len(inspect.NetworkSettings.Ports["5432/tcp"]) > 0 {
		fmt.Sscanf(inspect.NetworkSettings.Ports["5432/tcp"][0].HostPort, "%d", &port)
	}

	node := &entities.ClusterNode{
		ID:          nodeID,
		ClusterID:   cluster.ID,
		ContainerID: containerID,
		Role:        "primary",
		Port:        port,
		VolumeID:    volumeName,
		IsHealthy:   true,
	}

	if err := s.clusterRepo.CreateNode(node); err != nil {
		return nil, err
	}

	return node, nil
}

func (s *postgreSQLClusterService) createReplicaNode(ctx context.Context, cluster *entities.PostgreSQLCluster, req dto.CreateClusterRequest, index int, networkName string, primary *entities.ClusterNode) (*entities.ClusterNode, error) {
	nodeID := uuid.New().String()
	containerName := fmt.Sprintf("iaas-pgcluster-%s-replica-%d", cluster.ID, index)
	volumeName := fmt.Sprintf("iaas-pgdata-%s-replica-%d", cluster.ID, index)

	if err := s.dockerSvc.CreateVolume(ctx, volumeName); err != nil {
		return nil, err
	}

	config := docker.ContainerConfig{
		Name:  containerName,
		Image: fmt.Sprintf("postgres:%s", req.PostgreSQLVersion),
		Env: []string{
			"POSTGRES_PASSWORD=" + req.PostgreSQLPassword,
			"PGDATA=/var/lib/postgresql/data/pgdata",
		},
		Ports:        map[string]string{"5432": "0"},
		Volumes:      map[string]string{volumeName: "/var/lib/postgresql/data"},
		Network:      networkName,
		NetworkAlias: fmt.Sprintf("replica-%d", index),
		Resources: docker.ResourceConfig{
			CPULimit:    req.CPUPerNode,
			MemoryLimit: req.MemoryPerNode,
		},
	}

	containerID, err := s.dockerSvc.CreateContainer(ctx, config)
	if err != nil {
		s.dockerSvc.RemoveVolume(ctx, volumeName)
		return nil, err
	}

	if err := s.dockerSvc.StartContainer(ctx, containerID); err != nil {
		s.dockerSvc.RemoveContainer(ctx, containerID)
		s.dockerSvc.RemoveVolume(ctx, volumeName)
		return nil, err
	}

	// Setup replication
	time.Sleep(3 * time.Second)
	baseBackupCmd := []string{
		"bash", "-c",
		fmt.Sprintf("PGPASSWORD=replicator_pass pg_basebackup -h primary -U replicator -D /var/lib/postgresql/data/pgdata -Fp -Xs -P -R"),
	}
	s.dockerSvc.ExecCommand(ctx, containerID, baseBackupCmd)

	// Restart to apply replication
	s.dockerSvc.RestartContainer(ctx, containerID)

	inspect, _ := s.dockerSvc.InspectContainer(ctx, containerID)
	port := 5432
	if len(inspect.NetworkSettings.Ports["5432/tcp"]) > 0 {
		fmt.Sscanf(inspect.NetworkSettings.Ports["5432/tcp"][0].HostPort, "%d", &port)
	}

	node := &entities.ClusterNode{
		ID:          nodeID,
		ClusterID:   cluster.ID,
		ContainerID: containerID,
		Role:        "replica",
		Port:        port,
		VolumeID:    volumeName,
		IsHealthy:   true,
	}

	if err := s.clusterRepo.CreateNode(node); err != nil {
		return nil, err
	}

	return node, nil
}

func (s *postgreSQLClusterService) GetClusterInfo(ctx context.Context, clusterID string) (*dto.ClusterInfoResponse, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	infra, _ := s.infraRepo.FindByID(cluster.InfrastructureID)
	nodes, _ := s.clusterRepo.ListNodes(clusterID)

	nodeInfos := make([]dto.ClusterNodeInfo, len(nodes))
	for i, node := range nodes {
		nodeInfos[i] = dto.ClusterNodeInfo{
			NodeID:           node.ID,
			NodeName:         fmt.Sprintf("node-%s", node.Role),
			ContainerID:      node.ContainerID,
			Role:             node.Role,
			Status:           "running",
			ReplicationDelay: int(node.ReplicationDelay),
			IsHealthy:        node.IsHealthy,
		}
	}

	return &dto.ClusterInfoResponse{
		ClusterID:         cluster.ID,
		InfrastructureID:  cluster.InfrastructureID,
		ClusterName:       infra.Name,
		PostgreSQLVersion: cluster.Version,
		Status:            string(infra.Status),
		HAProxyPort:       cluster.HAProxyPort,
		Nodes:             nodeInfos,
		CreatedAt:         cluster.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         cluster.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *postgreSQLClusterService) StartCluster(ctx context.Context, clusterID string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	infra, _ := s.infraRepo.FindByID(cluster.InfrastructureID)
	if infra.Status == entities.StatusRunning {
		return fmt.Errorf("cluster already running")
	}

	nodes, _ := s.clusterRepo.ListNodes(clusterID)

	// Start primary first
	for _, node := range nodes {
		if node.Role == "primary" {
			if err := s.dockerSvc.StartContainer(ctx, node.ContainerID); err != nil {
				return err
			}
			time.Sleep(5 * time.Second)
		}
	}

	// Start replicas
	for _, node := range nodes {
		if node.Role == "replica" {
			s.dockerSvc.StartContainer(ctx, node.ContainerID)
		}
	}

	infra.Status = entities.StatusRunning
	s.infraRepo.Update(infra)

	s.publishEvent(ctx, "cluster.started", cluster.InfrastructureID, clusterID, string(entities.StatusRunning))
	return nil
}

func (s *postgreSQLClusterService) StopCluster(ctx context.Context, clusterID string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	nodes, _ := s.clusterRepo.ListNodes(clusterID)

	// Stop replicas first
	for _, node := range nodes {
		if node.Role == "replica" {
			s.dockerSvc.StopContainer(ctx, node.ContainerID)
		}
	}

	// Stop primary
	for _, node := range nodes {
		if node.Role == "primary" {
			s.dockerSvc.StopContainer(ctx, node.ContainerID)
		}
	}

	infra, _ := s.infraRepo.FindByID(cluster.InfrastructureID)
	infra.Status = entities.StatusStopped
	s.infraRepo.Update(infra)

	s.publishEvent(ctx, "cluster.stopped", cluster.InfrastructureID, clusterID, string(entities.StatusStopped))
	return nil
}

func (s *postgreSQLClusterService) RestartCluster(ctx context.Context, clusterID string) error {
	if err := s.StopCluster(ctx, clusterID); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return s.StartCluster(ctx, clusterID)
}

func (s *postgreSQLClusterService) DeleteCluster(ctx context.Context, clusterID string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	// Stop if running
	s.StopCluster(ctx, clusterID)
	time.Sleep(2 * time.Second)

	nodes, _ := s.clusterRepo.ListNodes(clusterID)
	for _, node := range nodes {
		s.dockerSvc.RemoveContainer(ctx, node.ContainerID)
		s.dockerSvc.RemoveVolume(ctx, node.VolumeID)
		s.clusterRepo.DeleteNode(node.ID)
	}

	if cluster.NetworkID != "" {
		s.dockerSvc.RemoveNetwork(ctx, cluster.NetworkID)
	}

	s.clusterRepo.Delete(clusterID)
	s.infraRepo.Delete(cluster.InfrastructureID)

	s.publishEvent(ctx, "cluster.deleted", cluster.InfrastructureID, clusterID, "deleted")
	return nil
}

func (s *postgreSQLClusterService) ScaleCluster(ctx context.Context, clusterID string, req dto.ScaleClusterRequest) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	currentNodes, _ := s.clusterRepo.ListNodes(clusterID)
	currentCount := len(currentNodes)

	if req.NodeCount == currentCount {
		return nil
	}

	if req.NodeCount < currentCount {
		// Scale down - remove replicas
		toRemove := currentCount - req.NodeCount
		for i, node := range currentNodes {
			if node.Role == "replica" && toRemove > 0 {
				s.dockerSvc.StopContainer(ctx, node.ContainerID)
				s.dockerSvc.RemoveContainer(ctx, node.ContainerID)
				s.dockerSvc.RemoveVolume(ctx, node.VolumeID)
				s.clusterRepo.DeleteNode(node.ID)
				toRemove--
				if i == toRemove {
					break
				}
			}
		}
	} else {
		// Scale up - add replicas
		var primary *entities.ClusterNode
		for _, node := range currentNodes {
			if node.Role == "primary" {
				primary = &node
				break
			}
		}

		clusterReq := dto.CreateClusterRequest{
			PostgreSQLVersion:  cluster.Version,
			PostgreSQLPassword: cluster.Password,
			CPUPerNode:         cluster.CPULimit,
			MemoryPerNode:      cluster.MemoryLimit,
		}

		for i := currentCount; i < req.NodeCount; i++ {
			s.createReplicaNode(ctx, cluster, clusterReq, i, fmt.Sprintf("iaas-cluster-%s", clusterID), primary)
		}
	}

	cluster.NodeCount = req.NodeCount
	s.clusterRepo.Update(cluster)

	s.publishEvent(ctx, "cluster.scaled", cluster.InfrastructureID, clusterID, fmt.Sprintf("nodes=%d", req.NodeCount))
	return nil
}

func (s *postgreSQLClusterService) GetClusterStats(ctx context.Context, clusterID string) (*dto.ClusterStatsResponse, error) {
	nodes, err := s.clusterRepo.ListNodes(clusterID)
	if err != nil {
		return nil, err
	}

	nodeStats := make([]dto.NodeStats, 0, len(nodes))
	for _, node := range nodes {
		stats, err := s.dockerSvc.GetContainerStats(ctx, node.ContainerID)
		if err != nil {
			continue
		}

		nodeStats = append(nodeStats, dto.NodeStats{
			NodeName:          fmt.Sprintf("node-%s", node.Role),
			Role:              node.Role,
			CPUPercent:        0,
			MemoryPercent:     0,
			ActiveConnections: 0,
		})
		stats.Body.Close()
	}

	return &dto.ClusterStatsResponse{
		ClusterID:        clusterID,
		TotalConnections: 0,
		TotalDatabases:   0,
		TotalSizeMB:      0,
		Nodes:            nodeStats,
	}, nil
}

func (s *postgreSQLClusterService) GetClusterLogs(ctx context.Context, clusterID string, tail string) (*dto.ClusterLogsResponse, error) {
	nodes, err := s.clusterRepo.ListNodes(clusterID)
	if err != nil {
		return nil, err
	}

	tailNum := 100
	if tail != "" {
		if num, err := strconv.Atoi(tail); err == nil {
			tailNum = num
		}
	}

	var nodeLogs []dto.NodeLog
	for _, node := range nodes {
		logs, err := s.dockerSvc.GetContainerLogs(ctx, node.ContainerID, tailNum)
		if err == nil {
			nodeLogs = append(nodeLogs, dto.NodeLog{
				NodeName:  fmt.Sprintf("%s-%d", node.Role, node.Port),
				Timestamp: time.Now().Format(time.RFC3339),
				Logs:      strings.Join(logs, "\n"),
			})
		}
	}

	return &dto.ClusterLogsResponse{
		ClusterID: clusterID,
		Logs:      nodeLogs,
	}, nil
}

func (s *postgreSQLClusterService) updateInfraStatus(infraID string, status entities.InfrastructureStatus) {
	infra, err := s.infraRepo.FindByID(infraID)
	if err == nil {
		infra.Status = status
		s.infraRepo.Update(infra)
	}
}

func (s *postgreSQLClusterService) cleanup(ctx context.Context, cluster *entities.PostgreSQLCluster, networkID string) {
	nodes, _ := s.clusterRepo.ListNodes(cluster.ID)
	for _, node := range nodes {
		s.dockerSvc.RemoveContainer(ctx, node.ContainerID)
		s.dockerSvc.RemoveVolume(ctx, node.VolumeID)
	}
	s.dockerSvc.RemoveNetwork(ctx, networkID)
	s.updateInfraStatus(cluster.InfrastructureID, entities.StatusFailed)
}

func (s *postgreSQLClusterService) publishEvent(ctx context.Context, eventType, infraID, clusterID, status string) {
	event := kafka.InfrastructureEvent{
		InstanceID: clusterID,
		UserID:     "",
		Type:       eventType,
		Action:     status,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"infrastructure_id": infraID,
			"cluster_id":        clusterID,
		},
	}
	s.kafkaProducer.PublishEvent(ctx, event)
}

// PromoteReplica promotes a replica to primary (manual failover)
func (s *postgreSQLClusterService) PromoteReplica(ctx context.Context, clusterID, nodeID string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	// Get all nodes
	nodes, err := s.clusterRepo.ListNodes(clusterID)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	// Find current primary and target replica
	var currentPrimary, targetReplica *entities.ClusterNode
	for i, node := range nodes {
		if node.ID == nodeID {
			if node.Role == "primary" {
				return fmt.Errorf("node is already primary")
			}
			targetReplica = &nodes[i]
		}
		if node.Role == "primary" {
			currentPrimary = &nodes[i]
		}
	}

	if targetReplica == nil {
		return fmt.Errorf("target node not found")
	}

	if currentPrimary == nil {
		return fmt.Errorf("current primary not found")
	}

	// Step 1: Promote replica to primary using pg_promote()
	promoteCmd := []string{"psql", "-U", "postgres", "-c", "SELECT pg_promote()"}
	_, err = s.dockerSvc.ExecCommand(ctx, targetReplica.ContainerID, promoteCmd)
	if err != nil {
		return fmt.Errorf("failed to promote replica: %w", err)
	}

	// Wait for promotion to complete
	time.Sleep(3 * time.Second)

	// Step 2: Update database records
	targetReplica.Role = "primary"
	targetReplica.UpdatedAt = time.Now()
	s.clusterRepo.UpdateNode(targetReplica)

	currentPrimary.Role = "replica"
	currentPrimary.UpdatedAt = time.Now()
	s.clusterRepo.UpdateNode(currentPrimary)

	// Step 3: Restart old primary to reconfigure as replica
	s.dockerSvc.RestartContainer(ctx, currentPrimary.ContainerID)

	// Publish event
	s.publishEvent(ctx, "cluster.failover", cluster.InfrastructureID, clusterID,
		fmt.Sprintf("promoted_node:%s", nodeID))

	return nil
}

// GetReplicationStatus returns replication health status
func (s *postgreSQLClusterService) GetReplicationStatus(ctx context.Context, clusterID string) (*dto.ReplicationStatusResponse, error) {
	nodes, err := s.clusterRepo.ListNodes(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var primaryNode *entities.ClusterNode
	var replicas []dto.ReplicaStatus

	// Find primary node
	for i, node := range nodes {
		if node.Role == "primary" {
			primaryNode = &nodes[i]
			break
		}
	}

	if primaryNode == nil {
		return &dto.ReplicationStatusResponse{
			Primary:  "none",
			Replicas: []dto.ReplicaStatus{},
		}, nil
	}

	// Get replica status from database records
	for _, node := range nodes {
		if node.Role == "replica" {
			replicas = append(replicas, dto.ReplicaStatus{
				NodeName:   fmt.Sprintf("node-%d", node.Port),
				State:      "streaming",
				SyncState:  "async",
				LagBytes:   int(node.ReplicationDelay),
				LagSeconds: 0,
				IsHealthy:  node.IsHealthy,
			})
		}
	}

	return &dto.ReplicationStatusResponse{
		Primary:  fmt.Sprintf("primary-%d", primaryNode.Port),
		Replicas: replicas,
	}, nil
}
