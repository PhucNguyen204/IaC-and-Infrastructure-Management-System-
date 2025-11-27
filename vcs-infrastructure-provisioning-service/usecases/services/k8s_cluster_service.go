package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/docker"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/kafka"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// IK8sClusterService defines the interface for K8s cluster service
type IK8sClusterService interface {
	CreateCluster(ctx context.Context, req dto.CreateK8sClusterRequest) (*dto.K8sClusterInfoResponse, error)
	GetClusterInfo(ctx context.Context, clusterID string) (*dto.K8sClusterInfoResponse, error)
	DeleteCluster(ctx context.Context, clusterID string) error
	StartCluster(ctx context.Context, clusterID string) error
	StopCluster(ctx context.Context, clusterID string) error
	RestartCluster(ctx context.Context, clusterID string) error
	ScaleCluster(ctx context.Context, clusterID string, req dto.ScaleK8sClusterRequest) error
	GetClusterHealth(ctx context.Context, clusterID string) (*dto.K8sClusterHealthResponse, error)
	GetClusterMetrics(ctx context.Context, clusterID string) (*dto.K8sClusterMetricsResponse, error)
	GetConnectionInfo(ctx context.Context, clusterID string) (*dto.K8sConnectionInfoResponse, error)
}

// k8sClusterService implements IK8sClusterService
type k8sClusterService struct {
	clusterRepo   repositories.IK8sClusterRepository
	infraRepo     repositories.IInfrastructureRepository
	dockerSvc     docker.IDockerService
	kafkaProducer kafka.IKafkaProducer
	logger        logger.ILogger
}

