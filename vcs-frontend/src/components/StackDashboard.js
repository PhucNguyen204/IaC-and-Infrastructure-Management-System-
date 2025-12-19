import React, { useState, useEffect, useCallback } from 'react';
import {
  Plus, RefreshCw, Search, Filter, Database, Globe, Box,
  Play, Square, RotateCcw, Trash2, Eye, Server, Clock,
  CheckCircle, AlertCircle, XCircle, Loader, Zap
} from 'lucide-react';
import { stackAPI } from '../api';
import Layout from './common/Layout';
import CreateStackModal from './CreateStackModal';
import toast from 'react-hot-toast';
import './StackDashboard.css';

const StackDashboard = ({ onLogout, onViewStack, onNavigate }) => {
  const [resources, setResources] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [filterStatus, setFilterStatus] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [actionLoading, setActionLoading] = useState({});

  const fetchResources = useCallback(async () => {
    try {
      const response = await stackAPI.getAll();
      const stackData = response.data?.data?.stacks || response.data?.stacks || [];

      // Flatten stacks to resources, but also show stacks without resources (in Creating state)
      const flatResources = [];
      if (Array.isArray(stackData)) {
        stackData.forEach(stack => {
          if (stack.resources && stack.resources.length > 0) {
            // Show each resource individually
            stack.resources.forEach(resource => {
              flatResources.push({
                ...resource,
                stackId: stack.id,
                environment: stack.environment,
                created_at: stack.created_at,
                // Inherit stack status if resource status is missing/pending
                status: resource.status || stack.status
              });
            });
          } else {
            // Stack has no resources yet - show the stack itself as a placeholder row
            // This happens during initial provisioning
            flatResources.push({
              resource_name: stack.name,
              resource_type: 'PROVISIONING', // Placeholder type
              stackId: stack.id,
              environment: stack.environment,
              created_at: stack.created_at,
              status: stack.status || 'creating'
            });
          }
        });
      }
      setResources(flatResources);
    } catch (error) {
      console.error('Error fetching resources:', error);
      toast.error('Failed to load infrastructure');
      setResources([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchResources();
    const interval = setInterval(fetchResources, 15000);
    return () => clearInterval(interval);
  }, [fetchResources]);

  const handleStackAction = async (stackId, action) => {
    setActionLoading(prev => ({ ...prev, [stackId]: action }));
    try {
      switch (action) {
        case 'start':
          await stackAPI.start(stackId);
          toast.success('Service started');
          break;
        case 'stop':
          await stackAPI.stop(stackId);
          toast.success('Service stopped');
          break;
        case 'restart':
          await stackAPI.restart(stackId);
          toast.success('Service restarting');
          break;
        case 'delete':
          if (window.confirm('Are you sure you want to delete this infrastructure? This action cannot be undone.')) {
            await stackAPI.delete(stackId);
            toast.success('Infrastructure deleted');
          }
          break;
        default:
          break;
      }
      fetchResources();
    } catch (error) {
      console.error(`Error ${action} stack:`, error);
      toast.error(`Failed to ${action} service`);
    } finally {
      setActionLoading(prev => ({ ...prev, [stackId]: null }));
    }
  };

  const filteredResources = resources.filter(res => {
    const matchesSearch = res.resource_name?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesType = filterType === 'all' || res.resource_type === filterType;
    const matchesStatus = filterStatus === 'all' || res.status === filterStatus;
    return matchesSearch && matchesType && matchesStatus;
  });

  const getStatusIcon = (status) => {
    switch (status?.toLowerCase()) {
      case 'running': return <CheckCircle size={16} className="status-icon running" />;
      case 'stopped': return <Square size={16} className="status-icon stopped" />;
      case 'creating':
      case 'updating': return <Loader size={16} className="status-icon creating spin" />;
      case 'failed': return <XCircle size={16} className="status-icon failed" />;
      default: return <AlertCircle size={16} className="status-icon unknown" />;
    }
  };

  const getResourceIcon = (type) => {
    switch (type?.toUpperCase()) {
      case 'POSTGRES_CLUSTER': return <Database size={16} className="resource-type-icon postgres" />;
      case 'NGINX_GATEWAY':
      case 'NGINX_CLUSTER': return <Globe size={16} className="resource-type-icon nginx" />;
      case 'DIND_ENVIRONMENT': return <Box size={16} className="resource-type-icon docker" />;
      case 'PROVISIONING': return <Loader size={16} className="resource-type-icon spin" />;
      default: return <Server size={16} className="resource-type-icon" />;
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
    });
  };

  return (
    <Layout onLogout={onLogout} activeTab="stacks" onNavigate={onNavigate}>
      <div className="stack-dashboard">
        <div className="toolbar">
          <div className="search-box">
            <Search size={20} />
            <input
              type="text"
              placeholder="Search services..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          <div className="filters">
            <select value={filterType} onChange={(e) => setFilterType(e.target.value)}>
              <option value="all">All Types</option>
              <option value="POSTGRES_CLUSTER">PostgreSQL</option>
              <option value="NGINX_CLUSTER">Nginx Cluster</option>
              <option value="DIND_ENVIRONMENT">Docker Sandbox</option>
            </select>
            <select value={filterStatus} onChange={(e) => setFilterStatus(e.target.value)}>
              <option value="all">All Status</option>
              <option value="running">Running</option>
              <option value="stopped">Stopped</option>
              <option value="creating">Creating</option>
              <option value="failed">Failed</option>
            </select>
          </div>
          <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
            <Plus size={20} />
            New Infrastructure
          </button>
        </div>

        {/* Resources Table */}
        {loading ? (
          <div className="loading-state">
            <Loader size={32} className="spin" />
            <p>Loading infrastructure...</p>
          </div>
        ) : filteredResources.length === 0 ? (
          <div className="empty-state">
            <Server size={48} />
            <h3>No infrastructure found</h3>
            <p>Create your first service to get started</p>
          </div>
        ) : (
          <div className="stacks-table-container">
            <table className="stacks-table">
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Name</th>
                  <th>Type</th>
                  <th>Environment</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredResources.map((res, idx) => (
                  <tr
                    key={`${res.stackId}-${idx}`}
                    className="stack-row"
                    onClick={() => onViewStack && onViewStack(res.stackId)}
                  >
                    <td className="status-cell">
                      {getStatusIcon(res.status)}
                      <span className="status-text">{res.status || 'Unknown'}</span>
                    </td>
                    <td className="name-cell">
                      <div className="stack-name">{res.resource_name}</div>
                    </td>
                    <td className="resources-cell">
                      <div className="resource-chip">
                        {getResourceIcon(res.resource_type)}
                        <span>{res.resource_type?.replace(/_/g, ' ')}</span>
                      </div>
                    </td>
                    <td className="env-cell">
                      <span className={`env-badge ${res.environment}`}>{res.environment}</span>
                    </td>
                    <td className="date-cell">
                      <div className="stack-meta">
                        <Clock size={14} />
                        <span>{formatDate(res.created_at)}</span>
                      </div>
                    </td>
                    <td className="actions-cell" onClick={(e) => e.stopPropagation()}>
                      <div className="stack-actions">
                        <button
                          className="btn-icon"
                          onClick={() => onViewStack && onViewStack(res.stackId)}
                          title="View Details"
                        >
                          <Eye size={18} />
                        </button>
                        <button
                          className="btn-icon danger"
                          title="Delete"
                          onClick={() => handleStackAction(res.stackId, 'delete')}
                        >
                          <Trash2 size={18} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <CreateStackModal
          isOpen={showCreateModal}
          onClose={() => setShowCreateModal(false)}
          onSuccess={() => {
            setShowCreateModal(false);
            fetchResources();
          }}
        />
      </div>
    </Layout>
  );
};

export default StackDashboard;
