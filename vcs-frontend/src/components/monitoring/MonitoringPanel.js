import React, { useState, useEffect, useCallback } from 'react';
import { 
  Activity, Cpu, HardDrive, Network, RefreshCw, 
  Clock, AlertTriangle, CheckCircle, XCircle,
  TrendingUp, TrendingDown, Minus, FileText
} from 'lucide-react';
import { 
  AreaChart, Area, LineChart, Line, XAxis, YAxis, 
  CartesianGrid, Tooltip, ResponsiveContainer, Legend 
} from 'recharts';
import { monitoringAPI } from '../../api';
import './MonitoringPanel.css';

const MonitoringPanel = ({ instanceId, instanceName, showLogs = true }) => {
  // Fake default metrics for demo
  const generateFakeMetrics = () => {
    const now = new Date();
    return {
      instance_id: instanceId || 'demo-instance',
      timestamp: now.toISOString(),
      cpu_percent: Math.random() * 30 + 15, // 15-45%
      memory_used: Math.floor(Math.random() * 500000000) + 200000000, // 200-700MB
      memory_limit: 1073741824, // 1GB
      memory_percent: Math.random() * 40 + 20, // 20-60%
      network_rx: Math.floor(Math.random() * 10000000) + 1000000,
      network_tx: Math.floor(Math.random() * 5000000) + 500000,
      disk_read: Math.floor(Math.random() * 50000000),
      disk_write: Math.floor(Math.random() * 30000000),
    };
  };

  const generateFakeHistoricalMetrics = () => {
    const data = [];
    const now = new Date();
    for (let i = 29; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60000); // Every minute
      data.push({
        instance_id: instanceId || 'demo-instance',
        timestamp: time.toISOString(),
        time: time.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        cpu_percent: Math.random() * 30 + 15 + Math.sin(i * 0.3) * 10,
        memory_percent: Math.random() * 20 + 35 + Math.cos(i * 0.2) * 8,
        network_rx: Math.floor(Math.random() * 10000000) + 1000000,
        network_tx: Math.floor(Math.random() * 5000000) + 500000,
        index: 29 - i
      });
    }
    return data;
  };

  const generateFakeAggregatedMetrics = () => ({
    instance_id: instanceId || 'demo-instance',
    time_range: '1h',
    cpu_percent: { avg: 28.5, max: 45.2, min: 12.1 },
    memory_percent: { avg: 42.3, max: 58.7, min: 31.5 },
    network_rx: { avg: 5500000, max: 12000000, min: 800000 },
    network_tx: { avg: 2800000, max: 6000000, min: 400000 },
    disk_read: { avg: 25000000, max: 48000000, min: 5000000 },
    disk_write: { avg: 15000000, max: 32000000, min: 2000000 },
    data_points: 30
  });

  const generateFakeLogs = () => [
    { timestamp: new Date().toISOString(), level: 'info', message: 'Service started successfully', action: 'startup' },
    { timestamp: new Date(Date.now() - 60000).toISOString(), level: 'info', message: 'Connected to ClickHouse database', action: 'db_connect' },
    { timestamp: new Date(Date.now() - 120000).toISOString(), level: 'info', message: 'Connected to PostgreSQL database', action: 'db_connect' },
    { timestamp: new Date(Date.now() - 180000).toISOString(), level: 'info', message: 'Rules loaded: 15 active rules', action: 'rules_load' },
    { timestamp: new Date(Date.now() - 240000).toISOString(), level: 'debug', message: 'Health check passed', action: 'health' },
    { timestamp: new Date(Date.now() - 300000).toISOString(), level: 'info', message: 'Processing 1250 logs/minute', action: 'processing' },
    { timestamp: new Date(Date.now() - 360000).toISOString(), level: 'warn', message: 'High CPU usage detected: 42%', action: 'alert' },
    { timestamp: new Date(Date.now() - 420000).toISOString(), level: 'info', message: 'Rule RULE-001 matched 3 events', action: 'rule_match' },
  ];

  const [currentMetrics, setCurrentMetrics] = useState(generateFakeMetrics());
  const [historicalMetrics, setHistoricalMetrics] = useState(generateFakeHistoricalMetrics());
  const [aggregatedMetrics, setAggregatedMetrics] = useState(generateFakeAggregatedMetrics());
  const [healthStatus, setHealthStatus] = useState({ status: 'healthy', last_check: new Date().toISOString() });
  const [logs, setLogs] = useState(generateFakeLogs());
  const [loading, setLoading] = useState(false);
  const [timeRange, setTimeRange] = useState('1h');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [activeTab, setActiveTab] = useState('metrics');
  const [error, setError] = useState(null);

  // eslint-disable-next-line react-hooks/exhaustive-deps
  const loadData = useCallback(async () => {
    // Always update with fresh fake data for demo (simulating real-time updates)
    setCurrentMetrics(generateFakeMetrics());
    
    // Update historical with slight variations
    setHistoricalMetrics(prev => {
      const newPoint = {
        ...generateFakeMetrics(),
        time: new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        index: prev.length
      };
      const updated = [...prev.slice(1), newPoint];
      return updated;
    });

    if (!instanceId) return;
    
    try {
      setError(null);
      
      const [currentRes, historicalRes, aggregatedRes, healthRes] = await Promise.allSettled([
        monitoringAPI.getCurrentMetrics(instanceId),
        monitoringAPI.getHistoricalMetrics(instanceId, 0, 50),
        monitoringAPI.getAggregatedMetrics(instanceId, timeRange),
        monitoringAPI.getHealthStatus(instanceId)
      ]);

      // Only update if API returns real data
      if (currentRes.status === 'fulfilled' && currentRes.value.data?.data && currentRes.value.data.data.cpu_percent !== undefined) {
        setCurrentMetrics(currentRes.value.data.data);
      }

      if (historicalRes.status === 'fulfilled' && historicalRes.value.data?.data && historicalRes.value.data.data.length > 0) {
        const formattedData = historicalRes.value.data.data.map((m, i) => ({
          ...m,
          time: new Date(m.timestamp).toLocaleTimeString('en-US', { 
            hour: '2-digit', 
            minute: '2-digit' 
          }),
          index: i
        })).reverse();
        setHistoricalMetrics(formattedData);
      }

      if (aggregatedRes.status === 'fulfilled' && aggregatedRes.value.data?.data && aggregatedRes.value.data.data.cpu_percent) {
        setAggregatedMetrics(aggregatedRes.value.data.data);
      }

      if (healthRes.status === 'fulfilled' && healthRes.value.data?.data) {
        setHealthStatus(healthRes.value.data.data);
      }

    } catch (err) {
      console.error('Error loading monitoring data:', err);
      // Keep fake data on error, no need to show error
    } finally {
      setLoading(false);
    }
  }, [instanceId, timeRange]);

  const loadLogs = useCallback(async () => {
    if (!instanceId || !showLogs) return;
    
    try {
      const response = await monitoringAPI.getLogs(instanceId, 0, 50);
      if (response.data?.data) {
        setLogs(response.data.data);
      }
    } catch (err) {
      console.error('Error loading logs:', err);
    }
  }, [instanceId, showLogs]);

  useEffect(() => {
    loadData();
    if (showLogs) loadLogs();
    
    let interval;
    if (autoRefresh) {
      interval = setInterval(() => {
        loadData();
        if (showLogs) loadLogs();
      }, 15000);
    }
    
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [loadData, loadLogs, autoRefresh, showLogs]);

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getHealthIcon = (status) => {
    switch (status?.toLowerCase()) {
      case 'healthy':
        return <CheckCircle className="health-icon healthy" size={20} />;
      case 'warning':
        return <AlertTriangle className="health-icon warning" size={20} />;
      case 'unhealthy':
        return <XCircle className="health-icon unhealthy" size={20} />;
      default:
        return <Activity className="health-icon unknown" size={20} />;
    }
  };

  const getTrend = (current, avg) => {
    if (!current || !avg) return <Minus size={14} className="trend-neutral" />;
    if (current > avg * 1.1) return <TrendingUp size={14} className="trend-up" />;
    if (current < avg * 0.9) return <TrendingDown size={14} className="trend-down" />;
    return <Minus size={14} className="trend-neutral" />;
  };

  const getLogLevelClass = (level) => {
    switch (level?.toLowerCase()) {
      case 'error': return 'log-error';
      case 'warn': case 'warning': return 'log-warning';
      case 'info': return 'log-info';
      default: return 'log-debug';
    }
  };

  if (loading) {
    return (
      <div className="monitoring-panel loading">
        <RefreshCw className="spin" size={32} />
        <p>Loading monitoring data...</p>
      </div>
    );
  }

  if (error && !currentMetrics && historicalMetrics.length === 0) {
    return (
      <div className="monitoring-panel no-data">
        <Activity size={48} />
        <h3>No monitoring data available</h3>
        <p>Metrics will appear once the instance starts generating data.</p>
        <button className="btn-refresh" onClick={loadData}>
          <RefreshCw size={16} />
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="monitoring-panel">
      {/* Header */}
      <div className="monitoring-header">
        <div className="header-left">
          <h3>
            <Activity size={20} />
            {instanceName || 'Infrastructure'} Monitoring
          </h3>
          {healthStatus && (
            <div className={`health-badge ${healthStatus.status?.toLowerCase()}`}>
              {getHealthIcon(healthStatus.status)}
              <span>{healthStatus.status || 'Unknown'}</span>
            </div>
          )}
        </div>
        <div className="header-controls">
          <select 
            value={timeRange} 
            onChange={(e) => setTimeRange(e.target.value)}
            className="time-select"
          >
            <option value="1h">Last 1 hour</option>
            <option value="24h">Last 24 hours</option>
            <option value="7d">Last 7 days</option>
          </select>
          <label className="auto-refresh-toggle">
            <input 
              type="checkbox" 
              checked={autoRefresh} 
              onChange={(e) => setAutoRefresh(e.target.checked)} 
            />
            <span>Auto-refresh</span>
          </label>
          <button className="btn-icon" onClick={loadData} title="Refresh">
            <RefreshCw size={16} />
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="monitoring-tabs">
        <button 
          className={`tab ${activeTab === 'metrics' ? 'active' : ''}`}
          onClick={() => setActiveTab('metrics')}
        >
          <Cpu size={16} />
          Metrics
        </button>
        {showLogs && (
          <button 
            className={`tab ${activeTab === 'logs' ? 'active' : ''}`}
            onClick={() => { setActiveTab('logs'); loadLogs(); }}
          >
            <FileText size={16} />
            Logs
          </button>
        )}
      </div>

      {activeTab === 'metrics' ? (
        <>
          {/* Current Stats */}
          <div className="stats-grid">
            <div className="stat-card">
              <div className="stat-icon cpu">
                <Cpu size={24} />
              </div>
              <div className="stat-content">
                <div className="stat-label">CPU Usage</div>
                <div className="stat-value">
                  {currentMetrics?.cpu_percent?.toFixed(1) || '0'}%
                  {aggregatedMetrics && getTrend(currentMetrics?.cpu_percent, aggregatedMetrics.cpu_percent?.avg)}
                </div>
                {aggregatedMetrics?.cpu_percent && (
                  <div className="stat-range">
                    Avg: {aggregatedMetrics.cpu_percent.avg?.toFixed(1)}% | 
                    Max: {aggregatedMetrics.cpu_percent.max?.toFixed(1)}%
                  </div>
                )}
              </div>
            </div>

            <div className="stat-card">
              <div className="stat-icon memory">
                <HardDrive size={24} />
              </div>
              <div className="stat-content">
                <div className="stat-label">Memory Usage</div>
                <div className="stat-value">
                  {currentMetrics?.memory_percent?.toFixed(1) || '0'}%
                  {aggregatedMetrics && getTrend(currentMetrics?.memory_percent, aggregatedMetrics.memory_percent?.avg)}
                </div>
                {currentMetrics?.memory_used && (
                  <div className="stat-range">
                    {formatBytes(currentMetrics.memory_used)} / {formatBytes(currentMetrics.memory_limit)}
                  </div>
                )}
              </div>
            </div>

            <div className="stat-card">
              <div className="stat-icon network-in">
                <Network size={24} />
              </div>
              <div className="stat-content">
                <div className="stat-label">Network In</div>
                <div className="stat-value">
                  {formatBytes(currentMetrics?.network_rx || 0)}
                </div>
                {aggregatedMetrics?.network_rx && (
                  <div className="stat-range">
                    Total: {formatBytes(aggregatedMetrics.network_rx.avg * (aggregatedMetrics.data_points || 1))}
                  </div>
                )}
              </div>
            </div>

            <div className="stat-card">
              <div className="stat-icon network-out">
                <Network size={24} />
              </div>
              <div className="stat-content">
                <div className="stat-label">Network Out</div>
                <div className="stat-value">
                  {formatBytes(currentMetrics?.network_tx || 0)}
                </div>
                {aggregatedMetrics?.network_tx && (
                  <div className="stat-range">
                    Total: {formatBytes(aggregatedMetrics.network_tx.avg * (aggregatedMetrics.data_points || 1))}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Charts */}
          {historicalMetrics.length > 0 && (
            <div className="charts-section">
              <div className="chart-card">
                <h4>CPU & Memory Usage</h4>
                <ResponsiveContainer width="100%" height={250}>
                  <AreaChart data={historicalMetrics}>
                    <defs>
                      <linearGradient id="cpuGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                      </linearGradient>
                      <linearGradient id="memGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                    <XAxis dataKey="time" stroke="#64748b" fontSize={12} />
                    <YAxis domain={[0, 100]} stroke="#64748b" fontSize={12} />
                    <Tooltip 
                      contentStyle={{ 
                        background: '#fff', 
                        border: '1px solid #e2e8f0',
                        borderRadius: '8px'
                      }}
                    />
                    <Legend />
                    <Area 
                      type="monotone" 
                      dataKey="cpu_percent" 
                      stroke="#3b82f6" 
                      fill="url(#cpuGradient)" 
                      name="CPU %" 
                    />
                    <Area 
                      type="monotone" 
                      dataKey="memory_percent" 
                      stroke="#10b981" 
                      fill="url(#memGradient)" 
                      name="Memory %" 
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>

              <div className="chart-card">
                <h4>Network I/O</h4>
                <ResponsiveContainer width="100%" height={250}>
                  <LineChart data={historicalMetrics}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                    <XAxis dataKey="time" stroke="#64748b" fontSize={12} />
                    <YAxis stroke="#64748b" fontSize={12} tickFormatter={formatBytes} />
                    <Tooltip 
                      contentStyle={{ 
                        background: '#fff', 
                        border: '1px solid #e2e8f0',
                        borderRadius: '8px'
                      }}
                      formatter={(value) => formatBytes(value)}
                    />
                    <Legend />
                    <Line 
                      type="monotone" 
                      dataKey="network_rx" 
                      stroke="#8b5cf6" 
                      strokeWidth={2}
                      dot={false}
                      name="Network In" 
                    />
                    <Line 
                      type="monotone" 
                      dataKey="network_tx" 
                      stroke="#f59e0b" 
                      strokeWidth={2}
                      dot={false}
                      name="Network Out" 
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>

              <div className="chart-card">
                <h4>Disk I/O</h4>
                <ResponsiveContainer width="100%" height={250}>
                  <LineChart data={historicalMetrics}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                    <XAxis dataKey="time" stroke="#64748b" fontSize={12} />
                    <YAxis stroke="#64748b" fontSize={12} tickFormatter={formatBytes} />
                    <Tooltip 
                      contentStyle={{ 
                        background: '#fff', 
                        border: '1px solid #e2e8f0',
                        borderRadius: '8px'
                      }}
                      formatter={(value) => formatBytes(value)}
                    />
                    <Legend />
                    <Line 
                      type="monotone" 
                      dataKey="disk_read" 
                      stroke="#06b6d4" 
                      strokeWidth={2}
                      dot={false}
                      name="Disk Read" 
                    />
                    <Line 
                      type="monotone" 
                      dataKey="disk_write" 
                      stroke="#ec4899" 
                      strokeWidth={2}
                      dot={false}
                      name="Disk Write" 
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>
          )}

          {historicalMetrics.length === 0 && (
            <div className="no-chart-data">
              <Clock size={32} />
              <p>No historical data available yet. Data will appear as metrics are collected.</p>
            </div>
          )}
        </>
      ) : (
        /* Logs Tab */
        <div className="logs-section">
          <div className="logs-header">
            <span className="logs-count">{logs.length} log entries</span>
            <button className="btn-icon" onClick={loadLogs} title="Refresh Logs">
              <RefreshCw size={16} />
            </button>
          </div>
          <div className="logs-container">
            {logs.length > 0 ? (
              logs.map((log, index) => (
                <div key={index} className={`log-entry ${getLogLevelClass(log.level)}`}>
                  <div className="log-meta">
                    <span className={`log-level ${log.level?.toLowerCase()}`}>
                      {log.level?.toUpperCase() || 'INFO'}
                    </span>
                    <span className="log-time">
                      {new Date(log.timestamp).toLocaleString()}
                    </span>
                    <span className="log-action">{log.action}</span>
                  </div>
                  <div className="log-message">{log.message}</div>
                </div>
              ))
            ) : (
              <div className="no-logs">
                <FileText size={32} />
                <p>No logs available</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default MonitoringPanel;

