import React, { useState, useEffect } from 'react';
import { 
  ArrowLeft, 
  MoreVertical,
  Download,
  Copy
} from 'lucide-react';
import { stackAPI } from '../api';
import Layout from './common/Layout';
import StatusBadge from './common/StatusBadge';
import ResourceIcon from './common/ResourceIcon';
import StackOverviewTab from './stack-detail/StackOverviewTab';
import StackLogsTab from './stack-detail/StackLogsTab';
import StackConfigTab from './stack-detail/StackConfigTab';
import StackMetricsTab from './stack-detail/StackMetricsTab';
import ConnectionPanel from './cluster/ConnectionPanel';
import SQLConsole from './cluster/SQLConsole';
import NginxClusterPanel from './nginx/NginxClusterPanel';
import toast from 'react-hot-toast';
import './StackDetailPage.css';

const StackDetailPage = ({ stackId, onBack, onLogout }) => {
  const [stack, setStack] = useState(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');
  const [showActions, setShowActions] = useState(false);

  useEffect(() => {
    loadStack();
  }, [stackId]);

  const loadStack = async () => {
    try {
      setLoading(true);
      const response = await stackAPI.getById(stackId);
      // Backend returns { success, code, message, data: { ... } }
      const stackData = response.data?.data || response.data;
      console.log('Stack data loaded:', stackData);
      setStack(stackData);
    } catch (error) {
      console.error('Error loading stack:', error);
      console.error('Error response:', error.response?.data);
      toast.error('Failed to load stack details');
    } finally {
      setLoading(false);
    }
  };

  const handleAction = async (action) => {
    try {
      const loadingToast = toast.loading(`${action} stack...`);
      await stackAPI[action](stackId);
      toast.success(`Stack ${action} successful`, { id: loadingToast });
      loadStack();
    } catch (error) {
      toast.error(`Failed to ${action} stack`);
    }
    setShowActions(false);
  };

  // Helper functions for PostgreSQL cluster
  const hasPostgresCluster = () => {
    return stack?.resources?.some(r => 
      r.resource_type === 'POSTGRES_CLUSTER' || 
      r.resource_type === 'postgres_cluster'
    );
  };

  const getClusterResource = () => {
    return stack?.resources?.find(r => 
      r.resource_type === 'POSTGRES_CLUSTER' || 
      r.resource_type === 'postgres_cluster'
    );
  };

  const getClusterId = () => {
    const resource = getClusterResource();
    return resource?.outputs?.cluster_id || null;
  };

  const hasNginxCluster = () => {
    return stack?.resources?.some(r => 
      r.resource_type?.toUpperCase() === 'NGINX_CLUSTER'
    );
  };

  const getNginxClusterResource = () => {
    return stack?.resources?.find(r => r.resource_type?.toUpperCase() === 'NGINX_CLUSTER');
  };

  const getNginxClusterId = () => {
    const resource = getNginxClusterResource();
    return resource?.outputs?.cluster_id || resource?.infrastructure_id || null;
  };

  if (loading) {
    return (
      <Layout onLogout={onLogout} activeTab="stacks">
        <div className="loading-state">Loading stack details...</div>
      </Layout>
    );
  }

  if (!stack) {
    return (
      <Layout onLogout={onLogout} activeTab="stacks">
        <div className="error-state">Stack not found</div>
      </Layout>
    );
  }

  return (
    <Layout onLogout={onLogout} activeTab="stacks">
      <div className="stack-detail-page">
        {/* Header */}
        <div className="stack-detail-header">
          <button className="back-btn" onClick={onBack}>
            <ArrowLeft size={20} />
            Back to Stacks
          </button>

          <div className="stack-title-section">
            <div className="title-row">
              <ResourceIcon type="stack" size={32} />
              <h1>{stack.name}</h1>
              <StatusBadge status={stack.status} />
            </div>
            
            <p className="description">{stack.description || 'No description'}</p>
            
            <div className="meta-info">
              <div className="meta-item">
                <span className="label">Environment:</span>
                <span className="value">{stack.environment}</span>
              </div>
              {stack.tags && Object.keys(stack.tags).length > 0 && (
                <div className="meta-item">
                  <span className="label">Tags:</span>
                  <div className="tags">
                    {Object.entries(stack.tags).map(([key, value]) => (
                      <span key={key} className="tag">#{value}</span>
                    ))}
                  </div>
                </div>
              )}
              <div className="meta-item">
                <span className="label">Created:</span>
                <span className="value">{new Date(stack.created_at).toLocaleString()}</span>
              </div>
              <div className="meta-item">
                <span className="label">Updated:</span>
                <span className="value">{new Date(stack.updated_at).toLocaleString()}</span>
              </div>
          </div>
        </div>

        <div className="action-buttons">
          <div className="dropdown">
            <button 
              className="btn btn-icon" 
              onClick={() => setShowActions(!showActions)}
            >
              <MoreVertical size={18} />
            </button>
            {showActions && (
              <div className="dropdown-menu">
                <button onClick={() => handleAction('clone')}>
                  <Copy size={16} />
                  Clone Stack
                </button>
                <button onClick={() => handleAction('exportTemplate')}>
                  <Download size={16} />
                  Export Template
                </button>
              </div>
            )}
          </div>
        </div>
      </div>        {/* Tabs */}
        <div className="tabs-navigation">
          <button 
            className={`tab ${activeTab === 'overview' ? 'active' : ''}`}
            onClick={() => setActiveTab('overview')}
          >
            Overview
          </button>
          {hasPostgresCluster() && (
            <>
              <button 
                className={`tab ${activeTab === 'connect' ? 'active' : ''}`}
                onClick={() => setActiveTab('connect')}
              >
                Connect
              </button>
              <button 
                className={`tab ${activeTab === 'sql' ? 'active' : ''}`}
                onClick={() => setActiveTab('sql')}
              >
                SQL Console
              </button>
            </>
          )}
          {hasNginxCluster() && (
            <button 
              className={`tab ${activeTab === 'nginx' ? 'active' : ''}`}
              onClick={() => setActiveTab('nginx')}
            >
              Nginx Cluster
            </button>
          )}
          <button 
            className={`tab ${activeTab === 'monitoring' ? 'active' : ''}`}
            onClick={() => setActiveTab('monitoring')}
          >
            Monitoring
          </button>
          <button 
            className={`tab ${activeTab === 'logs' ? 'active' : ''}`}
            onClick={() => setActiveTab('logs')}
          >
            Logs
          </button>
          <button 
            className={`tab ${activeTab === 'config' ? 'active' : ''}`}
            onClick={() => setActiveTab('config')}
          >
            Configuration
          </button>
        </div>

        {/* Tab Content */}
        <div className="tab-content">
          {activeTab === 'overview' && <StackOverviewTab stack={stack} onRefresh={loadStack} />}
          {activeTab === 'connect' && getClusterId() && (
            <ConnectionPanel clusterId={getClusterId()} clusterInfo={getClusterResource()} />
          )}
          {activeTab === 'sql' && getClusterId() && (
            <SQLConsole clusterId={getClusterId()} />
          )}
          {activeTab === 'nginx' && getNginxClusterId() && (
            <NginxClusterPanel 
              clusterId={getNginxClusterId()} 
              resource={getNginxClusterResource()}
              onRefresh={loadStack}
            />
          )}
          {activeTab === 'monitoring' && (
            <StackMetricsTab stackId={stackId} resources={stack.resources || []} />
          )}
          {activeTab === 'logs' && <StackLogsTab stackId={stackId} />}
          {activeTab === 'config' && <StackConfigTab stack={stack} onRefresh={loadStack} />}
        </div>
      </div>
    </Layout>
  );
};

export default StackDetailPage;
