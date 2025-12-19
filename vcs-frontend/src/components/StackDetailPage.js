import React, { useState, useEffect, useCallback } from 'react';
import {
  ArrowLeft, RefreshCw, Play, Square, RotateCcw, Trash2,
  Database, Globe, Box, Server, ChevronDown, ChevronRight,
  Copy, Download, AlertTriangle, CheckCircle,
  XCircle, Loader, Crown, Users, Plus, Minus
} from 'lucide-react';
import { stackAPI, clusterAPI, nginxAPI, dinDAPI, monitoringAPI } from '../api';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import toast from 'react-hot-toast';
import Layout from './common/Layout';
import './StackDetail.css';

// Simple Status Badge
const StatusBadge = ({ status }) => {
  const getClass = () => {
    switch (status?.toLowerCase()) {
      case 'running': return 'badge badge-success';
      case 'stopped': return 'badge badge-warning';
      case 'failed': return 'badge badge-danger';
      case 'creating': return 'badge badge-info';
      default: return 'badge badge-default';
    }
  };
  return <span className={getClass()}>{status || 'unknown'}</span>;
};

// Role Badge for cluster nodes
const RoleBadge = ({ role }) => {
  if (role === 'leader' || role === 'primary') {
    return <span className="badge badge-primary"><Crown size={12} /> Primary</span>;
  }
  return <span className="badge badge-secondary"><Users size={12} /> Replica</span>;
};

