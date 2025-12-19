import React, { useState, useEffect } from 'react';
import { 
  Rocket, 
  Plus, 
  Trash2, 
  RefreshCw, 
  CheckCircle, 
  XCircle, 
  Loader, 
  Server, 
  Database,
  ExternalLink,
  Activity,
  ChevronDown,
  ChevronUp,
  Copy,
  Terminal
} from 'lucide-react';
import { deployAPI } from '../api';
import Layout from './common/Layout';
import toast from 'react-hot-toast';
import './Deploy.css';

// Predefined image templates
const IMAGE_TEMPLATES = [
  {
    id: 'detection-engine',
    name: 'Detection Engine',
    image: 'iaas-detection-engine:1.0',
    description: 'Log detection engine with ClickHouse + PostgreSQL (auto-provisioned)',
    defaultEnv: {
      CH_HOST: 'auto',
      CH_PORT: '9000',
      DB_USER: 'default',
      DB_PASSWORD: '',
      DB_NAME: 'detection_db',
      PG_HOST: 'auto',
      PG_PORT: '5432',
      PG_USER: 'postgres',
      PG_PASSWORD: 'postgres123',
      PG_DATABASE: 'alerts_db',
      LOG_LEVEL: 'info',
      LOG_TABLE: 'log_edr_v2',
      MATCHING_TABLE: 'matching',
      ALERT_TABLE: 'alert',
      LOOKUP_TABLE: 'matching_lists',
      TIME_QUERY_WINDOW: '2h',
      RULES_STORAGE_PATH: '/opt/rules_storage',
      LOG_FILE_PATH: '/var/log/detection-engine/engine.log',
      API_PORT: '8000'
    },
    exposedPort: 8000,
    requiredInfra: ['clickhouse', 'postgresql']
  },
  {
    id: 'custom',
    name: 'Custom Image',
    image: '',
    description: 'Deploy any Docker image with custom configuration',
    defaultEnv: {},
    exposedPort: 8080,
    requiredInfra: []
  }
];

