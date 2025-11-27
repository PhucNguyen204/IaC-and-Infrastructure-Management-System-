import React, { useState, useEffect, useCallback } from 'react';
import { 
  Plus, RefreshCw, Search, Filter, Database, Globe, Container, 
  Play, Square, RotateCcw, Trash2, Eye, Server, Clock, 
  CheckCircle, AlertCircle, XCircle, Loader, MoreVertical,
  ChevronRight, Layers
} from 'lucide-react';
import { stackAPI } from '../api';
import toast from 'react-hot-toast';
import CreateStackModal from './CreateStackModal';
import StackDetail from './StackDetail';
import './Dashboard.css';

function Dashboard({ onLogout }) {
  const [stacks, setStacks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterEnv, setFilterEnv] = useState('all');
  const [filterStatus, setFilterStatus] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedStack, setSelectedStack] = useState(null);
  const [actionLoading, setActionLoading] = useState({});

  const fetchStacks = useCallback(async () => {
    try {
      const response = await stackAPI.getAll();
      const stackData = response.data?.data?.stacks || response.data?.stacks || [];
      setStacks(Array.isArray(stackData) ? stackData : []);
    } catch (error) {
      console.error('Error fetching stacks:', error);
      toast.error('Failed to load stacks');
      setStacks([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStacks();
    const interval = setInterval(fetchStacks, 15000);
    return () => clearInterval(interval);
  }, [fetchStacks]);

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
          if (window.confirm('Are you sure you want to delete this stack? This action cannot be undone.')) {
            await stackAPI.delete(stackId);
            toast.success('Stack deleted');
          }
          break;
        default:
          break;
      }
      fetchStacks();
    } catch (error) {
      console.error(`Error ${action} stack:`, error);
      toast.error(`Failed to ${action} stack`);
    } finally {
      setActionLoading(prev => ({ ...prev, [stackId]: null }));
    }
  };

  const filteredStacks = stacks.filter(stack => {
    const matchesSearch = stack.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         stack.description?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesEnv = filterEnv === 'all' || stack.environment === filterEnv;
    const matchesStatus = filterStatus === 'all' || stack.status === filterStatus;
    return matchesSearch && matchesEnv && matchesStatus;
  });

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

  const getResourceIcon = (type) => {
    switch (type?.toUpperCase()) {
      case 'POSTGRES_CLUSTER':
        return <Database size={14} className="resource-type-icon postgres" />;
      case 'NGINX_GATEWAY':
        return <Globe size={14} className="resource-type-icon nginx" />;
      case 'DOCKER_SERVICE':
        return <Container size={14} className="resource-type-icon docker" />;
      default:
        return <Server size={14} className="resource-type-icon" />;
    }
  };

  const getEnvBadgeClass = (env) => {
    switch (env?.toLowerCase()) {
      case 'production':
        return 'env-badge production';
      case 'staging':
        return 'env-badge staging';
      default:
        return 'env-badge development';
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

  // If a stack is selected, show StackDetail
  if (selectedStack) {
    return (
      <StackDetail 
        stackId={selectedStack} 
        onBack={() => {
          setSelectedStack(null);
          fetchStacks();
        }}
        onLogout={onLogout}
      />
    );
  }

  return (
    <div className="dashboard">
      {/* Header */}
      <header className="dashboard-header">
        <div className="header-left">
          <Layers size={28} className="logo-icon" />
          <div>
            <h1>Infrastructure Stacks</h1>
            <p className="header-subtitle">Manage your infrastructure resources</p>
          </div>
        </div>
        <div className="header-right">
          <button className="btn btn-secondary" onClick={fetchStacks} disabled={loading}>
            <RefreshCw size={16} className={loading ? 'spin' : ''} />
          </button>
          <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
            <Plus size={16} />
            New Stack
          </button>
          <button className="btn btn-ghost" onClick={onLogout}>
            Logout
          </button>
        </div>
      </header>

      {/* Filters */}
      <div className="dashboard-filters">
        <div className="search-box">
          <Search size={18} />
          <input
            type="text"
            placeholder="Search stacks..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <div className="filter-group">
          <Filter size={16} />
          <select value={filterEnv} onChange={(e) => setFilterEnv(e.target.value)}>
            <option value="all">All Environments</option>
            <option value="development">Development</option>
            <option value="staging">Staging</option>
            <option value="production">Production</option>
          </select>
          <select value={filterStatus} onChange={(e) => setFilterStatus(e.target.value)}>
            <option value="all">All Status</option>
            <option value="running">Running</option>
            <option value="stopped">Stopped</option>
            <option value="creating">Creating</option>
            <option value="failed">Failed</option>
          </select>
        </div>
      </div>

      {/* Stack Table */}
      <div className="stacks-container">
        {loading ? (
          <div className="loading-state">
            <Loader size={40} className="spin" />
            <p>Loading stacks...</p>
          </div>
        ) : filteredStacks.length === 0 ? (
          <div className="empty-state">
            <Layers size={64} />
            <h3>{searchTerm || filterEnv !== 'all' || filterStatus !== 'all' ? 'No stacks found' : 'No stacks yet'}</h3>
            <p>
              {searchTerm || filterEnv !== 'all' || filterStatus !== 'all'
                ? 'Try adjusting your search or filters'
                : 'Create your first infrastructure stack to get started'}
            </p>
            {!searchTerm && filterEnv === 'all' && filterStatus === 'all' && (
              <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
                <Plus size={16} />
                Create Stack
              </button>
            )}
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
                  <tr key={stack.id} className="stack-row" onClick={() => setSelectedStack(stack.id)}>
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
                              {getResourceIcon(resource.resource_type)}
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
                          onClick={() => setSelectedStack(stack.id)}
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

      {/* Create Modal */}
      <CreateStackModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={() => {
          setShowCreateModal(false);
          fetchStacks();
        }}
      />
    </div>
  );
}

export default Dashboard;
