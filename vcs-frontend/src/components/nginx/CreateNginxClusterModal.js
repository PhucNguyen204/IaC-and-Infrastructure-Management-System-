import React, { useState } from 'react';
import { X, Plus, AlertCircle } from 'lucide-react';
import { nginxClusterAPI } from '../../api';
import toast from 'react-hot-toast';
import './CreateNginxClusterModal.css';

const CreateNginxClusterModal = ({ isOpen, onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    cluster_name: '',
    node_count: 2,
    http_port: 8080,
    https_port: 8443,
    load_balance_mode: 'round_robin',
    virtual_ip: '',
    worker_connections: 2048,
    worker_processes: 2,
    ssl_enabled: false,
    gzip_enabled: true,
    health_check_enabled: true,
    health_check_path: '/health',
    rate_limit_enabled: false,
    cache_enabled: false,
  });

  const [isCreating, setIsCreating] = useState(false);

  if (!isOpen) return null;

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.cluster_name.trim()) {
      toast.error('Cluster name is required');
      return;
    }

    if (formData.node_count < 2) {
      toast.error('Minimum 2 nodes required for HA cluster');
      return;
    }

    setIsCreating(true);

    try {
      const response = await nginxClusterAPI.create(formData);
      
      if (response.data.success) {
        toast.success('Nginx Cluster created successfully!');
        onSuccess && onSuccess(response.data.data);
        onClose();
      } else {
        toast.error(response.data.message || 'Failed to create cluster');
      }
    } catch (error) {
      console.error('Error creating nginx cluster:', error);
      toast.error(error.response?.data?.message || 'Failed to create cluster');
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal-container nginx-cluster-modal">
        <div className="modal-header">
          <h2>Create Nginx HA Cluster</h2>
          <button className="close-btn" onClick={onClose}>
            <X size={24} />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="modal-body">
            <div className="info-banner">
              <AlertCircle size={16} />
              <span>High Availability Nginx Cluster with Keepalived + VRRP</span>
            </div>

            {/* Basic Configuration */}
            <div className="config-section">
              <h3>Basic Configuration</h3>
              
              <div className="form-row">
                <div className="form-group">
                  <label>Cluster Name *</label>
                  <input
                    type="text"
                    name="cluster_name"
                    value={formData.cluster_name}
                    onChange={handleInputChange}
                    placeholder="my-nginx-cluster"
                    required
                  />
                </div>

                <div className="form-group">
                  <label>Node Count *</label>
                  <input
                    type="number"
                    name="node_count"
                    value={formData.node_count}
                    onChange={handleInputChange}
                    min="2"
                    max="10"
                    required
                  />
                  <small>Minimum 2 nodes for HA</small>
                </div>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>HTTP Port *</label>
                  <input
                    type="number"
                    name="http_port"
                    value={formData.http_port}
                    onChange={handleInputChange}
                    min="1024"
                    max="65535"
                    required
                  />
                </div>

                <div className="form-group">
                  <label>HTTPS Port</label>
                  <input
                    type="number"
                    name="https_port"
                    value={formData.https_port}
                    onChange={handleInputChange}
                    min="1024"
                    max="65535"
                  />
                </div>
              </div>
            </div>

            {/* Network Configuration */}
            <div className="config-section">
              <h3>Network & Load Balancing</h3>
              
              <div className="form-row">
                <div className="form-group">
                  <label>Virtual IP (Optional)</label>
                  <input
                    type="text"
                    name="virtual_ip"
                    value={formData.virtual_ip}
                    onChange={handleInputChange}
                    placeholder="192.168.0.100"
                  />
                  <small>Keepalived VIP for failover</small>
                </div>

                <div className="form-group">
                  <label>Load Balance Mode</label>
                  <select
                    name="load_balance_mode"
                    value={formData.load_balance_mode}
                    onChange={handleInputChange}
                  >
                    <option value="round_robin">Round Robin</option>
                    <option value="least_conn">Least Connections</option>
                    <option value="ip_hash">IP Hash</option>
                    <option value="random">Random</option>
                  </select>
                </div>
              </div>
            </div>

            {/* Performance Configuration */}
            <div className="config-section">
              <h3>Performance</h3>
              
              <div className="form-row">
                <div className="form-group">
                  <label>Worker Processes</label>
                  <input
                    type="number"
                    name="worker_processes"
                    value={formData.worker_processes}
                    onChange={handleInputChange}
                    min="1"
                    max="16"
                  />
                </div>

                <div className="form-group">
                  <label>Worker Connections</label>
                  <input
                    type="number"
                    name="worker_connections"
                    value={formData.worker_connections}
                    onChange={handleInputChange}
                    min="512"
                    max="65535"
                  />
                </div>
              </div>
            </div>

            {/* Features */}
            <div className="config-section">
              <h3>Features</h3>
              
              <div className="checkbox-group">
                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    name="ssl_enabled"
                    checked={formData.ssl_enabled}
                    onChange={handleInputChange}
                  />
                  <span>Enable SSL/TLS</span>
                </label>

                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    name="gzip_enabled"
                    checked={formData.gzip_enabled}
                    onChange={handleInputChange}
                  />
                  <span>Enable Gzip Compression</span>
                </label>

                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    name="health_check_enabled"
                    checked={formData.health_check_enabled}
                    onChange={handleInputChange}
                  />
                  <span>Enable Health Check</span>
                </label>

                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    name="rate_limit_enabled"
                    checked={formData.rate_limit_enabled}
                    onChange={handleInputChange}
                  />
                  <span>Enable Rate Limiting</span>
                </label>

                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    name="cache_enabled"
                    checked={formData.cache_enabled}
                    onChange={handleInputChange}
                  />
                  <span>Enable Caching</span>
                </label>
              </div>

              {formData.health_check_enabled && (
                <div className="form-group">
                  <label>Health Check Path</label>
                  <input
                    type="text"
                    name="health_check_path"
                    value={formData.health_check_path}
                    onChange={handleInputChange}
                    placeholder="/health"
                  />
                </div>
              )}
            </div>
          </div>

          <div className="modal-footer">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={onClose}
              disabled={isCreating}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isCreating}
            >
              {isCreating ? 'Creating...' : 'Create Cluster'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateNginxClusterModal;