const Deploy = ({ onLogout, onNavigate }) => {
  const [deployments, setDeployments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [deploying, setDeploying] = useState(false);
  const [showDeployForm, setShowDeployForm] = useState(false);
  const [expandedDeployment, setExpandedDeployment] = useState(null);
  
  // Form state - default to detection-engine body
  const [selectedTemplate, setSelectedTemplate] = useState(IMAGE_TEMPLATES[0]);
  const [formData, setFormData] = useState({
    name: "detection-engine-test",
    image: "iaas-detection-engine:1.0",
    environment: {
      CH_HOST: "auto",
      CH_PORT: "9000",
      DB_USER: "default",
      DB_PASSWORD: "",
      DB_NAME: "detection_db",
      PG_HOST: "auto",
      PG_PORT: "5432",
      PG_USER: "postgres",
      PG_PASSWORD: "postgres123",
      PG_DATABASE: "alerts_db",
      LOG_LEVEL: "info",
      LOG_TABLE: "log_edr_v2",
      MATCHING_TABLE: "matching",
      ALERT_TABLE: "alert",
      LOOKUP_TABLE: "matching_lists",
      TIME_QUERY_WINDOW: "2h",
      RULES_STORAGE_PATH: "/opt/rules_storage",
      LOG_FILE_PATH: "/var/log/detection-engine/engine.log",
      API_PORT: "8000"
    },
    exposedPort: 8000,
    cpu: 1,
    memory: 512
  });
  const [customEnvKey, setCustomEnvKey] = useState('');
  const [customEnvValue, setCustomEnvValue] = useState('');

  useEffect(() => {
    loadDeployments();
    const interval = setInterval(loadDeployments, 10000);
    return () => clearInterval(interval);
  }, []);

  const loadDeployments = async () => {
    try {
      setLoading(true);
      // Try to get from API first
      const response = await deployAPI.list();
      let deploymentsData = response.data?.data?.deployments || response.data?.deployments || [];
      
      // If API returns empty (not implemented yet), use localStorage
      if (deploymentsData.length === 0) {
        const savedDeployments = localStorage.getItem('iaas_deployments');
        if (savedDeployments) {
          deploymentsData = JSON.parse(savedDeployments);
        }
      }
      
      setDeployments(deploymentsData);
    } catch (error) {
      if (error.response?.status !== 401) {
        console.error('Error loading deployments:', error);
        // Fallback to localStorage
        const savedDeployments = localStorage.getItem('iaas_deployments');
        if (savedDeployments) {
          setDeployments(JSON.parse(savedDeployments));
          return;
        }
      }
      setDeployments([]);
    } finally {
      setLoading(false);
    }
  };

  const handleTemplateChange = (templateId) => {
    const template = IMAGE_TEMPLATES.find(t => t.id === templateId);
    if (template) {
      setSelectedTemplate(template);
      setFormData(prev => ({
        ...prev,
        image: template.image,
        environment: { ...template.defaultEnv },
        exposedPort: template.exposedPort
      }));
    }
  };

  const handleEnvChange = (key, value) => {
    setFormData(prev => ({
      ...prev,
      environment: { ...prev.environment, [key]: value }
    }));
  };

  const handleAddCustomEnv = () => {
    if (customEnvKey && customEnvValue) {
      handleEnvChange(customEnvKey, customEnvValue);
      setCustomEnvKey('');
      setCustomEnvValue('');
    }
  };

  const handleRemoveEnv = (key) => {
    setFormData(prev => {
      const newEnv = { ...prev.environment };
      delete newEnv[key];
      return { ...prev, environment: newEnv };
    });
  };

  const handleDeploy = async (e) => {
    e.preventDefault();
    
    if (!formData.name || !formData.image) {
      toast.error('Please fill in required fields');
      return;
    }

    setDeploying(true);
    try {
      const payload = {
        name: formData.name,
        image: formData.image,
        environment: formData.environment,
        volumes: [],
        exposed_port: parseInt(formData.exposedPort),
        cpu: parseFloat(formData.cpu),
        memory: parseInt(formData.memory)
      };

      const response = await deployAPI.deploy(payload);
      const newDeployment = response.data?.data || response.data;
      
      // Save to localStorage for persistence (since backend List API may not be implemented)
      if (newDeployment && newDeployment.deployment_id) {
        const savedDeployments = localStorage.getItem('iaas_deployments');
        let allDeployments = savedDeployments ? JSON.parse(savedDeployments) : [];
        // Check if deployment already exists
        const existingIndex = allDeployments.findIndex(d => d.deployment_id === newDeployment.deployment_id);
        if (existingIndex >= 0) {
          allDeployments[existingIndex] = newDeployment;
        } else {
          allDeployments.unshift(newDeployment);
        }
        localStorage.setItem('iaas_deployments', JSON.stringify(allDeployments));
      }
      
      toast.success('Deployment started successfully!');
      setShowDeployForm(false);
      setFormData({
        name: 'detection-engine-test',
        image: 'iaas-detection-engine:1.0',
        environment: {
          CH_HOST: 'auto',
          CH_PORT: '9000',
          DB_USER: 'default',
          DB_PASSWORD: '',
          DB_NAME: 'detection_db',
          PG_HOST: 'auto',
          PG_PORT: '5432',
          PG_USER: 'postgres',
          PG_PASSWORD: 'postgres123',
          PG_DATABASE: 'alerts_db',
          LOG_LEVEL: 'info',
          LOG_TABLE: 'log_edr_v2',
          MATCHING_TABLE: 'matching',
          ALERT_TABLE: 'alert',
          LOOKUP_TABLE: 'matching_lists',
          TIME_QUERY_WINDOW: '2h',
          RULES_STORAGE_PATH: '/opt/rules_storage',
          LOG_FILE_PATH: '/var/log/detection-engine/engine.log',
          API_PORT: '8000'
        },
        exposedPort: 8000,
        cpu: 1,
        memory: 512
      });
      loadDeployments();
    } catch (error) {
      console.error('Deploy error:', error);
      toast.error(error.response?.data?.message || 'Deployment failed');
    } finally {
      setDeploying(false);
    }
  };

  const handleDeleteDeployment = async (id) => {
    if (!window.confirm('Are you sure you want to delete this deployment and its infrastructure?')) {
      return;
    }
    try {
      await deployAPI.delete(id);
      // Also remove from localStorage
      const savedDeployments = localStorage.getItem('iaas_deployments');
      if (savedDeployments) {
        const allDeployments = JSON.parse(savedDeployments);
        const updatedDeployments = allDeployments.filter(d => d.deployment_id !== id);
        localStorage.setItem('iaas_deployments', JSON.stringify(updatedDeployments));
      }
      toast.success('Deployment deleted');
      loadDeployments();
    } catch (error) {
      // Still remove from localStorage even if API fails
      const savedDeployments = localStorage.getItem('iaas_deployments');
      if (savedDeployments) {
        const allDeployments = JSON.parse(savedDeployments);
        const updatedDeployments = allDeployments.filter(d => d.deployment_id !== id);
        localStorage.setItem('iaas_deployments', JSON.stringify(updatedDeployments));
        loadDeployments();
      }
      toast.error('Failed to delete deployment from server, removed from local list');
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    toast.success('Copied to clipboard');
  };

  const getStatusClass = (status) => {
    switch (status?.toLowerCase()) {
      case 'running': return 'status-running';
      case 'creating':
      case 'deploying': return 'status-creating';
      case 'failed': return 'status-failed';
      case 'stopped': return 'status-stopped';
      default: return 'status-unknown';
    }
  };

  const getStatusIcon = (status) => {
    switch (status?.toLowerCase()) {
      case 'running': return <CheckCircle size={16} />;
      case 'creating':
      case 'deploying': return <Loader size={16} className="spin" />;
      case 'failed': return <XCircle size={16} />;
      default: return <Activity size={16} />;
    }
  };

  return (
    <Layout onLogout={onLogout} activeTab="deploy" onNavigate={onNavigate}>
      <div className="deploy-page">
        <div className="deploy-header">
          <div className="header-info">
            <h1><Rocket size={28} /> Auto Deploy</h1>
            <p>Deploy containers with auto-provisioned infrastructure</p>
          </div>
          <div className="header-actions">
            <button className="btn btn-secondary" onClick={loadDeployments}>
              <RefreshCw size={18} />
              Refresh
            </button>
            <button className="btn btn-primary" onClick={() => setShowDeployForm(true)}>
              <Plus size={18} />
              New Deployment
            </button>
          </div>
        </div>

        {/* Deploy Form Modal */}
        {showDeployForm && (
          <div className="modal-overlay" onClick={() => setShowDeployForm(false)}>
            <div className="deploy-modal" onClick={e => e.stopPropagation()}>
              <div className="modal-header">
                <h2><Rocket size={24} /> New Deployment</h2>
                <button className="close-btn" onClick={() => setShowDeployForm(false)}>&times;</button>
              </div>
              
              <form onSubmit={handleDeploy} className="deploy-form">
                {/* Template Selection */}
                <div className="form-group">
                  <label>Image Template</label>
                  <div className="template-selector">
                    {IMAGE_TEMPLATES.map(template => (
                      <div 
                        key={template.id}
                        className={`template-card ${selectedTemplate.id === template.id ? 'selected' : ''}`}
                        onClick={() => handleTemplateChange(template.id)}
                      >
                        <Server size={24} />
                        <div className="template-info">
                          <strong>{template.name}</strong>
                          <span>{template.description}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Basic Info */}
                <div className="form-row">
                  <div className="form-group">
                    <label>Deployment Name *</label>
                    <input
                      type="text"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      placeholder="my-app-deployment"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label>Docker Image *</label>
                    <input
                      type="text"
                      value={formData.image}
                      onChange={(e) => setFormData({ ...formData, image: e.target.value })}
                      placeholder="image:tag"
                      required
                    />
                  </div>
                </div>

                {/* Resources */}
                <div className="form-row">
                  <div className="form-group">
                    <label>Exposed Port</label>
                    <input
                      type="number"
                      value={formData.exposedPort}
                      onChange={(e) => setFormData({ ...formData, exposedPort: e.target.value })}
                    />
                  </div>
                  <div className="form-group">
                    <label>CPU (cores)</label>
                    <input
                      type="number"
                      step="0.5"
                      min="0.5"
                      value={formData.cpu}
                      onChange={(e) => setFormData({ ...formData, cpu: e.target.value })}
                    />
                  </div>
                  <div className="form-group">
                    <label>Memory (MB)</label>
                    <input
                      type="number"
                      step="128"
                      min="128"
                      value={formData.memory}
                      onChange={(e) => setFormData({ ...formData, memory: e.target.value })}
                    />
                  </div>
                </div>

                {/* Environment Variables */}
                <div className="form-group">
                  <label>Environment Variables</label>
                  <div className="env-hint">
                    <span className="hint-badge">💡 Tip:</span> Set <code>CH_HOST=auto</code> or <code>PG_HOST=auto</code> to auto-provision ClickHouse or PostgreSQL
                  </div>
                  <div className="env-vars-container">
                    {Object.entries(formData.environment).map(([key, value]) => (
                      <div key={key} className="env-var-row">
                        <input 
                          type="text" 
                          value={key} 
                          readOnly 
                          className="env-key"
                        />
                        <input 
                          type="text" 
                          value={value}
                          onChange={(e) => handleEnvChange(key, e.target.value)}
                          className="env-value"
                        />
                        <button 
                          type="button" 
                          className="btn-icon danger"
                          onClick={() => handleRemoveEnv(key)}
                        >
                          <Trash2 size={16} />
                        </button>
                      </div>
                    ))}
                    <div className="env-var-row add-new">
                      <input
                        type="text"
                        placeholder="KEY"
                        value={customEnvKey}
                        onChange={(e) => setCustomEnvKey(e.target.value.toUpperCase())}
                        className="env-key"
                      />
                      <input
                        type="text"
                        placeholder="value"
                        value={customEnvValue}
                        onChange={(e) => setCustomEnvValue(e.target.value)}
                        className="env-value"
                      />
                      <button 
                        type="button" 
                        className="btn-icon primary"
                        onClick={handleAddCustomEnv}
                      >
                        <Plus size={16} />
                      </button>
                    </div>
                  </div>
                </div>

                {/* Detected Infrastructure */}
                {selectedTemplate.requiredInfra.length > 0 && (
                  <div className="form-group">
                    <label>Auto-Created Infrastructure</label>
                    <div className="infra-preview">
                      {selectedTemplate.requiredInfra.includes('clickhouse') && (
                        <div className="infra-chip clickhouse">
                          <Database size={16} />
                          ClickHouse
                        </div>
                      )}
                      {selectedTemplate.requiredInfra.includes('postgresql') && (
                        <div className="infra-chip postgresql">
                          <Database size={16} />
                          PostgreSQL HA
                        </div>
                      )}
                    </div>
                  </div>
                )}

                <div className="form-actions">
                  <button type="button" className="btn btn-secondary" onClick={() => setShowDeployForm(false)}>
                    Cancel
                  </button>
                  <button type="submit" className="btn btn-primary" disabled={deploying}>
                    {deploying ? (
                      <>
                        <Loader size={18} className="spin" />
                        Deploying...
                      </>
                    ) : (
                      <>
                        <Rocket size={18} />
                        Deploy
                      </>
                    )}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* Deployments List */}
        <div className="deployments-section">
          <h2>Deployments</h2>
          
          {loading ? (
            <div className="loading-state">
              <Loader size={32} className="spin" />
              <p>Loading deployments...</p>
            </div>
          ) : deployments.length === 0 ? (
            <div className="empty-state">
              <Rocket size={48} />
              <h3>No deployments yet</h3>
              <p>Deploy your first container with auto-provisioned infrastructure</p>
              <button className="btn btn-primary" onClick={() => setShowDeployForm(true)}>
                <Plus size={18} />
                Create Deployment
              </button>
            </div>
          ) : (
            <div className="deployments-list">
              {deployments.map(deployment => (
                <div key={deployment.deployment_id} className="deployment-card">
                  <div className="deployment-header" onClick={() => 
                    setExpandedDeployment(expandedDeployment === deployment.deployment_id ? null : deployment.deployment_id)
                  }>
                    <div className="deployment-info">
                      <div className={`status-badge ${getStatusClass(deployment.status)}`}>
                        {getStatusIcon(deployment.status)}
                        {deployment.status}
                      </div>
                      <h3>{deployment.name}</h3>
                    </div>
                    <div className="deployment-actions">
                      <button 
                        className="btn-icon danger"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeleteDeployment(deployment.deployment_id);
                        }}
                        title="Delete Deployment"
                      >
                        <Trash2 size={18} />
                      </button>
                      {expandedDeployment === deployment.deployment_id ? 
                        <ChevronUp size={20} /> : <ChevronDown size={20} />
                      }
                    </div>
                  </div>

                  {expandedDeployment === deployment.deployment_id && (
                    <div className="deployment-details">
                      {/* Container Info */}
                      <div className="detail-section">
                        <h4><Terminal size={16} /> Container</h4>
                        <div className="detail-grid">
                          <div className="detail-item">
                            <span className="label">Name:</span>
                            <span className="value">{deployment.container?.name}</span>
                          </div>
                          <div className="detail-item">
                            <span className="label">Status:</span>
                            <span className="value">{deployment.container?.status}</span>
                          </div>
                          {deployment.container?.endpoint && (
                            <div className="detail-item">
                              <span className="label">Endpoint:</span>
                              <span className="value endpoint">
                                <a href={deployment.container.endpoint} target="_blank" rel="noopener noreferrer">
                                  {deployment.container.endpoint}
                                  <ExternalLink size={14} />
                                </a>
                                <button className="btn-copy" onClick={() => copyToClipboard(deployment.container.endpoint)}>
                                  <Copy size={14} />
                                </button>
                              </span>
                            </div>
                          )}
                        </div>
                      </div>

                      {/* Infrastructure */}
                      {deployment.created_infrastructure?.length > 0 && (
                        <div className="detail-section">
                          <h4><Database size={16} /> Auto-Created Infrastructure</h4>
                          <div className="infra-list">
                            {deployment.created_infrastructure.map((infra, idx) => (
                              <div key={idx} className="infra-item">
                                <div className={`infra-type ${infra.type}`}>
                                  <Database size={16} />
                                  {infra.type}
                                </div>
                                <div className="infra-details">
                                  <span className="infra-name">{infra.name}</span>
                                  <span className={`infra-status ${infra.status}`}>{infra.status}</span>
                                  <span className="infra-endpoint">{infra.endpoint}</span>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* Endpoints */}
                      {deployment.endpoints && Object.keys(deployment.endpoints).length > 0 && (
                        <div className="detail-section">
                          <h4><ExternalLink size={16} /> Endpoints</h4>
                          <div className="endpoints-list">
                            {Object.entries(deployment.endpoints).map(([key, url]) => (
                              <div key={key} className="endpoint-item">
                                <span className="endpoint-key">{key}:</span>
                                <a href={url} target="_blank" rel="noopener noreferrer" className="endpoint-url">
                                  {url}
                                  <ExternalLink size={14} />
                                </a>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
};

export default Deploy;
