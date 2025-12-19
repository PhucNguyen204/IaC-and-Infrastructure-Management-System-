import React, { useState } from 'react';
import { X, Database, Globe, Box, Server, Shield, Network, Cpu, HardDrive, Clock, Activity } from 'lucide-react';
import { stackAPI } from '../api';
import toast from 'react-hot-toast';
import './CreateStackModal.css';

const CreateStackModal = ({ isOpen, onClose, onSuccess }) => {
  const [isCreating, setIsCreating] = useState(false);

  // Single resource state instead of stack form
  const [resource, setResource] = useState({
    type: 'POSTGRES_CLUSTER',
    name: '',
    role: 'database',
    spec: {
      // Default Postgres Spec
      postgres_version: '17',
      node_count: 2,
      replication_mode: 'async',
      postgres_password: ''
    }
  });

  if (!isOpen) return null;

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

  const handleTypeChange = (newType) => {
    setResource(prev => ({
      ...prev,
      type: newType,
      spec: getDefaultSpec(newType)
    }));
  };

  const handleConfigChange = (key, value) => {
    setResource(prev => ({
      ...prev,
      spec: { ...prev.spec, [key]: value }
    }));
  };

  const validateResource = () => {
    if (!resource.name) {
      toast.error('Please enter service name');
      return false;
    }
    if (!/^[a-z0-9-]+$/.test(resource.name)) {
      toast.error('Name must contain only lowercase letters, numbers, and hyphens');
      return false;
    }

    if (resource.type === 'POSTGRES_CLUSTER') {
      if (!resource.spec.postgres_password || resource.spec.postgres_password.length < 8) {
        toast.error('Postgres password is required (min 8 characters)');
        return false;
      }
    }
    return true;
  };

  const handleSubmit = async () => {
    if (!validateResource()) return;

    try {
      setIsCreating(true);
      const loadingToast = toast.loading('Provisioning infrastructure...');

      // Construct backend-compatible Stack payload
      // Hiding the stack concept by using resource name as stack name
      const payload = {
        name: resource.name,
        environment: 'production', // Defaulting to production for simplicity
        description: `Auto-generated stack for ${resource.name}`,
        resources: [{
          resource_type: resource.type,
          resource_name: resource.name,
          role: resource.role,
          spec: resource.spec,
          order: 1
        }]
      };

      console.log('Creating infrastructure:', payload);
      await stackAPI.create(payload);

      toast.dismiss(loadingToast);
      toast.success('Infrastructure provisioning started!');

      onSuccess();
      handleClose();
    } catch (error) {
      console.error('Error creating infrastructure:', error);
      toast.error(error.response?.data?.error || 'Failed to create infrastructure');
    } finally {
      setIsCreating(false);
    }
  };

  const handleClose = () => {
    setResource({
      type: 'POSTGRES_CLUSTER',
      name: '',
      role: 'database',
      spec: getDefaultSpec('POSTGRES_CLUSTER')
    });
    onClose();
  };

  // --- Rendering Helpers ---

  const renderPostgresConfig = () => (
    <div className="config-container">
      <div className="form-group">
        <label><Database size={14} /> PostgreSQL Version</label>
        <select
          value={resource.spec.postgres_version || '17'}
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
          value={resource.spec.postgres_password || ''}
          onChange={(e) => handleConfigChange('postgres_password', e.target.value)}
          placeholder="Min 8 characters"
        />
      </div>
      <div className="form-group">
        <label><Server size={14} /> Node Count</label>
        <input
          type="number"
          min="1"
          max="5"
          value={resource.spec.node_count || 2}
          onChange={(e) => handleConfigChange('node_count', parseInt(e.target.value))}
        />
      </div>
    </div>
  );

  const renderNginxConfig = () => (
    <div className="config-container">
      <div className="form-group full-width">
        <label>Nginx Configuration</label>
        <textarea
          value={resource.spec.config || ''}
          onChange={(e) => handleConfigChange('config', e.target.value)}
          rows="6"
          className="code-textarea"
        />
      </div>
    </div>
  );

  const renderDinDConfig = () => (
    <div className="config-container">
      <div className="form-group">
        <label><Cpu size={14} /> Resource Plan</label>
        <select
          value={resource.spec.resource_plan || 'medium'}
          onChange={(e) => handleConfigChange('resource_plan', e.target.value)}
        >
          <option value="small">Small (1 CPU / 1GB)</option>
          <option value="medium">Medium (2 CPU / 2GB)</option>
          <option value="large">Large (4 CPU / 4GB)</option>
        </select>
      </div>
    </div>
  );

  const renderConfig = () => {
    switch (resource.type) {
      case 'POSTGRES_CLUSTER': return renderPostgresConfig();
      case 'NGINX_GATEWAY': return renderNginxConfig();
      case 'DIND_ENVIRONMENT': return renderDinDConfig();
      default: return <div className="text-muted">No specific configuration</div>;
    }
  };

  return (
    <div className="modal-overlay" onClick={handleClose}>
      <div className="modal-content create-infrastructure-modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>New Infrastructure</h2>
          <button className="close-btn" onClick={handleClose}><X size={20} /></button>
        </div>

        <div className="modal-body">
          <div className="form-section">
            <div className="config-grid">
              <div className="form-group">
                <label>Service Type</label>
                <select
                  value={resource.type}
                  onChange={(e) => handleTypeChange(e.target.value)}
                >
                  <option value="POSTGRES_CLUSTER">PostgreSQL Cluster</option>
                  <option value="NGINX_GATEWAY">Nginx Gateway</option>
                  <option value="DIND_ENVIRONMENT">Docker Sandbox</option>
                </select>
              </div>
              <div className="form-group">
                <label>Service Name *</label>
                <input
                  type="text"
                  value={resource.name}
                  onChange={(e) => setResource(prev => ({ ...prev, name: e.target.value }))}
                  placeholder="e.g. production-db"
                />
              </div>
            </div>

            <div className="config-section-divider">
              <span>Configuration</span>
            </div>

            {renderConfig()}
          </div>
        </div>

        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={handleClose}>Cancel</button>
          <button className="btn btn-primary" onClick={handleSubmit} disabled={isCreating}>
            {isCreating ? 'Provisioning...' : 'Create Service'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default CreateStackModal;
