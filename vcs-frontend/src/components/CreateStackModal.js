import React, { useState } from 'react';
import { X, Plus, Trash2, Database, Globe, Box, Server, Shield, Network, Cpu, HardDrive, Clock } from 'lucide-react';
import { stackAPI } from '../api';
import toast from 'react-hot-toast';
import './CreateStackModal.css';

const CreateStackModal = ({ isOpen, onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    name: '',
    environment: 'development',
    description: '',
    resources: []
  });

  // Default spec for POSTGRES_CLUSTER
  const getInitialPostgresSpec = () => ({
    postgres_version: '17',
    node_count: 2,
    replication_mode: 'async',
    postgres_password: ''
  });

  const [newResource, setNewResource] = useState({
    type: 'POSTGRES_CLUSTER',
    name: '',
    role: 'database',
    spec: getInitialPostgresSpec()
  });

  const [isCreating, setIsCreating] = useState(false);

  if (!isOpen) return null;

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleResourceChange = (e) => {
    const { name, value } = e.target;
    setNewResource(prev => ({
      ...prev,
      [name]: value,
      // Reset spec when type changes
      ...(name === 'type' ? { spec: getDefaultSpec(value) } : {})
    }));
  };

  const getDefaultSpec = (type) => {
    switch (type) {
      case 'POSTGRES_CLUSTER':
        return {
          postgres_version: '17',
          node_count: 2,
          replication_mode: 'async',
          postgres_password: ''
        };
      case 'NGINX_GATEWAY':
        return {
          port: 8080,
          config: `server {
    listen 80;
    server_name localhost;
    
    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}`
        };
      case 'DIND_ENVIRONMENT':
        return {
          resource_plan: 'medium',
          description: '',
          auto_cleanup: false,
          ttl_hours: 24
        };
      case 'NGINX_CLUSTER':
        return {
          node_count: 2,
          http_port: 8080,
          https_port: 8443,
          load_balance_mode: 'round_robin',
          virtual_ip: '192.168.0.100',
          worker_connections: 2048,
          worker_processes: 2,
          ssl_enabled: false,
          gzip_enabled: true,
          health_check_enabled: true,
          health_check_path: '/health',
          rate_limit_enabled: false,
          cache_enabled: false
        };
      default:
        return {};
    }
  };

  const handleConfigChange = (key, value) => {
    setNewResource(prev => ({
      ...prev,
      spec: {
        ...prev.spec,
        [key]: value
      }
    }));
  };

  const handleArrayConfigChange = (arrayKey, index, key, value) => {
    setNewResource(prev => {
      const newArray = [...(prev.spec[arrayKey] || [])];
      newArray[index] = { ...newArray[index], [key]: value };
      return {
        ...prev,
        spec: {
          ...prev.spec,
          [arrayKey]: newArray
        }
      };
    });
  };

  const addArrayItem = (arrayKey, defaultItem) => {
    setNewResource(prev => ({
      ...prev,
      spec: {
        ...prev.spec,
        [arrayKey]: [...(prev.spec[arrayKey] || []), defaultItem]
      }
    }));
  };

  const removeArrayItem = (arrayKey, index) => {
    setNewResource(prev => ({
      ...prev,
      spec: {
        ...prev.spec,
        [arrayKey]: prev.spec[arrayKey].filter((_, i) => i !== index)
      }
    }));
  };

  const validateResource = () => {
    if (!newResource.name) {
      toast.error('Please enter resource name');
      return false;
    }

    if (newResource.type === 'POSTGRES_CLUSTER') {
      if (!newResource.spec.postgres_password || newResource.spec.postgres_password.length < 8) {
        toast.error('Postgres password is required (min 8 characters)');
        return false;
      }
    }

    if (newResource.type === 'NGINX_GATEWAY') {
      if (!newResource.spec.config) {
        toast.error('Nginx config is required');
        return false;
      }
    }

    if (newResource.type === 'DIND_ENVIRONMENT') {
      if (!newResource.spec.resource_plan) {
        toast.error('Resource plan is required');
        return false;
      }
    }

    if (newResource.type === 'NGINX_CLUSTER') {
      if (!newResource.spec.http_port) {
        toast.error('HTTP port is required');
        return false;
      }
      if ((newResource.spec.node_count || 0) < 2) {
        toast.error('Nginx cluster requires at least 2 nodes');
        return false;
      }
    }

    return true;
  };

  const addResource = () => {
    if (!validateResource()) return;

    const resourceWithOrder = {
      resource_type: newResource.type,
      resource_name: newResource.name,
      role: newResource.role,
      spec: newResource.spec,
      order: formData.resources.length + 1
    };

    setFormData(prev => ({
      ...prev,
      resources: [...prev.resources, resourceWithOrder]
    }));

    // Reset to new resource with defaults
    setNewResource({
      type: 'POSTGRES_CLUSTER',
      name: '',
      role: 'database',
      spec: getInitialPostgresSpec()
    });

    toast.success('Resource added to stack');
  };

  const removeResource = (index) => {
    setFormData(prev => ({
      ...prev,
      resources: prev.resources.filter((_, i) => i !== index)
    }));
    toast.success('Resource removed');
  };

  const handleSubmit = async () => {
    if (!formData.name) {
      toast.error('Please enter stack name');
      return;
    }

    if (formData.resources.length === 0) {
      toast.error('Please add at least one resource');
      return;
    }

    try {
      setIsCreating(true);
      const loadingToast = toast.loading('Creating stack and provisioning infrastructure...');

      console.log('Creating stack with data:', JSON.stringify(formData, null, 2));
      const response = await stackAPI.create(formData);
      console.log('Stack created:', response.data);

      toast.dismiss(loadingToast);
      toast.success(
        `Stack "${formData.name}" is being created! Infrastructure provisioning in progress...`,
        { duration: 5000 }
      );

      onSuccess();
      handleClose();
    } catch (error) {
      console.error('Error creating stack:', error);
      console.error('Error response:', error.response?.data);
      toast.error(error.response?.data?.error || 'Failed to create stack');
    } finally {
      setIsCreating(false);
    }
  };

  const handleClose = () => {
    setFormData({
      name: '',
      environment: 'development',
      description: '',
      resources: []
    });
    setNewResource({
      type: 'POSTGRES_CLUSTER',
      name: '',
      role: 'database',
      spec: getInitialPostgresSpec()
    });
    onClose();
  };

  const getResourceIcon = (type) => {
    switch (type) {
      case 'POSTGRES_CLUSTER':
        return <Database size={18} />;
      case 'NGINX_GATEWAY':
      case 'NGINX_CLUSTER':
        return <Globe size={18} />;
      case 'DIND_ENVIRONMENT':
        return <Box size={18} />;
      default:
        return <Server size={18} />;
    }
  };

  const getResourceColor = (type) => {
    switch (type) {
      case 'POSTGRES_CLUSTER':
        return '#336791';
      case 'NGINX_GATEWAY':
      case 'NGINX_CLUSTER':
        return '#009639';
      case 'DIND_ENVIRONMENT':
        return '#6366f1';
      default:
        return '#6b7280';
    }
  };

  const renderPostgresConfig = () => {
    const { spec } = newResource;
    return (
      <div className="config-container">
        <div className="config-section">
          <div className="config-section-header">
            <div className="config-section-title">
              <Database size={16} />
              <span>PostgreSQL Configuration</span>
            </div>
          </div>
          <div className="config-section-content">
            <div className="config-grid">
              <div className="form-group">
                <label><Database size={14} /> PostgreSQL Version</label>
                <select
                  value={spec.postgres_version || '17'}
                  onChange={(e) => handleConfigChange('postgres_version', e.target.value)}
                >
                  <option value="15">PostgreSQL 15</option>
                  <option value="16">PostgreSQL 16</option>
                  <option value="17">PostgreSQL 17</option>
                </select>
              </div>
              <div className="form-group">
                <label><Shield size={14} /> Password *</label>
                <input
                  type="password"
                  value={spec.postgres_password || ''}
                  onChange={(e) => handleConfigChange('postgres_password', e.target.value)}
                  placeholder="Min 8 characters"
                />
              </div>
              <div className="form-group">
                <label><Server size={14} /> Node Count</label>
                <input
                  type="number"
                  min="1"
                  max="10"
                  value={spec.node_count || 2}
                  onChange={(e) => handleConfigChange('node_count', parseInt(e.target.value))}
                />
              </div>
              <div className="form-group">
                <label><Network size={14} /> Replication Mode</label>
                <select
                  value={spec.replication_mode || 'async'}
                  onChange={(e) => handleConfigChange('replication_mode', e.target.value)}
                >
                  <option value="async">Asynchronous</option>
                  <option value="sync">Synchronous</option>
                </select>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  };

  const renderNginxConfig = () => {
    const { spec } = newResource;
    return (
      <div className="config-container">
        <div className="config-section">
          <div className="config-section-header">
            <div className="config-section-title">
              <Globe size={16} />
              <span>Nginx Configuration</span>
            </div>
          </div>
          <div className="config-section-content">
            <div className="config-grid">
              <div className="form-group">
                <label>HTTP Port *</label>
                <input
                  type="number"
                  min="1"
                  max="65535"
                  value={spec.port || 8080}
                  onChange={(e) => handleConfigChange('port', parseInt(e.target.value))}
                />
              </div>
            </div>
            <div className="form-group full-width">
              <label>Nginx Configuration *</label>
              <textarea
                value={spec.config || ''}
                onChange={(e) => handleConfigChange('config', e.target.value)}
                placeholder="Enter nginx configuration..."
                rows="8"
                className="code-textarea"
              />
            </div>
          </div>
        </div>
      </div>
    );
  };

  const renderNginxClusterConfig = () => {
    const { spec } = newResource;
    return (
      <div className="config-container">
        <div className="config-section">
          <div className="config-section-header">
            <div className="config-section-title">
              <Globe size={16} />
              <span>Nginx Cluster Configuration</span>
            </div>
          </div>
          <div className="config-section-content">
            <div className="config-grid">
              <div className="form-group">
                <label>Node Count *</label>
                <input
                  type="number"
                  min="2"
                  max="10"
                  value={spec.node_count || 2}
                  onChange={(e) => handleConfigChange('node_count', parseInt(e.target.value, 10))}
                />
              </div>
              <div className="form-group">
                <label>HTTP Port *</label>
                <input
                  type="number"
                  min="1"
                  max="65535"
                  value={spec.http_port || 8080}
                  onChange={(e) => handleConfigChange('http_port', parseInt(e.target.value, 10))}
                />
              </div>
              <div className="form-group">
                <label>HTTPS Port</label>
                <input
                  type="number"
                  min="1"
                  max="65535"
                  value={spec.https_port || ''}
                  onChange={(e) => handleConfigChange('https_port', parseInt(e.target.value || 0, 10))}
                  placeholder="Optional"
                />
              </div>
              <div className="form-group">
                <label>Load Balancing Mode</label>
                <select
                  value={spec.load_balance_mode || 'round_robin'}
                  onChange={(e) => handleConfigChange('load_balance_mode', e.target.value)}
                >
                  <option value="round_robin">Round Robin</option>
                  <option value="least_conn">Least Connections</option>
                  <option value="ip_hash">IP Hash</option>
                </select>
              </div>
              <div className="form-group">
                <label>Virtual IP</label>
                <input
                  type="text"
                  value={spec.virtual_ip || ''}
                  onChange={(e) => handleConfigChange('virtual_ip', e.target.value)}
                  placeholder="e.g., 192.168.0.100"
                />
              </div>
              <div className="form-group">
                <label>Worker Connections</label>
                <input
                  type="number"
                  min="256"
                  max="100000"
                  value={spec.worker_connections || 2048}
                  onChange={(e) => handleConfigChange('worker_connections', parseInt(e.target.value, 10))}
                />
              </div>
            </div>
            <div className="config-grid">
              <div className="form-group checkbox">
                <label>
                  <input
                    type="checkbox"
                    checked={spec.ssl_enabled || false}
                    onChange={(e) => handleConfigChange('ssl_enabled', e.target.checked)}
                  />
                  Enable SSL (provide certs later)
                </label>
              </div>
              <div className="form-group checkbox">
                <label>
                  <input
                    type="checkbox"
                    checked={spec.gzip_enabled || false}
                    onChange={(e) => handleConfigChange('gzip_enabled', e.target.checked)}
                  />
                  Enable Gzip compression
                </label>
              </div>
              <div className="form-group checkbox">
                <label>
                  <input
                    type="checkbox"
                    checked={spec.rate_limit_enabled || false}
                    onChange={(e) => handleConfigChange('rate_limit_enabled', e.target.checked)}
                  />
                  Enable Rate Limiting
                </label>
              </div>
              <div className="form-group checkbox">
                <label>
                  <input
                    type="checkbox"
                    checked={spec.cache_enabled || false}
                    onChange={(e) => handleConfigChange('cache_enabled', e.target.checked)}
                  />
                  Enable Disk Cache
                </label>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  };

  const renderDinDConfig = () => {
    const { spec } = newResource;
    
    const planDetails = {
      small: { cpu: '1 CPU', memory: '1 GB RAM', desc: 'For light development & testing' },
      medium: { cpu: '2 CPU', memory: '2 GB RAM', desc: 'Balanced for most workloads' },
      large: { cpu: '4 CPU', memory: '4 GB RAM', desc: 'For heavy builds & multiple containers' }
    };

    return (
      <div className="config-container">
        <div className="config-section">
          <div className="config-section-header">
            <div className="config-section-title">
              <Box size={16} />
              <span>Docker Sandbox Configuration</span>
            </div>
          </div>
          <div className="config-section-content">
            {/* Resource Plan Selection */}
            <div className="form-group full-width">
              <label><Cpu size={14} /> Resource Plan *</label>
              <div className="plan-grid">
                {['small', 'medium', 'large'].map(plan => (
                  <div 
                    key={plan}
                    className={`plan-card ${spec.resource_plan === plan ? 'selected' : ''}`}
                    onClick={() => handleConfigChange('resource_plan', plan)}
                  >
                    <div className="plan-header">
                      <span className="plan-name">{plan.charAt(0).toUpperCase() + plan.slice(1)}</span>
                      {spec.resource_plan === plan && <span className="plan-check">✓</span>}
                    </div>
                    <div className="plan-specs">
                      <span><Cpu size={12} /> {planDetails[plan].cpu}</span>
                      <span><HardDrive size={12} /> {planDetails[plan].memory}</span>
                    </div>
                    <div className="plan-desc">{planDetails[plan].desc}</div>
                  </div>
                ))}
              </div>
            </div>

            <div className="form-group full-width">
              <label>Description</label>
              <textarea
                value={spec.description || ''}
                onChange={(e) => handleConfigChange('description', e.target.value)}
                placeholder="What will you use this Docker Sandbox for?"
                rows="2"
              />
            </div>

            <div className="config-grid">
              <div className="form-group checkbox">
                <label>
                  <input
                    type="checkbox"
                    checked={spec.auto_cleanup || false}
                    onChange={(e) => handleConfigChange('auto_cleanup', e.target.checked)}
                  />
                  <Clock size={14} /> Auto cleanup after TTL
                </label>
              </div>
              {spec.auto_cleanup && (
                <div className="form-group">
                  <label>TTL (hours)</label>
                  <input
                    type="number"
                    min="1"
                    max="720"
                    value={spec.ttl_hours || 24}
                    onChange={(e) => handleConfigChange('ttl_hours', parseInt(e.target.value))}
                  />
                </div>
              )}
            </div>

          </div>
        </div>
      </div>
    );
  };

  const renderConfigFields = () => {
    switch (newResource.type) {
      case 'POSTGRES_CLUSTER':
        return renderPostgresConfig();
      case 'NGINX_GATEWAY':
        return renderNginxConfig();
      case 'NGINX_CLUSTER':
        return renderNginxClusterConfig();
      case 'DIND_ENVIRONMENT':
        return renderDinDConfig();
      default:
        return null;
    }
  };

  const getResourceSummary = (resource) => {
    switch (resource.resource_type) {
      case 'POSTGRES_CLUSTER':
        return `${resource.spec.node_count || 3} nodes • v${resource.spec.postgres_version || '17'} • ${resource.spec.replication_mode || 'async'}`;
      case 'NGINX_GATEWAY':
        return `Port ${resource.spec.port || 8080}${resource.spec.ssl_port ? ` / SSL ${resource.spec.ssl_port}` : ''}`;
      case 'NGINX_CLUSTER':
        return `${resource.spec.node_count || 2} nodes • HTTP ${resource.spec.http_port || 8080}${resource.spec.https_port ? ` / HTTPS ${resource.spec.https_port}` : ''}`;
      case 'DIND_ENVIRONMENT':
        return `${(resource.spec.resource_plan || 'medium').toUpperCase()} plan${resource.spec.auto_cleanup ? ` • TTL ${resource.spec.ttl_hours}h` : ''}`;
      default:
        return '';
    }
  };

  return (
    <div className="modal-overlay" onClick={handleClose}>
      <div className="modal-content create-stack-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Create New Stack</h2>
          <button className="close-btn" onClick={handleClose}>
            <X size={20} />
          </button>
        </div>

        <div className="modal-body">
          {/* Basic Info */}
          <div className="form-section">
            <h3>Stack Information</h3>
            <div className="config-grid">
              <div className="form-group">
                <label>Stack Name *</label>
                <input
                  type="text"
                  name="name"
                  value={formData.name}
                  onChange={handleInputChange}
                  placeholder="e.g., my-production-app"
                />
              </div>
              <div className="form-group">
                <label>Environment</label>
                <select
                  name="environment"
                  value={formData.environment}
                  onChange={handleInputChange}
                >
                  <option value="development">Development</option>
                  <option value="staging">Staging</option>
                  <option value="production">Production</option>
                </select>
              </div>
            </div>
            <div className="form-group">
              <label>Description</label>
              <textarea
                name="description"
                value={formData.description}
                onChange={handleInputChange}
                placeholder="Describe your stack..."
                rows="2"
              />
            </div>
          </div>

          {/* Added Resources */}
          {formData.resources.length > 0 && (
            <div className="form-section">
              <h3>Resources ({formData.resources.length})</h3>
              <div className="resources-list">
                {formData.resources.map((resource, index) => (
                  <div key={index} className="resource-item">
                    <div
                      className="resource-icon"
                      style={{ background: getResourceColor(resource.resource_type) }}
                    >
                      {getResourceIcon(resource.resource_type)}
                    </div>
                    <div className="resource-info">
                      <div className="resource-name">{resource.resource_name}</div>
                      <div className="resource-type">{getResourceSummary(resource)}</div>
                    </div>
                    <span className="resource-badge">{resource.resource_type.replace(/_/g, ' ')}</span>
                    <button
                      type="button"
                      className="btn-icon btn-danger"
                      onClick={() => removeResource(index)}
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Add Resource Form */}
          <div className="form-section">
            <h3>Add Resource</h3>
            <div className="resource-form">
              <div className="config-grid">
                <div className="form-group">
                  <label>Resource Type</label>
                  <select
                    name="type"
                    value={newResource.type}
                    onChange={handleResourceChange}
                  >
                    <option value="POSTGRES_CLUSTER">PostgreSQL Cluster</option>
                    <option value="NGINX_GATEWAY">Nginx Gateway</option>
                    <option value="NGINX_CLUSTER">Nginx HA Cluster</option>
                    <option value="DIND_ENVIRONMENT">Docker Sandbox (DinD)</option>
                  </select>
                </div>
                <div className="form-group">
                  <label>Resource Name *</label>
                  <input
                    type="text"
                    name="name"
                    value={newResource.name}
                    onChange={handleResourceChange}
                    placeholder="e.g., main-database"
                  />
                </div>
                <div className="form-group">
                  <label>Role</label>
                  <select
                    name="role"
                    value={newResource.role}
                    onChange={handleResourceChange}
                  >
                    <option value="database">Database</option>
                    <option value="gateway">Gateway</option>
                    <option value="app">Application</option>
                    <option value="cache">Cache</option>
                    <option value="queue">Message Queue</option>
                    <option value="sandbox">Docker Sandbox</option>
                  </select>
                </div>
              </div>

              {renderConfigFields()}

              <div className="form-actions">
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={addResource}
                >
                  <Plus size={16} />
                  Add Resource to Stack
                </button>
              </div>
            </div>
          </div>
        </div>

        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={handleClose} disabled={isCreating}>
            Cancel
          </button>
          <button className="btn btn-primary" onClick={handleSubmit} disabled={isCreating || formData.resources.length === 0}>
            {isCreating ? (
              <>
                <span className="spinner-small"></span>
                Creating Stack...
              </>
            ) : (
              <>Create Stack ({formData.resources.length} resources)</>
            )}
          </button>
        </div>
      </div>
    </div>
  );
};

export default CreateStackModal;
