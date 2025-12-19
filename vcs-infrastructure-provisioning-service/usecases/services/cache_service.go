package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/redis/go-redis/v9"
)

type ICacheService interface {
	GetClusterInfo(ctx context.Context, clusterID string) (*dto.ClusterInfoResponse, bool)
	SetClusterInfo(ctx context.Context, clusterID string, info *dto.ClusterInfoResponse, ttl time.Duration) error
	GetClusterStats(ctx context.Context, clusterID string) (*dto.ClusterStatsResponse, bool)
	SetClusterStats(ctx context.Context, clusterID string, stats *dto.ClusterStatsResponse, ttl time.Duration) error
	GetReplicationStatus(ctx context.Context, clusterID string) (*dto.ReplicationStatusResponse, bool)
	SetReplicationStatus(ctx context.Context, clusterID string, status *dto.ReplicationStatusResponse, ttl time.Duration) error
	InvalidateCluster(ctx context.Context, clusterID string) error
	InvalidateClusterInfo(ctx context.Context, clusterID string) error
	IsAvailable() bool
	// Monitoring integration
	RegisterContainerForMonitoring(ctx context.Context, infraID, containerID string) error
	UnregisterContainerFromMonitoring(ctx context.Context, infraID string) error
	GetRedisClient() *redis.Client
}

type cacheService struct {
	redis   *redis.Client
	enabled bool
}

func NewCacheService(redis *redis.Client) ICacheService {
	enabled := redis != nil
	if enabled {
		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := redis.Ping(ctx).Err(); err != nil {
			enabled = false
		}
	}
	return &cacheService{redis: redis, enabled: enabled}
}

func (s *cacheService) IsAvailable() bool {
	return s.enabled && s.redis != nil
}

func (s *cacheService) GetClusterInfo(ctx context.Context, clusterID string) (*dto.ClusterInfoResponse, bool) {
	if !s.IsAvailable() {
		return nil, false
	}

	key := fmt.Sprintf("cluster:info:%s", clusterID)
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var info dto.ClusterInfoResponse
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, false
	}

	return &info, true
}

func (s *cacheService) SetClusterInfo(ctx context.Context, clusterID string, info *dto.ClusterInfoResponse, ttl time.Duration) error {
	if !s.IsAvailable() {
		return nil // Skip silently when Redis unavailable
	}

	key := fmt.Sprintf("cluster:info:%s", clusterID)
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, key, data, ttl).Err()
}

func (s *cacheService) GetClusterStats(ctx context.Context, clusterID string) (*dto.ClusterStatsResponse, bool) {
	if !s.IsAvailable() {
		return nil, false
	}

	key := fmt.Sprintf("cluster:stats:%s", clusterID)
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var stats dto.ClusterStatsResponse
	if err := json.Unmarshal([]byte(data), &stats); err != nil {
		return nil, false
	}

	return &stats, true
}

func (s *cacheService) SetClusterStats(ctx context.Context, clusterID string, stats *dto.ClusterStatsResponse, ttl time.Duration) error {
	if !s.IsAvailable() {
		return nil
	}

	key := fmt.Sprintf("cluster:stats:%s", clusterID)
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, key, data, ttl).Err()
}

func (s *cacheService) GetReplicationStatus(ctx context.Context, clusterID string) (*dto.ReplicationStatusResponse, bool) {
	if !s.IsAvailable() {
		return nil, false
	}

	key := fmt.Sprintf("cluster:replication:%s", clusterID)
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var status dto.ReplicationStatusResponse
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, false
	}

	return &status, true
}

func (s *cacheService) SetReplicationStatus(ctx context.Context, clusterID string, status *dto.ReplicationStatusResponse, ttl time.Duration) error {
	if !s.IsAvailable() {
		return nil
	}

	key := fmt.Sprintf("cluster:replication:%s", clusterID)
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, key, data, ttl).Err()
}

func (s *cacheService) InvalidateCluster(ctx context.Context, clusterID string) error {
	if !s.IsAvailable() {
		return nil
	}

	keys := []string{
		fmt.Sprintf("cluster:info:%s", clusterID),
		fmt.Sprintf("cluster:stats:%s", clusterID),
		fmt.Sprintf("cluster:replication:%s", clusterID),
	}

	return s.redis.Del(ctx, keys...).Err()
}

func (s *cacheService) InvalidateClusterInfo(ctx context.Context, clusterID string) error {
	if !s.IsAvailable() {
		return nil
	}

	key := fmt.Sprintf("cluster:info:%s", clusterID)
	return s.redis.Del(ctx, key).Err()
}

// RegisterContainerForMonitoring registers a container for monitoring service to collect metrics
func (s *cacheService) RegisterContainerForMonitoring(ctx context.Context, infraID, containerID string) error {
	if !s.IsAvailable() {
		return nil
	}

	// Set infra:container:{infraID} -> containerID (used by monitoring service to find containers)
	containerKey := fmt.Sprintf("infra:container:%s", infraID)
	if err := s.redis.Set(ctx, containerKey, containerID, 0).Err(); err != nil {
		return err
	}

	// Set initial status as healthy
	statusKey := fmt.Sprintf("infra:status:%s", containerID)
	return s.redis.Set(ctx, statusKey, "healthy", time.Hour).Err()
}

// UnregisterContainerFromMonitoring removes a container from monitoring
func (s *cacheService) UnregisterContainerFromMonitoring(ctx context.Context, infraID string) error {
	if !s.IsAvailable() {
		return nil
	}

	containerKey := fmt.Sprintf("infra:container:%s", infraID)
	containerID, err := s.redis.Get(ctx, containerKey).Result()
	if err == nil && containerID != "" {
		statusKey := fmt.Sprintf("infra:status:%s", containerID)
		s.redis.Del(ctx, statusKey)
	}
	return s.redis.Del(ctx, containerKey).Err()
}

// GetRedisClient returns the underlying Redis client
func (s *cacheService) GetRedisClient() *redis.Client {
	return s.redis
}
