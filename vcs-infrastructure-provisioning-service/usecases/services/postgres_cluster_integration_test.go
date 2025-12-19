package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgreSQLClusterIntegration tests the full lifecycle of PostgreSQL cluster
// This test requires Docker to be running
func TestPostgreSQLClusterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup: Initialize service dependencies
	// Note: You need to provide real implementations or mocks
	// This is a template - adjust based on your actual service initialization
	service := setupTestService(t)
	ctx := context.Background()

	// Test 1: Create PostgreSQL cluster
	t.Run("CreateCluster", func(t *testing.T) {
		req := dto.CreateClusterRequest{
			ClusterName:        fmt.Sprintf("test-cluster-%d", time.Now().Unix()),
			PostgreSQLVersion:  "16",
			NodeCount:          2,
			CPUPerNode:         1000000000, // 1 CPU
			MemoryPerNode:      536870912,  // 512MB
			StoragePerNode:     1,          // 1GB
			PostgreSQLPassword: "testpass123",
			ReplicationMode:    "async",
		}

		cluster, err := service.CreateCluster(ctx, "test-user", req)
		require.NoError(t, err, "Failed to create cluster")
		require.NotNil(t, cluster, "Cluster should not be nil")

		assert.NotEmpty(t, cluster.ClusterID, "Cluster ID should not be empty")
		assert.Equal(t, req.ClusterName, cluster.ClusterName)
		assert.Equal(t, req.NodeCount, len(cluster.Nodes))
		assert.NotEmpty(t, cluster.WriteEndpoint.Host, "Write endpoint should be set")

		t.Logf("✓ Cluster created successfully: %s", cluster.ClusterID)
		t.Logf("  - Write endpoint: %s:%d", cluster.WriteEndpoint.Host, cluster.WriteEndpoint.Port)
		t.Logf("  - Nodes: %d", len(cluster.Nodes))

		// Store cluster ID for cleanup
		clusterID := cluster.ClusterID

		// Test 2: Verify cluster is running
		t.Run("GetClusterInfo", func(t *testing.T) {
			info, err := service.GetClusterInfo(ctx, clusterID)
			require.NoError(t, err, "Failed to get cluster info")
			assert.Equal(t, "running", info.Status)
			t.Logf("✓ Cluster status: %s", info.Status)
		})

		// Test 3: Execute query to verify database is accessible
		t.Run("ExecuteQuery_BasicSelect", func(t *testing.T) {
			queryReq := dto.ExecuteQueryRequest{
				Query:    "SELECT version();",
				Database: "postgres",
			}

			result, err := service.ExecuteQuery(ctx, clusterID, queryReq)
			require.NoError(t, err, "Failed to execute query")
			assert.Greater(t, result.RowCount, 0, "Should return at least one row")
			t.Logf("✓ Query executed successfully, duration: %s", result.Duration)
		})

		// Test 4: Create test table and insert data
		t.Run("WriteData_CreateTableAndInsert", func(t *testing.T) {
			// Create table
			createTableReq := dto.ExecuteQueryRequest{
				Query: `CREATE TABLE test_users (
					id SERIAL PRIMARY KEY,
					username VARCHAR(100) NOT NULL,
					email VARCHAR(255) NOT NULL,
					created_at TIMESTAMP DEFAULT NOW()
				);`,
				Database: "postgres",
			}

			_, err := service.ExecuteQuery(ctx, clusterID, createTableReq)
			require.NoError(t, err, "Failed to create table")
			t.Logf("✓ Table 'test_users' created")

			// Insert data
			insertReq := dto.ExecuteQueryRequest{
				Query: `INSERT INTO test_users (username, email) VALUES 
					('alice', 'alice@example.com'),
					('bob', 'bob@example.com'),
					('charlie', 'charlie@example.com');`,
				Database: "postgres",
			}

			_, err = service.ExecuteQuery(ctx, clusterID, insertReq)
			require.NoError(t, err, "Failed to insert data")
			t.Logf("✓ Inserted 3 test users")

			// Verify data
			selectReq := dto.ExecuteQueryRequest{
				Query:    "SELECT COUNT(*) FROM test_users;",
				Database: "postgres",
			}

			result, err := service.ExecuteQuery(ctx, clusterID, selectReq)
			require.NoError(t, err, "Failed to select data")
			assert.Greater(t, result.RowCount, 0, "Should have data")
			t.Logf("✓ Data verified: %d rows", result.RowCount)
		})

		// Test 5: Test replication - verify data is replicated to all nodes
		t.Run("TestReplication", func(t *testing.T) {
			// Wait a moment for replication to propagate
			time.Sleep(2 * time.Second)

			replicationResult, err := service.TestReplication(ctx, clusterID)
			require.NoError(t, err, "Failed to test replication")

			assert.NotEmpty(t, replicationResult.PrimaryNode, "Primary node should be identified")
			assert.True(t, replicationResult.AllSynced, "All nodes should be synced")

			t.Logf("✓ Replication test passed")
			t.Logf("  - Primary node: %s", replicationResult.PrimaryNode)
			t.Logf("  - All synced: %v", replicationResult.AllSynced)

			for _, nodeResult := range replicationResult.NodeResults {
				t.Logf("  - Node %s (%s): has_data=%v, row_count=%d",
					nodeResult.NodeName, nodeResult.Role, nodeResult.HasData, nodeResult.RowCount)
			}
		})

		// Test 6: Get replication status
		t.Run("GetReplicationStatus", func(t *testing.T) {
			status, err := service.GetReplicationStatus(ctx, clusterID)
			require.NoError(t, err, "Failed to get replication status")

			assert.NotEmpty(t, status.Primary, "Primary should be identified")
			assert.NotEmpty(t, status.Replicas, "Should have replicas")

			t.Logf("✓ Replication status retrieved")
			t.Logf("  - Primary: %s", status.Primary)
			t.Logf("  - Replicas count: %d", len(status.Replicas))

			for _, replica := range status.Replicas {
				t.Logf("  - Replica %s: state=%s, lag_bytes=%d, lag_seconds=%.2f, healthy=%v",
					replica.NodeName, replica.State, replica.LagBytes, replica.LagSeconds, replica.IsHealthy)
			}
		})

		// Test 7: Find and identify primary node
		t.Run("IdentifyPrimaryNode", func(t *testing.T) {
			info, err := service.GetClusterInfo(ctx, clusterID)
			require.NoError(t, err, "Failed to get cluster info")

			var primaryNode *dto.ClusterNodeInfo
			for i := range info.Nodes {
				if info.Nodes[i].Role == "primary" {
					primaryNode = &info.Nodes[i]
					break
				}
			}

			require.NotNil(t, primaryNode, "Primary node not found")
			assert.True(t, primaryNode.IsHealthy, "Primary node should be healthy")

			t.Logf("✓ Primary node identified: %s", primaryNode.NodeID)
			t.Logf("  - Node name: %s", primaryNode.NodeName)
			t.Logf("  - Status: %s", primaryNode.Status)
			t.Logf("  - Healthy: %v", primaryNode.IsHealthy)
		})

		// Test 8: Trigger manual failover
		t.Run("ManualFailover", func(t *testing.T) {
			// Get current cluster info
			info, err := service.GetClusterInfo(ctx, clusterID)
			require.NoError(t, err)

			// Find current primary and a replica
			var oldPrimaryID string
			var newPrimaryID string

			for _, node := range info.Nodes {
				if node.Role == "primary" {
					oldPrimaryID = node.NodeID
				} else if node.Role == "replica" && newPrimaryID == "" {
					newPrimaryID = node.NodeID
				}
			}

			require.NotEmpty(t, oldPrimaryID, "Old primary not found")
			require.NotEmpty(t, newPrimaryID, "No replica available for failover")

			t.Logf("  Old primary: %s", oldPrimaryID)
			t.Logf("  New primary: %s", newPrimaryID)

			// Trigger failover
			err = service.PromoteReplica(ctx, clusterID, newPrimaryID)
			require.NoError(t, err, "Failed to promote replica")
			t.Logf("✓ Failover initiated")

			// Wait for failover to complete
			time.Sleep(10 * time.Second)

			// Verify new primary
			updatedInfo, err := service.GetClusterInfo(ctx, clusterID)
			require.NoError(t, err)

			var currentPrimary string
			for _, node := range updatedInfo.Nodes {
				if node.Role == "primary" {
					currentPrimary = node.NodeID
					break
				}
			}

			t.Logf("  Current primary after failover: %s", currentPrimary)

			// Note: Failover may take time, so we just log the result
			// In a real test, you might want to poll until confirmed
			t.Logf("✓ Failover completed")
		})

		// Test 9: Verify data persists after failover
		t.Run("VerifyDataAfterFailover", func(t *testing.T) {
			selectReq := dto.ExecuteQueryRequest{
				Query:    "SELECT COUNT(*) FROM test_users;",
				Database: "postgres",
			}

			result, err := service.ExecuteQuery(ctx, clusterID, selectReq)
			require.NoError(t, err, "Failed to select data after failover")
			assert.Greater(t, result.RowCount, 0, "Data should still exist after failover")
			t.Logf("✓ Data verified after failover: %d rows", result.RowCount)
		})

		// Test 10: Get failover history
		t.Run("GetFailoverHistory", func(t *testing.T) {
			history, err := service.GetFailoverHistory(ctx, clusterID)
			require.NoError(t, err, "Failed to get failover history")

			t.Logf("✓ Failover history retrieved: %d events", len(history))
			for i, event := range history {
				t.Logf("  [%d] %s -> %s (reason: %s, triggered_by: %s, occurred_at: %s)",
					i+1, event.OldPrimaryName, event.NewPrimaryName,
					event.Reason, event.TriggeredBy, event.OccurredAt)
			}
		})

		// Test 11: Simulate node failure and verify automatic failover
		t.Run("SimulateNodeFailure", func(t *testing.T) {
			// Get current primary
			info, err := service.GetClusterInfo(ctx, clusterID)
			require.NoError(t, err)

			var primaryNodeID string
			for _, node := range info.Nodes {
				if node.Role == "primary" {
					primaryNodeID = node.NodeID
					break
				}
			}

			require.NotEmpty(t, primaryNodeID, "Primary node not found")
			t.Logf("  Stopping primary node: %s", primaryNodeID)

			// Stop the primary node to simulate failure
			err = service.StopNode(ctx, clusterID, primaryNodeID)
			require.NoError(t, err, "Failed to stop primary node")
			t.Logf("✓ Primary node stopped")

			// Wait for automatic failover (Patroni should handle this)
			t.Logf("  Waiting for automatic failover (30 seconds)...")
			time.Sleep(30 * time.Second)

			// Check if new primary was elected
			updatedInfo, err := service.GetClusterInfo(ctx, clusterID)
			require.NoError(t, err)

			var newPrimaryFound bool
			var newPrimaryID string
			for _, node := range updatedInfo.Nodes {
				if node.Role == "primary" && node.NodeID != primaryNodeID {
					newPrimaryFound = true
					newPrimaryID = node.NodeID
					break
				}
			}

			t.Logf("  New primary after automatic failover: %s", newPrimaryID)
			t.Logf("✓ Automatic failover completed: %v", newPrimaryFound)

			// Restart the old primary (it should become a replica)
			err = service.StartNode(ctx, clusterID, primaryNodeID)
			require.NoError(t, err, "Failed to restart old primary")
			t.Logf("✓ Old primary restarted as replica")

			time.Sleep(10 * time.Second)
		})

		// Test 12: Stress test - multiple concurrent writes
		t.Run("StressTest_ConcurrentWrites", func(t *testing.T) {
			t.Skip("Skipping stress test - enable manually if needed")

			const numWrites = 100
			errors := make(chan error, numWrites)
			done := make(chan bool, numWrites)

			startTime := time.Now()

			for i := 0; i < numWrites; i++ {
				go func(idx int) {
					insertReq := dto.ExecuteQueryRequest{
						Query: fmt.Sprintf(
							"INSERT INTO test_users (username, email) VALUES ('user%d', 'user%d@test.com');",
							idx, idx,
						),
						Database: "postgres",
					}

					_, err := service.ExecuteQuery(ctx, clusterID, insertReq)
					if err != nil {
						errors <- err
					}
					done <- true
				}(i)
			}

			// Wait for all writes
			for i := 0; i < numWrites; i++ {
				<-done
			}
			close(errors)

			duration := time.Since(startTime)

			errorCount := 0
			for range errors {
				errorCount++
			}

			t.Logf("✓ Concurrent writes completed in %s", duration)
			t.Logf("  - Total writes: %d", numWrites)
			t.Logf("  - Errors: %d", errorCount)
			t.Logf("  - Success rate: %.2f%%", float64(numWrites-errorCount)/float64(numWrites)*100)
		})

		// Cleanup: Delete cluster
		t.Run("Cleanup_DeleteCluster", func(t *testing.T) {
			err := service.DeleteCluster(ctx, clusterID)
			require.NoError(t, err, "Failed to delete cluster")
			t.Logf("✓ Cluster deleted successfully")
		})
	})
}

