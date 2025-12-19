import React, { useState, useEffect } from 'react';
import { 
  Link2, 
  Copy, 
  Check, 
  RefreshCw, 
  Server, 
  Database, 
  User, 
  Key,
  ExternalLink,
  Zap,
  Shield,
  AlertCircle
} from 'lucide-react';
import { clusterAPI } from '../../api';
import toast from 'react-hot-toast';
import './ConnectionPanel.css';

const ConnectionPanel = ({ clusterId, clusterInfo }) => {
  const [connectionInfo, setConnectionInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState(null);
  const [copiedField, setCopiedField] = useState(null);

  useEffect(() => {
    loadConnectionInfo();
  }, [clusterId]);

  const loadConnectionInfo = async () => {
    try {
      setLoading(true);
      // Use getById instead of getConnectionInfo since that endpoint doesn't exist
      const response = await clusterAPI.getById(clusterId);
      const clusterData = response.data?.data || response.data;
      
      // Transform cluster info to connection info format
      const transformedInfo = {
        endpoints: {
          write: {
            host: clusterData.write_endpoint?.host || 'localhost',
            port: clusterData.haproxy_port || clusterData.write_endpoint?.port || 5000,
            description: 'Primary (Read/Write via HAProxy)'
          },
          read: {
            host: clusterData.read_endpoints?.[0]?.host || 'localhost',
            port: (clusterData.haproxy_port || 5000) + 1,
            description: 'Replicas (Read-only via HAProxy)'
          }
        },
        credentials: {
          username: 'postgres',
          database: 'postgres',
          password: '********'
        },
        databases: ['postgres'],
        connectionString: `postgresql://postgres:****@localhost:${clusterData.haproxy_port || 5000}/postgres`
      };
      
      setConnectionInfo(transformedInfo);
    } catch (error) {
      console.error('Error loading connection info:', error);
      toast.error('Failed to load connection info');
    } finally {
      setLoading(false);
    }
  };

  const testConnection = async () => {
    try {
      setTesting(true);
      setTestResult(null);
      const response = await clusterAPI.testConnection(clusterId, {});
      setTestResult(response.data);
      if (response.data.success) {
        toast.success('Connection successful!');
      } else {
        toast.error('Connection failed');
      }
    } catch (error) {
      toast.error('Connection test failed');
      setTestResult({ success: false, message: error.message });
    } finally {
      setTesting(false);
    }
  };

  const copyToClipboard = async (text, fieldName) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(fieldName);
      toast.success('Copied to clipboard');
      setTimeout(() => setCopiedField(null), 2000);
    } catch (err) {
      toast.error('Failed to copy');
    }
  };

  if (loading) {
    return (
      <div className="connection-panel loading">
        <RefreshCw className="spin" size={24} />
        <span>Loading connection info...</span>
      </div>
    );
  }

  if (!connectionInfo) {
    return (
      <div className="connection-panel error">
        <AlertCircle size={24} />
        <span>Failed to load connection info</span>
        <button onClick={loadConnectionInfo}>Retry</button>
      </div>
    );
  }

  const { endpoints, credentials, databases } = connectionInfo;

  return (
    <div className="connection-panel">
      {/* Connection Test Section */}
      <div className="section test-section">
        <div className="section-header">
          <Zap size={20} />
          <h3>Connection Test</h3>
        </div>
        <div className="test-content">
          <button 
            className={`test-btn ${testing ? 'testing' : ''}`}
            onClick={testConnection}
            disabled={testing}
          >
            {testing ? (
              <>
                <RefreshCw className="spin" size={16} />
                Testing...
              </>
            ) : (
              <>
                <Zap size={16} />
                Test Connection
              </>
            )}
          </button>
          
          {testResult && (
            <div className={`test-result ${testResult.success ? 'success' : 'error'}`}>
              <div className="result-header">
                {testResult.success ? <Check size={16} /> : <AlertCircle size={16} />}
                <span>{testResult.message}</span>
              </div>
              {testResult.success && (
                <div className="result-details">
                  <div className="detail-item">
                    <span>Latency:</span>
                    <strong>{testResult.latency}</strong>
                  </div>
                  <div className="detail-item">
                    <span>Server:</span>
                    <strong>{testResult.server_version?.split(' ')[0]}</strong>
                  </div>
                  <div className="detail-item">
                    <span>Node:</span>
                    <strong>{testResult.node_role}</strong>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* HAProxy Endpoints */}
      <div className="section endpoints-section">
        <div className="section-header">
          <Shield size={20} />
          <h3>HAProxy Load Balancer</h3>
          <span className="badge recommended">Recommended</span>
        </div>
        <div className="endpoint-grid">
          <div className="endpoint-card primary">
            <div className="endpoint-label">
              <Server size={16} />
              Write Endpoint (Primary)
            </div>
            <div className="endpoint-value">
              <code>{endpoints.haproxy?.write_url || `postgresql://postgres@localhost:${endpoints.haproxy?.write_port}/postgres`}</code>
              <button 
                className="copy-btn"
                onClick={() => copyToClipboard(endpoints.haproxy?.write_url, 'write')}
              >
                {copiedField === 'write' ? <Check size={14} /> : <Copy size={14} />}
              </button>
            </div>
            <span className="endpoint-hint">Use for INSERT, UPDATE, DELETE operations</span>
          </div>

          <div className="endpoint-card secondary">
            <div className="endpoint-label">
              <Database size={16} />
              Read Endpoint (Replicas)
            </div>
            <div className="endpoint-value">
              <code>{endpoints.haproxy?.read_url || `postgresql://postgres@localhost:${endpoints.haproxy?.read_port}/postgres`}</code>
              <button 
                className="copy-btn"
                onClick={() => copyToClipboard(endpoints.haproxy?.read_url, 'read')}
              >
                {copiedField === 'read' ? <Check size={14} /> : <Copy size={14} />}
              </button>
            </div>
            <span className="endpoint-hint">Use for SELECT operations (load-balanced)</span>
          </div>

          <div className="endpoint-card stats">
            <div className="endpoint-label">
              <ExternalLink size={16} />
              HAProxy Stats Dashboard
            </div>
            <div className="endpoint-value">
              <code>{endpoints.haproxy?.stats_url || `http://localhost:${endpoints.haproxy?.stats_port}`}</code>
              <a 
                href={endpoints.haproxy?.stats_url}
                target="_blank"
                rel="noopener noreferrer"
                className="open-btn"
              >
                <ExternalLink size={14} />
              </a>
            </div>
          </div>
        </div>
      </div>

      {/* Direct Node Connections */}
      <div className="section nodes-section">
        <div className="section-header">
          <Server size={20} />
          <h3>Direct Node Connections</h3>
        </div>
        <div className="nodes-list">
          {endpoints.primary && (
            <div className="node-item primary">
              <div className="node-badge primary">PRIMARY</div>
              <div className="node-info">
                <span className="node-name">{endpoints.primary.node_name}</span>
                <code>{endpoints.primary.connection_url}</code>
              </div>
              <button 
                className="copy-btn"
                onClick={() => copyToClipboard(endpoints.primary.connection_url, 'primary-direct')}
              >
                {copiedField === 'primary-direct' ? <Check size={14} /> : <Copy size={14} />}
              </button>
            </div>
          )}
          
          {endpoints.replicas?.map((replica, idx) => (
            <div key={idx} className="node-item replica">
              <div className="node-badge replica">REPLICA</div>
              <div className="node-info">
                <span className="node-name">{replica.node_name}</span>
                <code>{replica.connection_url}</code>
              </div>
              <button 
                className="copy-btn"
                onClick={() => copyToClipboard(replica.connection_url, `replica-${idx}`)}
              >
                {copiedField === `replica-${idx}` ? <Check size={14} /> : <Copy size={14} />}
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Credentials */}
      <div className="section credentials-section">
        <div className="section-header">
          <Key size={20} />
          <h3>Credentials</h3>
        </div>
        <div className="credentials-grid">
          <div className="credential-item">
            <User size={16} />
            <span className="label">Username</span>
            <strong>{credentials.username}</strong>
            <button 
              className="copy-btn"
              onClick={() => copyToClipboard(credentials.username, 'username')}
            >
              {copiedField === 'username' ? <Check size={14} /> : <Copy size={14} />}
            </button>
          </div>
          <div className="credential-item">
            <Key size={16} />
            <span className="label">Password</span>
            <strong>{credentials.password_hint}</strong>
            <span className="hint">(Set during cluster creation)</span>
          </div>
          <div className="credential-item">
            <Database size={16} />
            <span className="label">Default Database</span>
            <strong>{credentials.database}</strong>
          </div>
        </div>
      </div>

      {/* Available Databases */}
      <div className="section databases-section">
        <div className="section-header">
          <Database size={20} />
          <h3>Available Databases</h3>
        </div>
        <div className="databases-list">
          {databases?.map((db, idx) => (
            <div key={idx} className="database-tag">
              <Database size={14} />
              {db}
            </div>
          ))}
        </div>
      </div>

      {/* Connection Examples */}
      <div className="section examples-section">
        <div className="section-header">
          <Link2 size={20} />
          <h3>Connection Examples</h3>
        </div>
        <div className="examples-tabs">
          <div className="example-item">
            <h4>psql</h4>
            <code>
              psql "postgresql://{credentials.username}:YOUR_PASSWORD@localhost:{endpoints.haproxy?.write_port}/{credentials.database}"
            </code>
          </div>
          <div className="example-item">
            <h4>Node.js (pg)</h4>
            <code>
{`const { Pool } = require('pg');
const pool = new Pool({
  host: 'localhost',
  port: ${endpoints.haproxy?.write_port},
  user: '${credentials.username}',
  password: 'YOUR_PASSWORD',
  database: '${credentials.database}'
});`}
            </code>
          </div>
          <div className="example-item">
            <h4>Python (psycopg2)</h4>
            <code>
{`import psycopg2
conn = psycopg2.connect(
    host="localhost",
    port=${endpoints.haproxy?.write_port},
    user="${credentials.username}",
    password="YOUR_PASSWORD",
    database="${credentials.database}"
)`}
            </code>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ConnectionPanel;