const StackDetailPage = ({ stackId, onBack, onLogout }) => {
  const [stack, setStack] = useState(null);
  const [loading, setLoading] = useState(true);
  const [expandedResources, setExpandedResources] = useState({});
  const [resourceDetails, setResourceDetails] = useState({});
  const [loadingDetails, setLoadingDetails] = useState({});
  const [actionLoading, setActionLoading] = useState({});
  const [activeTab, setActiveTab] = useState('overview'); // overview, metrics, logs
  const [logs, setLogs] = useState({});

  const loadStack = useCallback(async () => {
    try {
      setLoading(true);
      const response = await stackAPI.getById(stackId);
      setStack(response.data?.data || response.data);
    } catch (error) {
      console.error('Error loading service:', error);
      toast.error('Failed to load service details');
    } finally {
      setLoading(false);
    }
  }, [stackId]);

  useEffect(() => {
    loadStack();
  }, [loadStack]);

  // Auto-expand if single resource
  useEffect(() => {
    if (stack?.resources?.length === 1) {
      const res = stack.resources[0];
      const key = res.id || res.resource_name;
      if (!expandedResources[key]) {
        setExpandedResources(prev => ({ ...prev, [key]: true }));
        loadResourceDetails(res);
      }
    }
  }, [stack]);

  const toggleResource = async (resource) => {
    const key = resource.id || resource.resource_name;
    const isExpanded = expandedResources[key];

    setExpandedResources(prev => ({ ...prev, [key]: !isExpanded }));

    if (!isExpanded && !resourceDetails[key]) {
      await loadResourceDetails(resource);
    }
  };

  const loadResourceDetails = async (resource) => {
    const key = resource.id || resource.resource_name;
    const infraId = resource.outputs?.cluster_id || resource.infrastructure_id;

    if (!infraId) return;

    setLoadingDetails(prev => ({ ...prev, [key]: true }));
    try {
      let response;
      switch (resource.resource_type) {
        case 'POSTGRES_CLUSTER':
          response = await clusterAPI.getById(infraId);
          break;
        case 'NGINX_GATEWAY':
          response = await nginxAPI.getById(infraId);
          break;
        case 'DOCKER_SERVICE':
          response = await dinDAPI.getById(infraId);
          break;
        default:
          return;
      }
      setResourceDetails(prev => ({
        ...prev,
        [key]: response.data?.data || response.data
      }));
    } catch (error) {
      console.error('Error loading details:', error);
      // toast.error('Failed to load resource details');
    } finally {
      setLoadingDetails(prev => ({ ...prev, [key]: false }));
    }
  };

  const loadLogs = async (resource) => {
    const key = resource.id || resource.resource_name;
    const infraId = resource.outputs?.cluster_id || resource.infrastructure_id;

    if (!infraId) return;

    try {
      let response;
      let logData;
      switch (resource.resource_type) {
        case 'POSTGRES_CLUSTER':
          response = await clusterAPI.getLogs(infraId, 100);
          logData = response.data?.data?.logs || response.data?.logs || [];
          break;
        case 'NGINX_GATEWAY':
          response = await nginxAPI.getLogs(infraId, 100);
          logData = response.data?.data || response.data;
          break;
        case 'DOCKER_SERVICE':
          response = await dinDAPI.getLogs(infraId, 100);
          logData = response.data?.data?.logs || response.data?.logs || response.data?.data || response.data;
          break;
        default:
          return;
      }
      setLogs(prev => ({ ...prev, [key]: logData }));
      toast.success('Logs updated');
    } catch (error) {
      console.error('Error loading logs:', error);
      toast.error('Failed to load logs');
    }
  };

  const handleAction = async (actionKey, action, resource, nodeId = null) => {
    setActionLoading(prev => ({ ...prev, [actionKey]: true }));
    const infraId = resource.outputs?.cluster_id || resource.infrastructure_id;

    try {
      switch (resource.resource_type) {
        case 'POSTGRES_CLUSTER':
          if (action === 'start') await clusterAPI.start(infraId);
          else if (action === 'stop') await clusterAPI.stop(infraId);
          else if (action === 'restart') await clusterAPI.restart(infraId);
          else if (action === 'failover') await clusterAPI.failover(infraId, nodeId);
          else if (action === 'backup') await clusterAPI.backup(infraId, { type: 'full' });
          else if (action === 'scale_up') {
            const currentNodes = resourceDetails[resource.id]?.nodes?.length || 1;
            await clusterAPI.scale(infraId, currentNodes + 1);
          } else if (action === 'scale_down') {
            const currentNodes = resourceDetails[resource.id]?.nodes?.length || 2;
            await clusterAPI.scale(infraId, Math.max(1, currentNodes - 1));
          }
          break;
        case 'NGINX_GATEWAY':
          if (action === 'start') await nginxAPI.start(infraId);
          else if (action === 'stop') await nginxAPI.stop(infraId);
          else if (action === 'restart') await nginxAPI.restart(infraId);
          else if (action === 'reload') await nginxAPI.reload(infraId);
          break;
        case 'DOCKER_SERVICE':
          if (action === 'start') await dinDAPI.start(infraId);
          else if (action === 'stop') await dinDAPI.stop(infraId);
          else if (action === 'restart') await dinDAPI.restart(infraId);
          break;
        default: break;
      }
      toast.success(`${action} successful`);
      await loadResourceDetails(resource);
    } catch (error) {
      toast.error(`${action} failed: ${error.message}`);
    } finally {
      setActionLoading(prev => ({ ...prev, [actionKey]: false }));
    }
  };

  const handleServiceDelete = async () => {
    if (window.confirm('Delete this infrastructure service? This action cannot be undone.')) {
      try {
        await stackAPI.delete(stackId);
        toast.success('Service deleted');
        onBack();
      } catch (error) {
        toast.error('Failed to delete service');
      }
    }
  };

  if (loading) {
    return (
      <Layout onLogout={onLogout} activeTab="stacks">
        <div className="stack-detail">
          <div className="loading-container">
            <Loader size={32} className="spin" />
            <p>Loading details...</p>
          </div>
        </div>
      </Layout>
    );
  }

  if (!stack) {
    return (
      <Layout onLogout={onLogout} activeTab="stacks">
        <div className="stack-detail">
          <div className="error-container">
            <AlertTriangle size={48} />
            <p>Service not found</p>
            <button className="btn btn-primary" onClick={onBack}>Back to Dashboard</button>
          </div>
        </div>
      </Layout>
    );
  }

  const resources = stack.resources || [];
  const isSingleResource = resources.length === 1;

  return (
    <Layout onLogout={onLogout} activeTab="stacks">
      <div className="stack-detail">
        {/* Header */}
        <header className="detail-header">
          <button className="btn-back" onClick={onBack}>
            <ArrowLeft size={20} /> Dashboard
          </button>
          <div className="header-info">
            <h1>{stack.name}</h1>
            <StatusBadge status={stack.status} />
            <span className="env-tag">{stack.environment}</span>
          </div>
          <div className="header-actions">
            <button className="btn btn-icon" onClick={loadStack} title="Refresh"><RefreshCw size={18} /></button>
            <button className="btn btn-danger" onClick={handleServiceDelete} title="Delete Service"><Trash2 size={16} /></button>
          </div>
        </header>

        {/* Tabs */}
        <div className="tabs">
          <button className={`tab ${activeTab === 'overview' ? 'active' : ''}`} onClick={() => setActiveTab('overview')}>
            Overview
          </button>
          <button className={`tab ${activeTab === 'metrics' ? 'active' : ''}`} onClick={() => setActiveTab('metrics')}>
            Metrics
          </button>
          <button className={`tab ${activeTab === 'logs' ? 'active' : ''}`} onClick={() => setActiveTab('logs')}>
            Logs
          </button>
        </div>

        {/* Content */}
        <div className="detail-content">
          {activeTab === 'overview' && (
            <div className="resources-list">
              {!isSingleResource && <h2>Components ({resources.length})</h2>}
              {resources.length === 0 ? (
                <div className="empty">No components found</div>
              ) : (
                resources.map((resource, idx) => {
                  const key = resource.id || `r-${idx}`;
                  const isExpanded = expandedResources[key];
                  const details = resourceDetails[key];
                  const isLoading = loadingDetails[key];

                  if (isSingleResource) {
                    return (
                      <div key={key} className="resource-body-single">
                        {isLoading ? (
                          <div className="loading"><Loader size={20} className="spin" /> Loading configuration...</div>
                        ) : details ? (
                          <ResourceDetails
                            type={resource.resource_type}
                            resource={resource}
                            details={details}
                            onAction={(action, nodeId) => handleAction(`${key}-${action}`, action, { ...resource, id: key }, nodeId)}
                            actionLoading={actionLoading}
                            actionKey={key}
                          />
                        ) : (
                          <div className="error">
                            <p>Waiting for provisioning to complete...</p>
                          </div>
                        )}
                      </div>
                    );
                  }

                  return (
                    <div key={key} className="resource-card">
                      <div className="resource-header" onClick={() => toggleResource({ ...resource, id: key })}>
                        <div className="resource-title">
                          {isExpanded ? <ChevronDown size={18} /> : <ChevronRight size={18} />}
                          <ResourceIcon type={resource.resource_type} />
                          <span className="name">{resource.resource_name}</span>
                          <span className="type">{resource.resource_type?.replace(/_/g, ' ')}</span>
                        </div>
                        <StatusBadge status={resource.status} />
                      </div>
                      {isExpanded && (
                        <div className="resource-body">
                          {isLoading ? (
                            <div className="loading"><Loader size={20} className="spin" /> Loading...</div>
                          ) : details ? (
                            <ResourceDetails
                              type={resource.resource_type}
                              resource={resource}
                              details={details}
                              onAction={(action, nodeId) => handleAction(`${key}-${action}`, action, { ...resource, id: key }, nodeId)}
                              actionLoading={actionLoading}
                              actionKey={key}
                            />
                          ) : (
                            <div className="error">Could not load details</div>
                          )}
                        </div>
                      )}
                    </div>
                  );
                })
              )}
            </div>
          )}

          {activeTab === 'metrics' && (
            <div className="metrics-list-section">
              {resources.length === 0 ? (
                <div className="empty">No resources to show metrics for.</div>
              ) : (
                resources.map((resource, idx) => {
                  const key = resource.id || `r-${idx}`;
                  return (
                    <div key={key} className="metric-resource-card">
                      {!isSingleResource && (
                        <div className="metric-card-header">
                          <ResourceIcon type={resource.resource_type} />
                          <span className="name">{resource.resource_name}</span>
                        </div>
                      )}
                      <ResourceMetrics resource={resource} />
                    </div>
                  );
                })
              )}
            </div>
          )}

          {activeTab === 'logs' && (
            <div className="logs-section">
              {resources.map((resource, idx) => {
                const key = resource.id || `r-${idx}`;
                const resourceLogs = logs[key];

                return (
                  <div key={key} className="log-card">
                    {!isSingleResource && (
                      <div className="log-header">
                        <ResourceIcon type={resource.resource_type} />
                        <span>{resource.resource_name}</span>
                      </div>
                    )}
                    <div className="log-toolbar">
                      <button
                        className="btn btn-sm btn-secondary"
                        onClick={() => loadLogs({ ...resource, id: key })}
                      >
                        <RefreshCw size={14} /> Refresh Logs
                      </button>
                    </div>
                    <div className="log-content">
                      {resourceLogs ? (
                        <LogsDisplay logs={resourceLogs} resourceType={resource.resource_type} />
                      ) : (
                        <p className="no-logs">Click "Refresh Logs" to fetch logs.</p>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
};

// --- Helper Components ---

const ResourceIcon = ({ type }) => {
  switch (type?.toUpperCase()) {
    case 'POSTGRES_CLUSTER': return <Database size={18} className="icon-db" />;
    case 'NGINX_GATEWAY': return <Globe size={18} className="icon-nginx" />;
    case 'DIND_ENVIRONMENT': return <Box size={18} className="icon-docker" />;
    default: return <Server size={18} />;
  }
};

const LogsDisplay = ({ logs, resourceType }) => {
  if (!logs) return null;
  if (resourceType === 'POSTGRES_CLUSTER' && Array.isArray(logs)) {
    if (logs.length === 0) return <p className="no-logs">No logs available</p>;
    return (
      <div className="cluster-logs">
        {logs.map((nodeLog, idx) => (
          <div key={idx} className="node-log-section">
            <div className="node-log-header">
              <Database size={14} />
              <span className="node-name">{nodeLog.node_name}</span>
              {nodeLog.timestamp && <span className="timestamp">{nodeLog.timestamp}</span>}
            </div>
            <pre className="log-text">{nodeLog.logs || 'No logs'}</pre>
          </div>
        ))}
      </div>
    );
  }
  if (Array.isArray(logs)) {
    if (logs.length === 0) return <p className="no-logs">No logs available</p>;
    return <pre className="log-text">{logs.join('\n')}</pre>;
  }
  if (typeof logs === 'string') return <pre className="log-text">{logs || 'No logs available'}</pre>;
  return <pre className="log-text">{JSON.stringify(logs, null, 2)}</pre>;
};

const ResourceDetails = ({ type, resource, details, onAction, actionLoading, actionKey }) => {
  switch (type) {
    case 'POSTGRES_CLUSTER':
      return <ClusterDetails resource={resource} details={details} onAction={onAction} actionLoading={actionLoading} actionKey={actionKey} />;
    case 'NGINX_GATEWAY':
      return <NginxDetails resource={resource} details={details} onAction={onAction} actionLoading={actionLoading} actionKey={actionKey} />;
    case 'DIND_ENVIRONMENT':
      return <DockerDetails resource={resource} details={details} onAction={onAction} actionLoading={actionLoading} actionKey={actionKey} />;
    case 'DOCKER_SERVICE':
      return <DockerDetails resource={resource} details={details} onAction={onAction} actionLoading={actionLoading} actionKey={actionKey} />;
    default:
      return <div>Unknown resource type: {type}</div>;
  }
};

const ClusterDetails = ({ resource, details, onAction, actionLoading, actionKey }) => {
  const nodes = details.nodes || [];
  const clusterId = resource?.outputs?.cluster_id || details?.cluster_id || 'N/A';
  const infraId = resource?.infrastructure_id || details?.infrastructure_id || 'N/A';
  
  const copyToClipboard = (text, label) => {
    navigator.clipboard.writeText(text);
    toast.success(`${label} copied!`);
  };

  return (
    <div className="cluster-details">
      {/* ID Cards */}
      <div className="id-cards" style={{ display: 'flex', gap: '12px', marginBottom: '16px', flexWrap: 'wrap' }}>
        <div className="id-card" style={{ 
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', 
          padding: '12px 16px', 
          borderRadius: '8px', 
          color: 'white',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          minWidth: '300px'
        }}>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: '11px', opacity: 0.8, marginBottom: '4px' }}>CLUSTER ID</div>
            <div style={{ fontFamily: 'monospace', fontSize: '12px', wordBreak: 'break-all' }}>{clusterId}</div>
          </div>
          <button onClick={() => copyToClipboard(clusterId, 'Cluster ID')} style={{ background: 'rgba(255,255,255,0.2)', border: 'none', padding: '6px', borderRadius: '4px', cursor: 'pointer', color: 'white' }}>
            <Copy size={14} />
          </button>
        </div>
        <div className="id-card" style={{ 
          background: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)', 
          padding: '12px 16px', 
          borderRadius: '8px', 
          color: 'white',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          minWidth: '300px'
        }}>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: '11px', opacity: 0.8, marginBottom: '4px' }}>INFRASTRUCTURE ID</div>
            <div style={{ fontFamily: 'monospace', fontSize: '12px', wordBreak: 'break-all' }}>{infraId}</div>
          </div>
          <button onClick={() => copyToClipboard(infraId, 'Infrastructure ID')} style={{ background: 'rgba(255,255,255,0.2)', border: 'none', padding: '6px', borderRadius: '4px', cursor: 'pointer', color: 'white' }}>
            <Copy size={14} />
          </button>
        </div>
      </div>

      <div className="action-bar">
        <button className="btn btn-success btn-sm" onClick={() => onAction('start')} disabled={actionLoading[`${actionKey}-start`]}>
          {actionLoading[`${actionKey}-start`] ? <Loader size={14} className="spin" /> : <Play size={14} />} Start
        </button>
        <button className="btn btn-warning btn-sm" onClick={() => onAction('stop')} disabled={actionLoading[`${actionKey}-stop`]}>
          {actionLoading[`${actionKey}-stop`] ? <Loader size={14} className="spin" /> : <Square size={14} />} Stop
        </button>
        <button className="btn btn-secondary btn-sm" onClick={() => onAction('restart')} disabled={actionLoading[`${actionKey}-restart`]}>
          {actionLoading[`${actionKey}-restart`] ? <Loader size={14} className="spin" /> : <RotateCcw size={14} />} Restart
        </button>
        <button className="btn btn-primary btn-sm" onClick={() => onAction('backup')} disabled={actionLoading[`${actionKey}-backup`]}>
          {actionLoading[`${actionKey}-backup`] ? <Loader size={14} className="spin" /> : <Download size={14} />} Backup
        </button>
        <button className="btn btn-sm" onClick={() => onAction('scale_up')} disabled={actionLoading[`${actionKey}-scale_up`]}>
          <Plus size={14} /> Add Node
        </button>
        <button className="btn btn-sm" onClick={() => onAction('scale_down')} disabled={actionLoading[`${actionKey}-scale_down`] || nodes.length <= 1}>
          <Minus size={14} /> Remove Node
        </button>
      </div>

      <div className="info-grid">
        <div className="info-item">
          <span className="label">Cluster Name</span>
          <span className="value">{details.cluster_name}</span>
        </div>
        <div className="info-item">
          <span className="label">Version</span>
          <span className="value">PostgreSQL {details.postgres_version}</span>
        </div>
        <div className="info-item">
          <span className="label">Replication</span>
          <span className="value">{details.replication_mode}</span>
        </div>
        <div className="info-item">
          <span className="label">HAProxy Port</span>
          <span className="value">{details.haproxy_port || 'N/A'}</span>
        </div>
      </div>

      {(details.connection_info || details.write_endpoint) && (
        <div className="endpoints">
          <h4>Connection</h4>
          {details.connection_info ? (
            <div className="connection-card">
              <div className="conn-row"><span className="conn-label">Host:</span><code className="conn-value">{details.connection_info.host}</code></div>
              <div className="conn-row"><span className="conn-label">Port:</span><code className="conn-value">{details.connection_info.port}</code></div>
              <div className="conn-row"><span className="conn-label">Database:</span><code className="conn-value">{details.connection_info.database}</code></div>
              <div className="conn-row"><span className="conn-label">Username:</span><code className="conn-value">{details.connection_info.username}</code></div>
              <div className="conn-string">
                <span className="label">Connection String:</span>
                <div className="conn-box">
                  <code>postgres://{details.connection_info.username}:******@{details.connection_info.host}:{details.connection_info.port}/{details.connection_info.database}</code>
                  <button className="btn-copy" onClick={() => {
                    const pwd = prompt("Enter password (optional):", "");
                    const passPart = pwd ? `:${pwd}` : '';
                    const uri = `postgres://${details.connection_info.username}${passPart}@${details.connection_info.host}:${details.connection_info.port}/${details.connection_info.database}?sslmode=${details.connection_info.ssl_mode || 'disable'}`;
                    navigator.clipboard.writeText(uri);
                    toast.success('Copied!');
                  }}><Copy size={14} /></button>
                </div>
              </div>
            </div>
          ) : (
            <div className="endpoint">
              <span className="label">Write Endpoint:</span>
              <code>{details.write_endpoint.host}:{details.write_endpoint.port}</code>
            </div>
          )}
        </div>
      )}

      <div className="nodes-section">
        <h4>Nodes</h4>
        <table className="nodes-table">
          <thead>
            <tr><th>Node Name</th><th>Role</th><th>Status</th><th>Health</th><th>Replication Lag</th></tr>
          </thead>
          <tbody>
            {nodes.map((node, idx) => (
              <tr key={idx} className={node.role === 'leader' || node.role === 'primary' ? 'row-primary' : ''}>
                <td><div className="node-name"><Database size={14} />{node.node_name}</div></td>
                <td><RoleBadge role={node.role} /></td>
                <td><StatusBadge status={node.status} /></td>
                <td>{node.is_healthy ? <CheckCircle size={16} className="icon-success" /> : <XCircle size={16} className="icon-error" />}</td>
                <td>{node.replication_delay ? `${node.replication_delay} bytes` : '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

const NginxDetails = ({ details, onAction, actionLoading, actionKey }) => (
  <div className="nginx-details">
    <div className="action-bar">
      <button className="btn btn-success btn-sm" onClick={() => onAction('start')} disabled={actionLoading[`${actionKey}-start`]}>Start</button>
      <button className="btn btn-warning btn-sm" onClick={() => onAction('stop')} disabled={actionLoading[`${actionKey}-stop`]}>Stop</button>
      <button className="btn btn-secondary btn-sm" onClick={() => onAction('restart')} disabled={actionLoading[`${actionKey}-restart`]}>Restart</button>
      <button className="btn btn-primary btn-sm" onClick={() => onAction('reload')} disabled={actionLoading[`${actionKey}-reload`]}>Reload</button>
    </div>
    <div className="info-grid">
      <div className="info-item"><span className="label">Name</span><span className="value">{details.name}</span></div>
      <div className="info-item"><span className="label">Port</span><span className="value">{details.port}</span></div>
      <div className="info-item"><span className="label">Status</span><StatusBadge status={details.status} /></div>
    </div>
  </div>
);

const DockerDetails = ({ details, onAction, actionLoading, actionKey }) => (
  <div className="docker-details">
    <div className="action-bar">
      <button className="btn btn-success btn-sm" onClick={() => onAction('start')} disabled={actionLoading[`${actionKey}-start`]}>Start</button>
      <button className="btn btn-warning btn-sm" onClick={() => onAction('stop')} disabled={actionLoading[`${actionKey}-stop`]}>Stop</button>
      <button className="btn btn-secondary btn-sm" onClick={() => onAction('restart')} disabled={actionLoading[`${actionKey}-restart`]}>Restart</button>
    </div>
    <div className="info-grid">
      <div className="info-item"><span className="label">Name</span><span className="value">{details.name}</span></div>
      <div className="info-item"><span className="label">Status</span><StatusBadge status={details.status} /></div>
      <div className="info-item"><span className="label">Image</span><span className="value">{details.image}</span></div>
    </div>
  </div>
);

const ResourceMetrics = ({ resource }) => {
  const [data, setData] = useState([]);
  const [uptime, setUptime] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isDemo, setIsDemo] = useState(false);

  // Generate demo data for display
  const generateDemoData = () => {
    const now = new Date();
    const demoMetrics = [];
    for (let i = 19; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60000);
      demoMetrics.push({
        timestamp: time.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        fullTime: time.toLocaleString(),
        cpu: Math.random() * 30 + 15 + Math.sin(i * 0.3) * 10,
        memory: (Math.random() * 300 + 200).toFixed(2),
        disk_read: Math.floor(Math.random() * 50000),
        disk_write: Math.floor(Math.random() * 30000)
      });
    }
    return demoMetrics;
  };

  const generateDemoUptime = () => ({
    uptime_percent: 99.85 + Math.random() * 0.1,
    total_uptime_seconds: 86000 + Math.random() * 400,
    total_downtime_seconds: Math.random() * 60
  });

  useEffect(() => {
    const fetchData = async () => {
      // Monitoring service indexes metrics/uptime by infrastructure_id (not cluster_id)
      const infraId =
        resource.infrastructure_id ||
        resource.outputs?.infrastructure_id ||
        resource.outputs?.infra_id ||
        resource.outputs?.environment_id ||
        resource.outputs?.cluster_id || // last fallback; might return empty metrics
        resource.id;
      
      try {
        if (!infraId) {
          // No infraId, use demo data
          setData(generateDemoData());
          setUptime(generateDemoUptime());
          setIsDemo(true);
          setLoading(false);
          return;
        }
        
        const [metricsRes, uptimeRes] = await Promise.all([
          monitoringAPI.getHistoricalMetrics(infraId, 0, 20),
          monitoringAPI.getUptime(infraId, '24h')
        ]);
        let mData = metricsRes.data?.data || metricsRes.data || [];
        if (mData.hits?.hits) mData = mData.hits.hits.map(h => h._source);
        const formattedData = Array.isArray(mData) ? mData.map(m => ({
          timestamp: new Date(m.timestamp).toLocaleTimeString(),
          fullTime: new Date(m.timestamp).toLocaleString(),
          cpu: m.cpu_percent || 0,
          memory: m.memory_used ? (m.memory_used / 1024 / 1024).toFixed(2) : 0,
          disk_read: m.disk_read || 0,
          disk_write: m.disk_write || 0
        })).reverse() : [];
        
        // If no real data, use demo data
        if (formattedData.length === 0) {
          setData(generateDemoData());
          setUptime(generateDemoUptime());
          setIsDemo(true);
        } else {
          setData(formattedData);
          setUptime(uptimeRes.data?.data || uptimeRes.data || generateDemoUptime());
          setIsDemo(false);
        }
      } catch (error) {
        console.error('Failed to load metrics, using demo data', error);
        // Fallback to demo data on error
        setData(generateDemoData());
        setUptime(generateDemoUptime());
        setIsDemo(true);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
    
    // Auto-refresh demo data every 15 seconds
    const interval = setInterval(() => {
      if (isDemo) {
        setData(prev => {
          const newPoint = {
            timestamp: new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
            fullTime: new Date().toLocaleString(),
            cpu: Math.random() * 30 + 15,
            memory: (Math.random() * 300 + 200).toFixed(2),
            disk_read: Math.floor(Math.random() * 50000),
            disk_write: Math.floor(Math.random() * 30000)
          };
          return [...prev.slice(1), newPoint];
        });
      }
    }, 15000);
    
    return () => clearInterval(interval);
  }, [resource, isDemo]);

  if (loading) return <div className="loading"><Loader className="spin" /> Loading metrics...</div>;

  return (
    <div className="metrics-container">
      {isDemo && (
        <div className="demo-banner" style={{
          background: 'linear-gradient(90deg, #f59e0b, #d97706)',
          color: 'white',
          padding: '8px 16px',
          borderRadius: '6px',
          marginBottom: '16px',
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          fontSize: '14px'
        }}>
          <span style={{ fontSize: '16px' }}>📊</span>
          <span>Demo Mode: Hiển thị dữ liệu mẫu. Metrics thực sẽ xuất hiện khi monitoring service thu thập dữ liệu.</span>
        </div>
      )}
      {uptime && (
        <div className="uptime-card">
          <div className="uptime-header">
            <h4>System Uptime (24h)</h4>
            <span className={`uptime-status ${uptime.uptime_percent > 99 ? 'good' : 'warning'}`}>{uptime.uptime_percent?.toFixed(2)}%</span>
          </div>
          <div className="uptime-grid">
            <div className="u-item"><span className="label">Total Uptime</span><span className="val">{(uptime.total_uptime_seconds / 3600).toFixed(2)} hrs</span></div>
            <div className="u-item"><span className="label">Downtime</span><span className="val">{(uptime.total_downtime_seconds / 60).toFixed(2)} mins</span></div>
          </div>
        </div>
      )}
      <div className="chart-section">
        <div className="chart-wrapper">
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={data}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="timestamp" />
              <YAxis yAxisId="left" label={{ value: 'CPU %', angle: -90, position: 'insideLeft' }} />
              <YAxis yAxisId="right" orientation="right" label={{ value: 'Mem (MB)', angle: 90, position: 'insideRight' }} />
              <Tooltip labelFormatter={(label, payload) => payload[0]?.payload?.fullTime || label} />
              <Line yAxisId="left" type="monotone" dataKey="cpu" stroke="#8884d8" name="CPU %" />
              <Line yAxisId="right" type="monotone" dataKey="memory" stroke="#82ca9d" name="Memory (MB)" />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
      <div className="analysis-table-section">
        <table className="metrics-table">
          <thead><tr><th>Timestamp</th><th>CPU</th><th>Memory</th><th>Disk Read</th><th>Disk Write</th></tr></thead>
          <tbody>
            {data.slice().reverse().map((row, idx) => (
              <tr key={idx}>
                <td>{row.fullTime}</td>
                <td>{row.cpu.toFixed(2)}%</td>
                <td>{row.memory} MB</td>
                <td>{(row.disk_read / 1024).toFixed(2)} KB</td>
                <td>{(row.disk_write / 1024).toFixed(2)} KB</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default StackDetailPage;