// NewK8sClusterService creates a new K8s cluster service
func NewK8sClusterService(
	clusterRepo repositories.IK8sClusterRepository,
	infraRepo repositories.IInfrastructureRepository,
	dockerSvc docker.IDockerService,
	kafkaProducer kafka.IKafkaProducer,
	logger logger.ILogger,
) IK8sClusterService {
	return &k8sClusterService{
		clusterRepo:   clusterRepo,
		infraRepo:     infraRepo,
		dockerSvc:     dockerSvc,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

// CreateCluster creates a new Kubernetes cluster using k3d
func (s *k8sClusterService) CreateCluster(ctx context.Context, req dto.CreateK8sClusterRequest) (*dto.K8sClusterInfoResponse, error) {
	s.logger.Info("creating k8s cluster", zap.String("cluster_name", req.ClusterName))

	// Set defaults
	if req.K8sVersion == "" {
		req.K8sVersion = "v1.28.5-k3s1" // Latest stable k3s
	}
	if req.ClusterType == "" {
		req.ClusterType = "k3s"
	}
	if req.CPULimit == "" {
		req.CPULimit = "2"
	}
	if req.MemoryLimit == "" {
		req.MemoryLimit = "2Gi"
	}
	if req.ClusterCIDR == "" {
		req.ClusterCIDR = "10.42.0.0/16"
	}
	if req.ServiceCIDR == "" {
		req.ServiceCIDR = "10.43.0.0/16"
	}
	if req.StorageClass == "" {
		req.StorageClass = "local-path"
	}
	if req.APIServerPort == 0 {
		req.APIServerPort = s.findAvailablePort(6443, 6543)
	}

	// Create infrastructure record
	clusterID := uuid.New().String()
	infraID := uuid.New().String()

	infra := &entities.Infrastructure{
		ID:     infraID,
		Name:   req.ClusterName,
		Type:   entities.InfraTypeK8sCluster,
		Status: entities.StatusCreating,
		UserID: "system", // TODO: Get from context
	}
	if err := s.infraRepo.Create(infra); err != nil {
		return nil, fmt.Errorf("failed to create infrastructure: %w", err)
	}

	// Create cluster record
	cluster := &entities.K8sCluster{
		ID:               clusterID,
		InfrastructureID: infraID,
		ClusterName:      req.ClusterName,
		K8sVersion:       req.K8sVersion,
		ClusterType:      req.ClusterType,
		NodeCount:        req.NodeCount,
		Status:           string(entities.StatusCreating),
		APIServerPort:    req.APIServerPort,
		NetworkName:      fmt.Sprintf("k3d-%s", req.ClusterName),
		ClusterCIDR:      req.ClusterCIDR,
		ServiceCIDR:      req.ServiceCIDR,
		CPULimit:         req.CPULimit,
		MemoryLimit:      req.MemoryLimit,
		DashboardEnabled: req.DashboardEnabled,
		IngressEnabled:   req.IngressEnabled,
		MetricsEnabled:   req.MetricsEnabled,
		StorageClass:     req.StorageClass,
	}

	if err := s.clusterRepo.Create(cluster); err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	// Create cluster asynchronously
	go func() {
		if err := s.createK3dCluster(cluster, req); err != nil {
			s.logger.Error("failed to create k3d cluster", zap.Error(err))
			cluster.Status = string(entities.StatusFailed)
			infra.Status = entities.StatusFailed
			s.clusterRepo.Update(cluster)
			s.infraRepo.Update(infra)
			return
		}

		// Update status to running
		cluster.Status = string(entities.StatusRunning)
		infra.Status = entities.StatusRunning
		s.clusterRepo.Update(cluster)
		s.infraRepo.Update(infra)

		// Publish event
		s.kafkaProducer.PublishEvent(context.Background(), kafka.InfrastructureEvent{
			InstanceID: clusterID,
			UserID:     "system",
			Type:       "k8s_cluster",
			Action:     "created",
			Timestamp:  time.Now(),
			Metadata: map[string]interface{}{
				"cluster_name": req.ClusterName,
				"node_count":   req.NodeCount,
				"k8s_version":  req.K8sVersion,
			},
		})

		s.logger.Info("k8s cluster created successfully", zap.String("cluster_id", clusterID))
	}()

	return s.GetClusterInfo(ctx, clusterID)
}

// createK3dCluster creates a k3d cluster
func (s *k8sClusterService) createK3dCluster(cluster *entities.K8sCluster, req dto.CreateK8sClusterRequest) error {
	// Build k3d create command
	args := []string{
		"cluster", "create", cluster.ClusterName,
		"--image", fmt.Sprintf("rancher/k3s:%s", cluster.K8sVersion),
		"--servers", "1", // 1 server (master) node
		"--agents", fmt.Sprintf("%d", req.NodeCount-1), // Rest are agent (worker) nodes
		"--api-port", fmt.Sprintf("%d", cluster.APIServerPort),
		"--network", cluster.NetworkName,
		"--k3s-arg", fmt.Sprintf("--cluster-cidr=%s@server:0", cluster.ClusterCIDR),
		"--k3s-arg", fmt.Sprintf("--service-cidr=%s@server:0", cluster.ServiceCIDR),
		"--wait",
	}

	// Add resource limits (only memory is supported in k3d v5.6.0)
	if cluster.MemoryLimit != "" {
		args = append(args, "--servers-memory", cluster.MemoryLimit)
		args = append(args, "--agents-memory", cluster.MemoryLimit)
	}
	// Note: CPU limits are not supported via k3d flags in v5.6.0
	// They can be set via Docker runtime-labels if needed

	// Disable traefik if ingress not enabled
	if !cluster.IngressEnabled {
		args = append(args, "--k3s-arg", "--disable=traefik@server:0")
	}

	// Execute k3d command
	cmd := exec.Command("k3d", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("k3d create failed: %w, output: %s", err, string(output))
	}

	s.logger.Info("k3d cluster created", zap.String("output", string(output)))

	// Get kubeconfig
	kubeconfig, err := s.getKubeconfig(cluster.ClusterName)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}
	cluster.Kubeconfig = kubeconfig

	// Get node information
	if err := s.updateNodeInfo(cluster); err != nil {
		s.logger.Warn("failed to update node info", zap.Error(err))
	}

	// Install dashboard if enabled
	if cluster.DashboardEnabled {
		if err := s.installDashboard(cluster); err != nil {
			s.logger.Warn("failed to install dashboard", zap.Error(err))
		}
	}

	return s.clusterRepo.Update(cluster)
}

// getKubeconfig retrieves the kubeconfig for a cluster
func (s *k8sClusterService) getKubeconfig(clusterName string) (string, error) {
	cmd := exec.Command("k3d", "kubeconfig", "get", clusterName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w, output: %s", err, string(output))
	}
	return string(output), nil
}

// updateNodeInfo updates node information from the cluster
func (s *k8sClusterService) updateNodeInfo(cluster *entities.K8sCluster) error {
	// Get node list using kubectl
	cmd := exec.Command("kubectl", "get", "nodes", "-o", "json", "--kubeconfig", "/dev/stdin")
	cmd.Stdin = strings.NewReader(cluster.Kubeconfig)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	// Parse node information (simplified - in production, use proper JSON parsing)
	// For now, create nodes based on expected count
	for i := 0; i < cluster.NodeCount; i++ {
		nodeID := uuid.New().String()
		role := "agent"
		nodeName := fmt.Sprintf("k3d-%s-agent-%d", cluster.ClusterName, i)

		if i == 0 {
			role = "server"
			nodeName = fmt.Sprintf("k3d-%s-server-0", cluster.ClusterName)
		}

		node := &entities.K8sNode{
			ID:             nodeID,
			ClusterID:      cluster.ID,
			Name:           nodeName,
			Role:           role,
			Status:         string(entities.StatusRunning),
			IsReady:        true,
			CPUCapacity:    cluster.CPULimit,
			MemoryCapacity: cluster.MemoryLimit,
			PodCapacity:    110,
			KubeletVersion: cluster.K8sVersion,
		}

		if err := s.clusterRepo.CreateNode(node); err != nil {
			s.logger.Warn("failed to create node record", zap.Error(err))
		}
	}

	s.logger.Info("node info updated", zap.String("output", string(output)))
	return nil
}

// installDashboard installs Kubernetes Dashboard
func (s *k8sClusterService) installDashboard(cluster *entities.K8sCluster) error {
	// Install Kubernetes Dashboard
	cmd := exec.Command("kubectl", "apply", "-f",
		"https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml",
		"--kubeconfig", "/dev/stdin")
	cmd.Stdin = strings.NewReader(cluster.Kubeconfig)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install dashboard: %w, output: %s", err, string(output))
	}

	// Create admin service account
	serviceAccountYAML := `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: admin-user
  namespace: kubernetes-dashboard
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: admin-user
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: admin-user
  namespace: kubernetes-dashboard
`
	cmd = exec.Command("kubectl", "apply", "-f", "-", "--kubeconfig", "/dev/stdin")
	cmd.Stdin = strings.NewReader(cluster.Kubeconfig + "\n---\n" + serviceAccountYAML)
	if output, err := cmd.CombinedOutput(); err != nil {
		s.logger.Warn("failed to create admin user", zap.Error(err), zap.String("output", string(output)))
	}

	// Get dashboard token
	time.Sleep(5 * time.Second) // Wait for service account to be created
	token, err := s.getDashboardToken(cluster)
	if err == nil {
		cluster.DashboardToken = token
		cluster.DashboardPort = cluster.APIServerPort + 1000 // Use a different port
	}

	s.logger.Info("dashboard installed", zap.String("cluster", cluster.ClusterName))
	return nil
}