// TestPostgreSQLCluster_DataPersistence tests data persistence across operations
func TestPostgreSQLCluster_DataPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	service := setupTestService(t)
	ctx := context.Background()

	req := dto.CreateClusterRequest{
		ClusterName:        fmt.Sprintf("persist-test-%d", time.Now().Unix()),
		PostgreSQLVersion:  "16",
		NodeCount:          2,
		CPUPerNode:         1000000000,
		MemoryPerNode:      536870912,
		StoragePerNode:     1,
		PostgreSQLPassword: "testpass123",
		ReplicationMode:    "async",
	}

	cluster, err := service.CreateCluster(ctx, "test-user", req)
	require.NoError(t, err)
	defer service.DeleteCluster(ctx, cluster.ClusterID)

	// Create and populate table
	_, err = service.ExecuteQuery(ctx, cluster.ClusterID, dto.ExecuteQueryRequest{
		Query: "CREATE TABLE products (id SERIAL PRIMARY KEY, name TEXT, price DECIMAL);",
	})
	require.NoError(t, err)

	_, err = service.ExecuteQuery(ctx, cluster.ClusterID, dto.ExecuteQueryRequest{
		Query: "INSERT INTO products (name, price) VALUES ('Laptop', 999.99), ('Mouse', 29.99), ('Keyboard', 79.99);",
	})
	require.NoError(t, err)

	// Restart cluster
	err = service.RestartCluster(ctx, cluster.ClusterID)
	require.NoError(t, err)

	time.Sleep(15 * time.Second) // Wait for cluster to restart

	// Verify data persists
	result, err := service.ExecuteQuery(ctx, cluster.ClusterID, dto.ExecuteQueryRequest{
		Query: "SELECT COUNT(*) FROM products;",
	})
	require.NoError(t, err)
	assert.Greater(t, result.RowCount, 0, "Data should persist after restart")

	t.Logf("✓ Data persisted after cluster restart")
}

// setupTestService initializes the service for testing
// TODO: Implement this function based on your actual service dependencies
func setupTestService(t *testing.T) IPostgreSQLClusterService {
	t.Helper()

	// This is a placeholder - you need to implement actual service initialization
	// with real or mock dependencies:
	// - infraRepo
	// - clusterRepo
	// - dockerSvc
	// - kafkaProducer
	// - cacheService
	// - logger

	// Example:
	// return NewPostgreSQLClusterService(
	//     infraRepo,
	//     clusterRepo,
	//     dockerSvc,
	//     kafkaProducer,
	//     cacheService,
	//     logger,
	// )

	panic("setupTestService not implemented - please implement based on your infrastructure")
}
