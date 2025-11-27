import React, { useState, useEffect } from 'react';
import { Plus, RefreshCw, Globe, Server, Activity } from 'lucide-react';
import { stackAPI, nginxClusterAPI } from '../../api';
import Layout from '../common/Layout';
import StatusBadge from '../common/StatusBadge';
import CreateNginxClusterModal from './CreateNginxClusterModal';
import NginxClusterPanel from './NginxClusterPanel';
import toast from 'react-hot-toast';
import './NginxClusterDashboard.css';

const NginxClusterDashboard = ({ onLogout }) => {
  const [clusters, setClusters] = useState([]);
  const [selectedCluster, setSelectedCluster] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    loadClusters();
    
    // Auto-refresh every 10 seconds
    const interval = setInterval(() => {
      loadClusters(true);
    }, 10000);
    
    return () => clearInterval(interval);
  }, []);

  const loadClusters = async (silent = false) => {
    try {
      if (!silent) setLoading(true);
      
      // Get all stacks and extract Nginx Cluster resources
      const response = await stackAPI.getAll();
      const stacksData = response.data?.data?.stacks || response.data?.stacks || [];
      
      const nginxClusters = [];
      
      for (const stack of stacksData) {
        const stackDetail = await stackAPI.getById(stack.id).catch(() => null);
        if (stackDetail?.data?.data) {
          const resources = stackDetail.data.data.resources || [];
          const nginxResources = resources.filter(r => r.resource_type === 'NGINX_CLUSTER');
          
          for (const resource of nginxResources) {
            if (resource.outputs?.cluster_id) {
              try {
                const clusterInfo = await nginxClusterAPI.getById(resource.outputs.cluster_id);
                nginxClusters.push({
                  ...clusterInfo.data.data,
                  stack_name: stack.name,
                  stack_id: stack.id,
                  resource_name: resource.resource_name,
                });
              } catch (err) {
                console.error('Failed to load cluster info:', err);
              }
            }
          }
        }
      }
      
      setClusters(nginxClusters);
    } catch (error) {
      console.error('Error loading nginx clusters:', error);
      if (!silent) {
        toast.error('Failed to load nginx clusters');
      }
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const handleRefresh = () => {
    setRefreshing(true);
    loadClusters();
  };

  const handleCreateSuccess = (cluster) => {
    toast.success('Nginx Cluster created successfully!');
    loadClusters();
    setShowCreateModal(false);
  };

  const handleClusterClick = (cluster) => {
    setSelectedCluster(cluster);
  };

  const handleBack = () => {
    setSelectedCluster(null);
    loadClusters();
  };

  if (selectedCluster) {
    return (
      <Layout onLogout={onLogout} activeTab="nginx">
        <div className="nginx-cluster-detail">
          <div className="detail-header">
            <button className="btn btn-secondary" onClick={handleBack}>
              ← Back to Clusters
            </button>
            <h2>{selectedCluster.cluster_name}</h2>
          </div>
          <NginxClusterPanel 
            clusterId={selectedCluster.id} 
            resource={null}
            onRefresh={handleBack}
          />
        </div>
      </Layout>
    );
  }

  return (
    <Layout onLogout={onLogout} activeTab="nginx">
      <div className="nginx-cluster-dashboard">
        <div className="dashboard-header">
          <div>
            <h1>Nginx HA Clusters</h1>
            <p>Manage high-availability Nginx clusters with Keepalived</p>
          </div>
          <div className="header-actions">
            <button 
              className="btn btn-secondary" 
              onClick={handleRefresh}
              disabled={refreshing}
            >
              <RefreshCw size={18} className={refreshing ? 'spin' : ''} />
              Refresh
            </button>
            <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
              <Plus size={20} />
              New Nginx Cluster
            </button>
          </div>
        </div>

        {loading ? (
          <div className="loading-state">
            <Activity size={32} className="spin" />
            <p>Loading clusters...</p>
          </div>
        ) : clusters.length === 0 ? (
          <div className="empty-state">
            <Globe size={64} />
            <h3>No Nginx Clusters</h3>
            <p>Create your first Nginx HA cluster to get started</p>
            <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
              <Plus size={20} />
              Create Nginx Cluster
            </button>
          </div>
        ) : (
          <div className="clusters-grid">
            {clusters.map(cluster => (
              <div 
                key={cluster.id} 
                className="cluster-card"
                onClick={() => handleClusterClick(cluster)}
              >
                <div className="cluster-card-header">
                  <div className="cluster-icon">
                    <Globe size={24} />
                  </div>
                  <div className="cluster-title">
                    <h3>{cluster.cluster_name}</h3>
                    <span className="cluster-stack">Stack: {cluster.stack_name}</span>
                  </div>
                  <StatusBadge status={cluster.status} />
                </div>

                <div className="cluster-card-body">
                  <div className="cluster-stat">
                    <Server size={16} />
                    <span>{cluster.node_count} Nodes</span>
                  </div>
                  
                  <div className="cluster-stat">
                    <Activity size={16} />
                    <span>{cluster.load_balance_mode}</span>
                  </div>

                  {cluster.virtual_ip && (
                    <div className="cluster-stat">
                      <Globe size={16} />
                      <span>VIP: {cluster.virtual_ip}</span>
                    </div>
                  )}
                </div>

                <div className="cluster-card-footer">
                  <div className="endpoint-info">
                    <span className="endpoint-label">HTTP:</span>
                    <code>{cluster.http_port}</code>
                  </div>
                  {cluster.https_port > 0 && (
                    <div className="endpoint-info">
                      <span className="endpoint-label">HTTPS:</span>
                      <code>{cluster.https_port}</code>
                    </div>
                  )}
                </div>

                {cluster.nodes && cluster.nodes.length > 0 && (
                  <div className="nodes-preview">
                    {cluster.nodes.map((node, idx) => (
                      <div key={idx} className={`node-indicator ${node.status}`} title={`${node.name} (${node.role})`}>
                        {node.role === 'master' ? 'M' : 'B'}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}

        {showCreateModal && (
          <CreateNginxClusterModal
            isOpen={showCreateModal}
            onClose={() => setShowCreateModal(false)}
            onSuccess={handleCreateSuccess}
          />
        )}
      </div>
    </Layout>
  );
};

export default NginxClusterDashboard;