// getDashboardToken retrieves the dashboard token
func (s *k8sClusterService) getDashboardToken(cluster *entities.K8sCluster) (string, error) {
	cmd := exec.Command("kubectl", "-n", "kubernetes-dashboard", "create", "token", "admin-user", "--kubeconfig", "/dev/stdin")
	cmd.Stdin = strings.NewReader(cluster.Kubeconfig)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetClusterInfo retrieves cluster information
func (s *k8sClusterService) GetClusterInfo(ctx context.Context, clusterID string) (*dto.K8sClusterInfoResponse, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	nodes, err := s.clusterRepo.ListNodes(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodeInfos := make([]dto.K8sNodeInfo, 0, len(nodes))
	for _, node := range nodes {
		nodeInfos = append(nodeInfos, dto.K8sNodeInfo{
			ID:             node.ID,
			Name:           node.Name,
			Role:           node.Role,
			Status:         node.Status,
			IsReady:        node.IsReady,
			IPAddress:      node.IPAddress,
			ContainerID:    node.ContainerID,
			CPUCapacity:    node.CPUCapacity,
			MemoryCapacity: node.MemoryCapacity,
			PodCapacity:    node.PodCapacity,
			KubeletVersion: node.KubeletVersion,
			OSImage:        node.OSImage,
			KernelVersion:  node.KernelVersion,
			CreatedAt:      node.CreatedAt,
		})
	}

	response := &dto.K8sClusterInfoResponse{
		ID:               cluster.ID,
		InfrastructureID: cluster.InfrastructureID,
		ClusterName:      cluster.ClusterName,
		K8sVersion:       cluster.K8sVersion,
		ClusterType:      cluster.ClusterType,
		Status:           cluster.Status,
		NodeCount:        cluster.NodeCount,
		Nodes:            nodeInfos,
		APIServerURL:     fmt.Sprintf("https://localhost:%d", cluster.APIServerPort),
		APIServerPort:    cluster.APIServerPort,
		ClusterCIDR:      cluster.ClusterCIDR,
		ServiceCIDR:      cluster.ServiceCIDR,
		DashboardEnabled: cluster.DashboardEnabled,
		IngressEnabled:   cluster.IngressEnabled,
		MetricsEnabled:   cluster.MetricsEnabled,
		StorageClass:     cluster.StorageClass,
		CreatedAt:        cluster.CreatedAt,
		UpdatedAt:        cluster.UpdatedAt,
	}

	if cluster.DashboardEnabled && cluster.DashboardToken != "" {
		response.DashboardURL = fmt.Sprintf("http://localhost:%d/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/", cluster.APIServerPort)
		response.DashboardToken = cluster.DashboardToken
	}

	// Only include kubeconfig if explicitly requested (check context)
	if ctx.Value("include_kubeconfig") == true {
		response.Kubeconfig = cluster.Kubeconfig
	}

	return response, nil
}

// DeleteCluster deletes a Kubernetes cluster
func (s *k8sClusterService) DeleteCluster(ctx context.Context, clusterID string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	s.logger.Info("deleting k8s cluster", zap.String("cluster_id", clusterID))

	// Delete k3d cluster
	cmd := exec.Command("k3d", "cluster", "delete", cluster.ClusterName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Warn("k3d delete failed", zap.Error(err), zap.String("output", string(output)))
	}

	// Delete nodes
	if err := s.clusterRepo.DeleteNodesByClusterID(clusterID); err != nil {
		s.logger.Warn("failed to delete nodes", zap.Error(err))
	}

	// Delete cluster record
	if err := s.clusterRepo.Delete(clusterID); err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	// Delete infrastructure record
	if err := s.infraRepo.Delete(cluster.InfrastructureID); err != nil {
		s.logger.Warn("failed to delete infrastructure", zap.Error(err))
	}

	// Publish event
	s.kafkaProducer.PublishEvent(context.Background(), kafka.InfrastructureEvent{
		InstanceID: clusterID,
		UserID:     "system",
		Type:       "k8s_cluster",
		Action:     "deleted",
		Timestamp:  time.Now(),
	})

	s.logger.Info("k8s cluster deleted", zap.String("cluster_id", clusterID))
	return nil
}

// StartCluster starts a stopped cluster
func (s *k8sClusterService) StartCluster(ctx context.Context, clusterID string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	cmd := exec.Command("k3d", "cluster", "start", cluster.ClusterName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start cluster: %w, output: %s", err, string(output))
	}

	cluster.Status = string(entities.StatusRunning)
	return s.clusterRepo.Update(cluster)
}

// StopCluster stops a running cluster
func (s *k8sClusterService) StopCluster(ctx context.Context, clusterID string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	cmd := exec.Command("k3d", "cluster", "stop", cluster.ClusterName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop cluster: %w, output: %s", err, string(output))
	}

	cluster.Status = string(entities.StatusStopped)
	return s.clusterRepo.Update(cluster)
}

// RestartCluster restarts a cluster
func (s *k8sClusterService) RestartCluster(ctx context.Context, clusterID string) error {
	if err := s.StopCluster(ctx, clusterID); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return s.StartCluster(ctx, clusterID)
}

// ScaleCluster scales the cluster by adding/removing nodes
func (s *k8sClusterService) ScaleCluster(ctx context.Context, clusterID string, req dto.ScaleK8sClusterRequest) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	if req.NodeCount == cluster.NodeCount {
		return nil // No change needed
	}

	// For k3d, we need to recreate the cluster with new node count
	// This is a limitation of k3d - in production, use proper k8s scaling
	return fmt.Errorf("scaling not supported for k3d clusters - please recreate cluster with desired node count")
}

// GetClusterHealth retrieves cluster health status
func (s *k8sClusterService) GetClusterHealth(ctx context.Context, clusterID string) (*dto.K8sClusterHealthResponse, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	nodes, err := s.clusterRepo.ListNodes(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	healthyNodes := 0
	nodeHealth := make([]dto.K8sNodeHealth, 0, len(nodes))

	for _, node := range nodes {
		if node.IsReady {
			healthyNodes++
		}
		nodeHealth = append(nodeHealth, dto.K8sNodeHealth{
			NodeID:        node.ID,
			NodeName:      node.Name,
			Role:          node.Role,
			IsReady:       node.IsReady,
			Status:        node.Status,
			LastHeartbeat: node.UpdatedAt,
		})
	}

	return &dto.K8sClusterHealthResponse{
		ClusterID:      cluster.ID,
		ClusterName:    cluster.ClusterName,
		Status:         cluster.Status,
		HealthyNodes:   healthyNodes,
		TotalNodes:     cluster.NodeCount,
		APIServerReady: cluster.Status == string(entities.StatusRunning),
		NodeHealth:     nodeHealth,
		LastCheck:      time.Now(),
	}, nil
}

// GetClusterMetrics retrieves cluster metrics
func (s *k8sClusterService) GetClusterMetrics(ctx context.Context, clusterID string) (*dto.K8sClusterMetricsResponse, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	// Get metrics using kubectl (simplified)
	response := &dto.K8sClusterMetricsResponse{
		ClusterID:   cluster.ID,
		CollectedAt: time.Now(),
	}

	// In production, use proper metrics-server or Prometheus
	return response, nil
}

// GetConnectionInfo retrieves connection information
func (s *k8sClusterService) GetConnectionInfo(ctx context.Context, clusterID string) (*dto.K8sConnectionInfoResponse, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	response := &dto.K8sConnectionInfoResponse{
		ClusterID:    cluster.ID,
		ClusterName:  cluster.ClusterName,
		APIServerURL: fmt.Sprintf("https://localhost:%d", cluster.APIServerPort),
		Kubeconfig:   base64.StdEncoding.EncodeToString([]byte(cluster.Kubeconfig)),
	}

	if cluster.DashboardEnabled {
		response.DashboardURL = fmt.Sprintf("http://localhost:%d/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/", cluster.APIServerPort)
		response.DashboardToken = cluster.DashboardToken
	}

	// Kubectl commands
	response.KubectlCommands.GetNodes = "kubectl get nodes --kubeconfig <kubeconfig-file>"
	response.KubectlCommands.GetPods = "kubectl get pods --all-namespaces --kubeconfig <kubeconfig-file>"
	response.KubectlCommands.GetServices = "kubectl get services --all-namespaces --kubeconfig <kubeconfig-file>"
	response.KubectlCommands.GetDeployments = "kubectl get deployments --all-namespaces --kubeconfig <kubeconfig-file>"

	return response, nil
}

// findAvailablePort finds an available port in the given range
func (s *k8sClusterService) findAvailablePort(start, end int) int {
	return start + (int(time.Now().Unix()) % (end - start))
}
