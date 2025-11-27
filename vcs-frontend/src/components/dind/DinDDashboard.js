import React, { useState, useEffect, useCallback } from 'react';
import { dinDAPI } from '../../api';
import DinDTerminal from './DinDTerminal';
import DinDImageBuilder from './DinDImageBuilder';
import DinDCompose from './DinDCompose';
import { 
  Plus, 
  Terminal, 
  Box, 
  Layers, 
  Play, 
  Square, 
  Trash2, 
  RefreshCw,
  Server,
  HardDrive,
  Cpu,
  Clock,
  Activity
} from 'lucide-react';
import './DinDDashboard.css';

const DinDDashboard = () => {
  const [environments, setEnvironments] = useState([]);
  const [selectedEnv, setSelectedEnv] = useState(null);
  const [activeTab, setActiveTab] = useState('terminal');
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [envStats, setEnvStats] = useState({});
  const [refreshing, setRefreshing] = useState(false);

  const loadEnvironments = useCallback(async () => {
    try {
      setLoading(true);
      const response = await dinDAPI.listEnvironments();
      if (response.data.success) {
        setEnvironments(response.data.data || []);
      }
    } catch (error) {
      console.error('Error loading environments:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  const loadEnvStats = useCallback(async (envId) => {
    try {
      const response = await dinDAPI.getStats(envId);
      if (response.data.success) {
        setEnvStats(prev => ({ ...prev, [envId]: response.data.data }));
      }
    } catch (error) {
      console.error('Error loading stats:', error);
    }
  }, []);

  useEffect(() => {
    loadEnvironments();
  }, [loadEnvironments]);

  useEffect(() => {
    if (selectedEnv) {
      loadEnvStats(selectedEnv.id);
      const interval = setInterval(() => loadEnvStats(selectedEnv.id), 10000);
      return () => clearInterval(interval);
    }
  }, [selectedEnv, loadEnvStats]);

  const handleCreateEnvironment = async (formData) => {
    try {
      const response = await dinDAPI.createEnvironment(formData);
      if (response.data.success) {
        setShowCreateModal(false);
        loadEnvironments();
        setSelectedEnv(response.data.data);
      }
    } catch (error) {
      alert('Error creating environment: ' + (error.response?.data?.message || error.message));
    }
  };

  const handleDeleteEnvironment = async (envId) => {
    if (!window.confirm('Bạn có chắc muốn xóa environment này? Tất cả containers và images bên trong sẽ bị mất.')) {
      return;
    }
    try {
      await dinDAPI.deleteEnvironment(envId);
      if (selectedEnv?.id === envId) {
        setSelectedEnv(null);
      }
      loadEnvironments();
    } catch (error) {
      alert('Error deleting environment: ' + error.message);
    }
  };

  const handleStartEnvironment = async (envId) => {
    try {
      await dinDAPI.startEnvironment(envId);
      loadEnvironments();
    } catch (error) {
      alert('Error starting environment: ' + error.message);
    }
  };

  const handleStopEnvironment = async (envId) => {
    try {
      await dinDAPI.stopEnvironment(envId);
      loadEnvironments();
    } catch (error) {
      alert('Error stopping environment: ' + error.message);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    await loadEnvironments();
    if (selectedEnv) {
      await loadEnvStats(selectedEnv.id);
    }
    setRefreshing(false);
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'running': return '#10b981';
      case 'stopped': return '#ef4444';
      case 'creating': return '#f59e0b';
      default: return '#6b7280';
    }
  };

  const stats = selectedEnv ? envStats[selectedEnv.id] : null;

  return (
    <div className="dind-dashboard">
      {/* Left Sidebar - Environment List */}
      <aside className="dind-sidebar">
        <div className="sidebar-header">
          <h2>
            <Box size={20} />
            Docker Sandboxes
          </h2>
          <div className="sidebar-actions">
            <button 
              className="btn-icon" 
              onClick={handleRefresh}
              title="Refresh"
            >
              <RefreshCw size={16} className={refreshing ? 'spinning' : ''} />
            </button>
            <button 
              className="btn-primary-sm" 
              onClick={() => setShowCreateModal(true)}
            >
              <Plus size={16} />
              New
            </button>
          </div>
        </div>

        <div className="env-list">
          {loading ? (
            <div className="loading-state">
              <RefreshCw className="spinning" size={24} />
              <span>Loading environments...</span>
            </div>
          ) : environments.length === 0 ? (
            <div className="empty-state">
              <Box size={48} strokeWidth={1} />
              <p>No Docker Sandboxes yet</p>
              <button 
                className="btn-primary" 
                onClick={() => setShowCreateModal(true)}
              >
                <Plus size={16} />
                Create your first sandbox
              </button>
            </div>
          ) : (
            environments.map(env => (
              <div 
                key={env.id}
                className={`env-card ${selectedEnv?.id === env.id ? 'selected' : ''}`}
                onClick={() => setSelectedEnv(env)}
              >
                <div className="env-header">
                  <span className="env-name">{env.name}</span>
                  <span 
                    className="env-status"
                    style={{ backgroundColor: getStatusColor(env.status) }}
                  >
                    {env.status}
                  </span>
                </div>
                <div className="env-info">
                  <span><Cpu size={12} /> {env.cpu_limit} CPU</span>
                  <span><HardDrive size={12} /> {env.memory_limit}</span>
                </div>
                <div className="env-actions">
                  {env.status === 'running' ? (
                    <button 
                      className="btn-icon-sm warning"
                      onClick={(e) => { e.stopPropagation(); handleStopEnvironment(env.id); }}
                      title="Stop"
                    >
                      <Square size={14} />
                    </button>
                  ) : (
                    <button 
                      className="btn-icon-sm success"
                      onClick={(e) => { e.stopPropagation(); handleStartEnvironment(env.id); }}
                      title="Start"
                    >
                      <Play size={14} />
                    </button>
                  )}
                  <button 
                    className="btn-icon-sm danger"
                    onClick={(e) => { e.stopPropagation(); handleDeleteEnvironment(env.id); }}
                    title="Delete"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="dind-main">
        {selectedEnv ? (
          <>
            {/* Environment Header */}
            <header className="env-detail-header">
              <div className="env-title">
                <h1>{selectedEnv.name}</h1>
                <span 
                  className="status-badge"
                  style={{ backgroundColor: getStatusColor(selectedEnv.status) }}
                >
                  {selectedEnv.status}
                </span>
              </div>
              
              {/* Quick Stats */}
              <div className="quick-stats">
                <div className="stat-item">
                  <Server size={16} />
                  <span>{stats?.container_count || 0}</span>
                  <label>Containers</label>
                </div>
                <div className="stat-item">
                  <Layers size={16} />
                  <span>{stats?.image_count || 0}</span>
                  <label>Images</label>
                </div>
                <div className="stat-item">
                  <Activity size={16} />
                  <span>{selectedEnv.resource_plan}</span>
                  <label>Plan</label>
                </div>
                <div className="stat-item">
                  <Clock size={16} />
                  <span>{new Date(selectedEnv.created_at).toLocaleDateString()}</span>
                  <label>Created</label>
                </div>
              </div>
            </header>

            {/* Tab Navigation */}
            <nav className="tab-nav">
              <button 
                className={`tab-btn ${activeTab === 'terminal' ? 'active' : ''}`}
                onClick={() => setActiveTab('terminal')}
              >
                <Terminal size={16} />
                Terminal
              </button>
              <button 
                className={`tab-btn ${activeTab === 'build' ? 'active' : ''}`}
                onClick={() => setActiveTab('build')}
              >
                <Box size={16} />
                Image Builder
              </button>
              <button 
                className={`tab-btn ${activeTab === 'compose' ? 'active' : ''}`}
                onClick={() => setActiveTab('compose')}
              >
                <Layers size={16} />
                Docker Compose
              </button>
            </nav>

            {/* Tab Content */}
            <div className="tab-content">
              {activeTab === 'terminal' && (
                <DinDTerminal 
                  environmentId={selectedEnv.id} 
                  environmentStatus={selectedEnv.status}
                  onRefresh={() => loadEnvStats(selectedEnv.id)}
                />
              )}
              {activeTab === 'build' && (
                <DinDImageBuilder 
                  environmentId={selectedEnv.id}
                  environmentStatus={selectedEnv.status}
                  onRefresh={() => loadEnvStats(selectedEnv.id)}
                />
              )}
              {activeTab === 'compose' && (
                <DinDCompose 
                  environmentId={selectedEnv.id}
                  environmentStatus={selectedEnv.status}
                  onRefresh={() => loadEnvStats(selectedEnv.id)}
                />
              )}
            </div>
          </>
        ) : (
          <div className="no-selection">
            <Box size={64} strokeWidth={1} />
            <h2>Select or create a Docker Sandbox</h2>
            <p>Docker Sandbox provides an isolated Docker environment where you can run any docker commands, build images, and test applications.</p>
            <button 
              className="btn-primary-lg"
              onClick={() => setShowCreateModal(true)}
            >
              <Plus size={20} />
              Create New Sandbox
            </button>
          </div>
        )}
      </main>

      {/* Create Environment Modal */}
      {showCreateModal && (
        <CreateEnvironmentModal 
          onClose={() => setShowCreateModal(false)}
          onCreate={handleCreateEnvironment}
        />
      )}
    </div>
  );
};

// Create Environment Modal Component
const CreateEnvironmentModal = ({ onClose, onCreate }) => {
  const [formData, setFormData] = useState({
    name: '',
    resource_plan: 'medium',
    description: '',
    auto_cleanup: false,
    ttl_hours: 24
  });
  const [creating, setCreating] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setCreating(true);
    try {
      await onCreate(formData);
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal create-env-modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2><Box size={24} /> Create Docker Sandbox</h2>
        </div>
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Sandbox Name</label>
            <input
              type="text"
              placeholder="my-docker-sandbox"
              value={formData.name}
              onChange={e => setFormData({ ...formData, name: e.target.value })}
              required
            />
          </div>

          <div className="form-group">
            <label>Resource Plan</label>
            <div className="plan-options">
              {['small', 'medium', 'large'].map(plan => (
                <div 
                  key={plan}
                  className={`plan-card ${formData.resource_plan === plan ? 'selected' : ''}`}
                  onClick={() => setFormData({ ...formData, resource_plan: plan })}
                >
                  <h4>{plan.charAt(0).toUpperCase() + plan.slice(1)}</h4>
                  <p>
                    {plan === 'small' && '1 CPU / 1GB RAM'}
                    {plan === 'medium' && '2 CPU / 2GB RAM'}
                    {plan === 'large' && '4 CPU / 4GB RAM'}
                  </p>
                </div>
              ))}
            </div>
          </div>

          <div className="form-group">
            <label>Description (Optional)</label>
            <textarea
              placeholder="What will you use this sandbox for?"
              value={formData.description}
              onChange={e => setFormData({ ...formData, description: e.target.value })}
              rows={3}
            />
          </div>

          <div className="form-group checkbox-group">
            <label>
              <input
                type="checkbox"
                checked={formData.auto_cleanup}
                onChange={e => setFormData({ ...formData, auto_cleanup: e.target.checked })}
              />
              Auto cleanup after TTL
            </label>
            {formData.auto_cleanup && (
              <div className="ttl-input">
                <input
                  type="number"
                  min="1"
                  max="720"
                  value={formData.ttl_hours}
                  onChange={e => setFormData({ ...formData, ttl_hours: parseInt(e.target.value) })}
                />
                <span>hours</span>
              </div>
            )}
          </div>

          <div className="modal-actions">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-primary" disabled={creating || !formData.name}>
              {creating ? (
                <>
                  <RefreshCw className="spinning" size={16} />
                  Creating...
                </>
              ) : (
                <>
                  <Plus size={16} />
                  Create Sandbox
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default DinDDashboard;

