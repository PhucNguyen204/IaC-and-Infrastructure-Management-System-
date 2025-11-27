import React, { useEffect, useMemo, useState } from 'react';
import {
  Activity,
  AlertTriangle,
  CheckCircle,
  Copy,
  Globe,
  Link2,
  Loader,
  RefreshCw,
  Server,
  Shield,
  Signal,
} from 'lucide-react';
import toast from 'react-hot-toast';
import { nginxClusterAPI } from '../../api';
import './NginxClusterPanel.css';

const STATUS_COLORS = {
  running: 'success',
  healthy: 'success',
  degraded: 'warning',
  stopped: 'muted',
  unhealthy: 'danger',
  failed: 'danger',
};

const NginxClusterPanel = ({ clusterId, resource, onRefresh }) => {
  const [info, setInfo] = useState(null);
  const [health, setHealth] = useState(null);
  const [connectionInfo, setConnectionInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState({});
  const [testResult, setTestResult] = useState(null);
  const [failoverTarget, setFailoverTarget] = useState('');

  const displayClusterId = useMemo(() => clusterId || resource?.outputs?.cluster_id, [clusterId, resource]);

  useEffect(() => {
    if (!displayClusterId) return;
    loadAll();
    
    // Auto-refresh every 5 seconds to update real-time status
    const interval = setInterval(() => {
      loadAll();
    }, 5000);
    
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [displayClusterId]);

  const loadAll = async () => {
    if (!displayClusterId) return;
    try {
      setLoading(true);
      const [infoResp, healthResp, connectionResp] = await Promise.all([
        nginxClusterAPI.getById(displayClusterId),
        nginxClusterAPI.getHealth(displayClusterId).catch(() => null),
        nginxClusterAPI.getConnectionInfo(displayClusterId).catch(() => null),
      ]);

      setInfo(infoResp.data?.data || infoResp.data);
      setHealth(healthResp?.data?.data || healthResp?.data || null);
      setConnectionInfo(connectionResp?.data?.data || connectionResp?.data || null);
      setTestResult(null);
      if (onRefresh) onRefresh();
    } catch (error) {
      console.error('Failed to load nginx cluster info', error);
      toast.error(error.response?.data?.message || 'Failed to load nginx cluster info');
    } finally {
      setLoading(false);
    }
  };

  const handleClusterAction = async (action) => {
    if (!displayClusterId) return;
    setActionLoading((prev) => ({ ...prev, [action]: true }));
    try {
      switch (action) {
        case 'start':
          await nginxClusterAPI.start(displayClusterId);
          toast.success('Cluster started');
          break;
        case 'stop':
          await nginxClusterAPI.stop(displayClusterId);
          toast.success('Cluster stopped');
          break;
        case 'restart':
          await nginxClusterAPI.restart(displayClusterId);
          toast.success('Cluster restarted');
          break;
        case 'sync':
          await nginxClusterAPI.syncConfig(displayClusterId);
          toast.success('Configuration synced across nodes');
          break;
        case 'test':
          const resp = await nginxClusterAPI.testConnection(displayClusterId);
          const result = resp.data?.data || resp.data;
          setTestResult(result);
          if (result?.success) {
            toast.success(`Connection OK via ${result?.node_name || 'master'} (${result?.latency || 'n/a'})`);
          } else {
            toast.error(result?.message || 'Connection test failed');
          }
          break;
        default:
          break;
      }
      if (action !== 'test') {
        await loadAll();
      }
    } catch (error) {
      toast.error(error.response?.data?.message || `Failed to ${action} cluster`);
    } finally {
      setActionLoading((prev) => ({ ...prev, [action]: false }));
    }
  };

  const handleFailover = async () => {
    if (!displayClusterId || !failoverTarget) {
      toast.error('Please pick a target node');
      return;
    }
    setActionLoading((prev) => ({ ...prev, failover: true }));
    try {
      await nginxClusterAPI.triggerFailover(displayClusterId, failoverTarget, 'Manual failover from UI');
      toast.success('Failover triggered');
      setFailoverTarget('');
      await loadAll();
    } catch (error) {
      toast.error(error.response?.data?.message || 'Failover failed');
    } finally {
      setActionLoading((prev) => ({ ...prev, failover: false }));
    }
  };

  const copyToClipboard = (text, label) => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    toast.success(`${label || 'Value'} copied`);
  };

  if (!displayClusterId) {
    return <div className="nginx-cluster-panel empty">Missing cluster reference</div>;
  }

  if (loading) {
    return (
      <div className="nginx-cluster-panel loading-state">
        <Loader className="spin" size={32} />
        <p>Loading Nginx cluster...</p>
      </div>
    );
  }

  if (!info) {
    return (
      <div className="nginx-cluster-panel error-state">
        <AlertTriangle size={20} />
        <p>Unable to load Nginx cluster details.</p>
        <button className="btn btn-secondary btn-sm" onClick={loadAll}>
          <RefreshCw size={14} />
          Retry
        </button>
      </div>
    );
  }

  const statusClass = STATUS_COLORS[info.status?.toLowerCase?.()] || 'muted';
  const nodes = info.nodes || [];
  const endpoints = connectionInfo?.endpoints || info.endpoints || {};

  return (
    <div className="nginx-cluster-panel">
      <div className="section-header">
        <div>
          <h3>{info.cluster_name}</h3>
          <p>High availability traffic gateway with Keepalived + VRRP</p>
        </div>
        <div className="actions-group">
          <button
            className="btn btn-secondary btn-sm"
            onClick={() => handleClusterAction('start')}
            disabled={actionLoading.start}
          >
            {actionLoading.start ? <Loader size={14} className="spin" /> : <PlayIcon />}
            Start
          </button>
          <button
            className="btn btn-secondary btn-sm"
            onClick={() => handleClusterAction('stop')}
            disabled={actionLoading.stop}
          >
            {actionLoading.stop ? <Loader size={14} className="spin" /> : <SquareIcon />}
            Stop
          </button>
          <button
            className="btn btn-secondary btn-sm"
            onClick={() => handleClusterAction('restart')}
            disabled={actionLoading.restart}
          >
            {actionLoading.restart ? <Loader size={14} className="spin" /> : <RefreshCw size={14} />}
            Restart
          </button>
          <button
            className="btn btn-secondary btn-sm"
            onClick={() => handleClusterAction('sync')}
            disabled={actionLoading.sync}
          >
            {actionLoading.sync ? <Loader size={14} className="spin" /> : <Shield size={14} />}
            Sync Config
          </button>
          <button
            className="btn btn-primary btn-sm"
            onClick={() => handleClusterAction('test')}
            disabled={actionLoading.test}
          >
            {actionLoading.test ? <Loader size={14} className="spin" /> : <Activity size={14} />}
            Test Connection
          </button>
        </div>
      </div>

      <div className="cluster-summary">
        <SummaryCard
          icon={<CheckCircle size={18} />}
          label="Status"
          value={info.status || 'unknown'}
          accent={statusClass}
        />
        <SummaryCard
          icon={<Server size={18} />}
          label="Nodes"
          value={`${info.node_count || nodes.length} nodes`}
          subValue={`${nodes.filter((n) => n.is_healthy).length} healthy`}
        />
        <SummaryCard
          icon={<Globe size={18} />}
          label="Virtual IP"
          value={info.virtual_ip || endpoints?.virtual_ip?.ip || 'Not configured'}
          subValue={info.http_port ? `HTTP ${info.http_port}${info.https_port ? ` / HTTPS ${info.https_port}` : ''}` : ''}
        />
        <SummaryCard
          icon={<Signal size={18} />}
          label="Load Balancer"
          value={info.load_balance_mode || 'round_robin'}
          subValue={info.ssl_enabled ? 'TLS enabled' : 'TLS disabled'}
        />
      </div>

      {connectionInfo && (
        <div className="panel-section">
          <div className="section-heading">
            <h4>Connection Info</h4>
            <button className="btn-link" onClick={() => loadAll()}>
              <RefreshCw size={14} />
              Refresh
            </button>
          </div>
          <div className="connections-grid">
            {endpoints?.virtual_ip && (
              <EndpointCard
                title="Virtual IP"
                endpoint={endpoints.virtual_ip}
                onCopy={copyToClipboard}
              />
            )}
            {endpoints?.master_node && (
              <EndpointCard
                title="Master Node"
                endpoint={endpoints.master_node}
                onCopy={copyToClipboard}
              />
            )}
            {endpoints?.backup_nodes?.map((node) => (
              <EndpointCard key={node.node_id} title="Backup Node" endpoint={node} onCopy={copyToClipboard} />
            ))}
          </div>
        </div>
      )}

      {health && (
        <div className="panel-section health-section">
          <div className="section-heading">
            <h4>Health</h4>
            <span className={`badge ${STATUS_COLORS[health.status?.toLowerCase()] || 'muted'}`}>
              {health.status}
            </span>
          </div>
          <div className="health-grid">
            {health.node_health?.map((node) => (
              <div key={node.node_id} className={`health-card ${node.is_healthy ? 'healthy' : 'unhealthy'}`}>
                <div className="health-title">
                  <Server size={16} />
                  <span>{node.node_name}</span>
                  <span className={`role ${node.role}`}>{node.role}</span>
                </div>
                <div className="health-status">
                  <span>NGINX: {node.nginx_status}</span>
                  <span>Keepalived: {node.keepalived_status}</span>
                </div>
                <div className="health-meta">
                  <span>Last Check: {node.last_check}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="panel-section">
        <div className="section-heading">
          <h4>Nodes</h4>
          <button className="btn-link" onClick={loadAll}>
            <RefreshCw size={14} />
            Reload
          </button>
        </div>
        <div className="nodes-table">
          <div className="nodes-header">
            <span>Name</span>
            <span>Role</span>
            <span>Status</span>
            <span>IP / Port</span>
            <span>Priority</span>
          </div>
          {nodes.map((node) => (
            <div key={node.id} className="nodes-row">
              <span>{node.name}</span>
              <span className={`badge ${node.is_master ? 'primary' : 'muted'}`}>{node.role}</span>
              <span className={`status-dot ${node.status?.toLowerCase() || 'unknown'}`}>
                {node.status}
              </span>
              <span>
                {node.ip_address || 'n/a'}:{node.http_port}
              </span>
              <span>{node.priority || '-'}</span>
            </div>
          ))}
          {nodes.length === 0 && <div className="empty-row">No nodes reported</div>}
        </div>
      </div>

      <div className="panel-section">
        <div className="section-heading">
          <h4>Manual Failover</h4>
        </div>
        <div className="failover-controls">
          <select value={failoverTarget} onChange={(e) => setFailoverTarget(e.target.value)}>
            <option value="">Select replica</option>
            {nodes
              .filter((node) => !node.is_master)
              .map((node) => (
                <option value={node.id} key={node.id}>
                  {node.name} ({node.status})
                </option>
              ))}
          </select>
          <button
            className="btn btn-danger btn-sm"
            onClick={handleFailover}
            disabled={!failoverTarget || actionLoading.failover}
          >
            {actionLoading.failover ? <Loader size={14} className="spin" /> : <AlertTriangle size={14} />}
            Promote to Master
          </button>
        </div>
        {testResult && (
          <div className={`test-result ${testResult.success ? 'success' : 'danger'}`}>
            <strong>Test Result:</strong> {testResult.message || (testResult.success ? 'Healthy' : 'Failed')}
          </div>
        )}
      </div>
    </div>
  );
};

const SummaryCard = ({ icon, label, value, subValue, accent = 'muted' }) => (
  <div className={`summary-card ${accent}`}>
    <div className="summary-icon">{icon}</div>
    <div className="summary-content">
      <span className="summary-label">{label}</span>
      <span className="summary-value">{value}</span>
      {subValue && <span className="summary-sub">{subValue}</span>}
    </div>
  </div>
);

const EndpointCard = ({ title, endpoint, onCopy }) => {
  if (!endpoint) return null;
  return (
    <div className="endpoint-card">
      <div className="endpoint-header">
        <h5>{title}</h5>
        <span className={`badge ${endpoint.role || 'muted'}`}>{endpoint.role || 'active'}</span>
      </div>
      <div className="endpoint-body">
        <div className="endpoint-row">
          <span>HTTP</span>
          <button className="link-btn" onClick={() => onCopy(endpoint.http_url, 'HTTP URL')}>
            <Link2 size={14} />
            {endpoint.http_url || `${endpoint.ip}:${endpoint.http_port}`}
          </button>
        </div>
        {endpoint.https_url && (
          <div className="endpoint-row">
            <span>HTTPS</span>
            <button className="link-btn" onClick={() => onCopy(endpoint.https_url, 'HTTPS URL')}>
              <Link2 size={14} />
              {endpoint.https_url}
            </button>
          </div>
        )}
        <div className="endpoint-row">
          <span>Node</span>
          <span className="code">
            {endpoint.node_name || endpoint.ip}:{endpoint.http_port}
          </span>
        </div>
        <button className="btn btn-ghost btn-sm" onClick={() => onCopy(endpoint.ip, 'IP')}>
          <Copy size={14} /> Copy IP
        </button>
      </div>
    </div>
  );
};

const PlayIcon = () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M8 5v14l11-7z" fill="currentColor" /></svg>;
const SquareIcon = () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><rect x="6" y="6" width="12" height="12" fill="currentColor" /></svg>;

export default NginxClusterPanel;

