import React, { useState, useEffect } from 'react';
import { Plus, Search, Eye, Play, Square, RotateCcw, Trash2, MoreVertical, Clock, CheckCircle, XCircle, AlertCircle, Loader } from 'lucide-react';
import { stackAPI } from '../api';
import Layout from './common/Layout';
import StatusBadge from './common/StatusBadge';
import ResourceIcon from './common/ResourceIcon';
import CreateStackModal from './CreateStackModal';
import toast from 'react-hot-toast';
import './StackDashboard.css';

const StackDashboard = ({ onLogout, onViewStack }) => {
  const [stacks, setStacks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterEnvironment, setFilterEnvironment] = useState('all');
  const [filterStatus, setFilterStatus] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [actionLoading, setActionLoading] = useState({});

  useEffect(() => {
    loadStacks();
    
    // Auto-refresh every 10 seconds to update stack statuses
    const interval = setInterval(() => {
      loadStacks();
    }, 10000);
    
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const loadStacks = async () => {
    try {
      setLoading(true);
      const response = await stackAPI.getAll();
      const stacksData = response.data?.data?.stacks || response.data?.stacks || [];
      setStacks(stacksData);
    } catch (error) {
      if (error.response?.status !== 401) {
        console.error('Error loading stacks:', error);
        toast.error('Failed to load stacks');
      }
      setStacks([]);
    } finally {
      setLoading(false);
    }
  };

  const filteredStacks = stacks.filter(stack => {
    const matchesSearch = stack.name?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesEnv = filterEnvironment === 'all' || stack.environment === filterEnvironment;
    const matchesStatus = filterStatus === 'all' || stack.status === filterStatus;
    return matchesSearch && matchesEnv && matchesStatus;
  });

  const handleCreateStack = () => {
    setShowCreateModal(true);
  };

  const handleCreateSuccess = () => {
    loadStacks();
  };

  const handleStackAction = async (stackId, action) => {
    setActionLoading(prev => ({ ...prev, [stackId]: action }));
    try {
      switch (action) {
        case 'start':
          await stackAPI.start(stackId);
          toast.success('Stack started');
          break;
        case 'stop':
          await stackAPI.stop(stackId);
          toast.success('Stack stopped');
          break;
        case 'restart':
          await stackAPI.restart(stackId);
          toast.success('Stack restarting');
          break;
        case 'delete':
          if (window.confirm('Are you sure you want to delete this stack?')) {
            await stackAPI.delete(stackId);
            toast.success('Stack deleted');
          }
          break;
        default:
          break;
      }
      loadStacks();
    } catch (error) {
      toast.error(`Failed to ${action} stack`);
    } finally {
      setActionLoading(prev => ({ ...prev, [stackId]: null }));
    }
  };

  const getStatusIcon = (status) => {
    switch (status?.toLowerCase()) {
      case 'running':
        return <CheckCircle size={16} className="status-icon running" />;
      case 'stopped':
        return <Square size={16} className="status-icon stopped" />;
      case 'creating':
      case 'updating':
        return <Loader size={16} className="status-icon creating spin" />;
      case 'failed':
        return <XCircle size={16} className="status-icon failed" />;
      default:
        return <AlertCircle size={16} className="status-icon unknown" />;
    }
  };

  const getEnvBadgeClass = (env) => {
    switch (env?.toLowerCase()) {
      case 'production': return 'env-badge production';
      case 'staging': return 'env-badge staging';
      default: return 'env-badge development';
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <Layout onLogout={onLogout} activeTab="stacks">
      <div className="stack-dashboard">
        {/* Search and Filters */}
        <div className="toolbar">
          <div className="search-box">
            <Search size={20} />
            <input
              type="text"
              placeholder="Search stacks..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
          <div className="filters">
            <select 
              value={filterEnvironment} 
              onChange={(e) => setFilterEnvironment(e.target.value)}
            >
              <option value="all">All Environments</option>
              <option value="production">Production</option>
              <option value="staging">Staging</option>
              <option value="development">Development</option>
            </select>
            <select 
              value={filterStatus} 
              onChange={(e) => setFilterStatus(e.target.value)}
            >
              <option value="all">All Status</option>
              <option value="running">Running</option>
              <option value="stopped">Stopped</option>
              <option value="degraded">Degraded</option>
              <option value="failed">Failed</option>
            </select>
          </div>
          <button className="btn btn-primary" onClick={handleCreateStack}>
            <Plus size={20} />
            New Stack
          </button>
        </div>

        {/* Stacks Table */}
        {loading ? (
          <div className="loading-state">
            <Loader size={32} className="spin" />
            <p>Loading stacks...</p>
          </div>
        ) : filteredStacks.length === 0 ? (
          <div className="empty-state">
            <ResourceIcon type="stack" size={48} />
            <h3>No stacks found</h3>
            <p>Create your first stack to get started</p>
          </div>
        ) : (
          <div className="stacks-table-container">
            <table className="stacks-table">
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Name</th>
                  <th>Environment</th>
                  <th>Resources</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredStacks.map(stack => (
                  <tr 
                    key={stack.id} 
                    className="stack-row" 
                    onClick={() => onViewStack && onViewStack(stack.id)}
                  >
                    <td className="status-cell">
                      {getStatusIcon(stack.status)}
                      <span className="status-text">{stack.status || 'Unknown'}</span>
                    </td>
                    <td className="name-cell">
                      <div className="stack-name">{stack.name}</div>
                      {stack.description && (
                        <div className="stack-description">{stack.description}</div>
                      )}
                    </td>
                    <td className="env-cell">
                      <span className={getEnvBadgeClass(stack.environment)}>
                        {stack.environment || 'N/A'}
                      </span>
                    </td>
                    <td className="resources-cell">
                      {stack.resources?.length > 0 ? (
                        <div className="resources-row">
                          {stack.resources.slice(0, 3).map((resource, idx) => (
                            <div key={idx} className="resource-chip">
                              <ResourceIcon type={resource.resource_type} size={14} />
                              <span>{resource.resource_name || resource.resource_type?.replace(/_/g, ' ')}</span>
                            </div>
                          ))}
                          {stack.resources.length > 3 && (
                            <span className="more-resources">+{stack.resources.length - 3}</span>
                          )}
                        </div>
                      ) : (
                        <span className="no-resources">No resources</span>
                      )}
                    </td>
                    <td className="date-cell">
                      <div className="stack-meta">
                        <Clock size={14} />
                        <span>{formatDate(stack.created_at)}</span>
                      </div>
                    </td>
                    <td className="actions-cell" onClick={(e) => e.stopPropagation()}>
                      <div className="stack-actions">
                        <button
                          className="btn-icon"
                          onClick={() => onViewStack && onViewStack(stack.id)}
                          title="View Details"
                        >
                          <Eye size={18} />
                        </button>
                        <div className="action-dropdown">
                          <button className="btn-icon" title="More Actions">
                            <MoreVertical size={18} />
                          </button>
                          <div className="dropdown-menu">
                            <button
                              onClick={() => handleStackAction(stack.id, 'start')}
                              disabled={stack.status === 'running' || actionLoading[stack.id]}
                            >
                              <Play size={14} /> Start
                            </button>
                            <button
                              onClick={() => handleStackAction(stack.id, 'stop')}
                              disabled={stack.status === 'stopped' || actionLoading[stack.id]}
                            >
                              <Square size={14} /> Stop
                            </button>
                            <button
                              onClick={() => handleStackAction(stack.id, 'restart')}
                              disabled={actionLoading[stack.id]}
                            >
                              <RotateCcw size={14} /> Restart
                            </button>
                            <hr />
                            <button
                              className="danger"
                              onClick={() => handleStackAction(stack.id, 'delete')}
                              disabled={actionLoading[stack.id]}
                            >
                              <Trash2 size={14} /> Delete
                            </button>
                          </div>
                        </div>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <CreateStackModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={handleCreateSuccess}
      />
    </Layout>
  );
};

export default StackDashboard;
