import axios from 'axios';

const AUTH_URL = 'http://localhost:8082';
const PROVISIONING_URL = 'http://localhost:8083/api/v1';
const MONITORING_URL = 'http://localhost:8084/api/v1';

// Auth API
export const authAPI = {
  login: (username, password) => 
    axios.post(`${AUTH_URL}/auth/login`, { username, password }),
  
  refreshToken: (refreshToken) => 
    axios.post(`${AUTH_URL}/auth/refresh`, { refresh_token: refreshToken }),
};

// Create axios instance with auth token
const createAuthAxios = () => {
  const token = localStorage.getItem('access_token');
  const instance = axios.create({
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  // Add response interceptor to handle 401 errors
  instance.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error.response?.status === 401) {
        // Token expired or invalid, clear storage and redirect to login
        localStorage.removeItem('access_token');
        window.location.reload(); // Force reload to show login page
      }
      return Promise.reject(error);
    }
  );

  return instance;
};

// PostgreSQL Cluster API
export const clusterAPI = {
  create: (data) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster`, data),
  
  // NOTE: Backend không có endpoint GET /postgres/cluster để list all clusters
  // Để lấy danh sách clusters, sử dụng stackAPI.getAll() và filter resources có type 'POSTGRES_CLUSTER'
  // getAll: () => 
  //   createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster`),
  
  getById: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${id}`),
  
  delete: (id) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/postgres/cluster/${id}`),
  
  start: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/start`),
  
  stop: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/stop`),
  
  restart: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/restart`),
  
  scale: (id, nodeCount) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/scale`, { node_count: nodeCount }),
  
  // Backup and Restore
  backup: (id, backupData) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/backup`, backupData),
  
  listBackups: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${id}/backups`),
  
  restore: (id, restoreData) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/restore`, restoreData),
  
  // Patroni Management
  patroniSwitchover: (id, switchoverData) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/patroni/switchover`, switchoverData),
  
  patroniReinit: (id, reinitData) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/patroni/reinit`, reinitData),
  
  patroniPause: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/patroni/pause`),
  
  patroniResume: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/patroni/resume`),
  
  patroniStatus: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${id}/patroni/status`),
  
  failover: (id, newPrimaryNodeId) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/failover`, { new_primary_node_id: newPrimaryNodeId }),
  
  // Node management
  addNode: (id, nodeName) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/nodes`, { node_name: nodeName }),
  
  removeNode: (id, nodeId) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/postgres/cluster/${id}/nodes`, { data: { node_id: nodeId } }),
  
  stopNode: (id, nodeId) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/nodes/stop`, { node_id: nodeId }),
  
  startNode: (id, nodeId) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/nodes/start`, { node_id: nodeId }),
  
  getFailoverHistory: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${id}/failover-history`),
  
  getReplication: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${id}/replication`),
  
  getStats: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${id}/stats`),
  
  getLogs: (id, lines = 100) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${id}/logs?lines=${lines}`),
  
  getEndpoints: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${id}/endpoints`),
  
  updateConfig: (id, config) => 
    createAuthAxios().put(`${PROVISIONING_URL}/postgres/cluster/${id}/config`, config),
  
  // User management
  createUser: (id, userData) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/users`, userData),
  
  listUsers: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${id}/users`),
  
  deleteUser: (id, username) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/postgres/cluster/${id}/users/${username}`),
  
  // Database management
  createDatabase: (id, dbData) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${id}/databases`, dbData),
  
  listDatabases: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${id}/databases`),
  
  deleteDatabase: (id, dbname) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/postgres/cluster/${id}/databases/${dbname}`),
  
  // Direct PostgreSQL queries (connect to container directly)
  getTables: (clusterId, database) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${clusterId}/databases/${database}/tables`),
  
  getTableSchema: (clusterId, database, table) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${clusterId}/databases/${database}/tables/${table}/schema`),
  
  getTableData: (clusterId, database, table, page = 1, limit = 50) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${clusterId}/databases/${database}/tables/${table}/data?page=${page}&limit=${limit}`),
  
  executeQuery: (clusterId, query, database = 'postgres', nodeId = null) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${clusterId}/query`, { 
      query, 
      database,
      node_id: nodeId 
    }),
  
  testReplication: (clusterId) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${clusterId}/test-replication`),

  // Connection Management
  testConnection: (clusterId, data = {}) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/cluster/${clusterId}/test-connection`, data),

  getConnectionInfo: (clusterId) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/cluster/${clusterId}/connection-info`),
};

// PostgreSQL Single Instance API
export const postgresAPI = {
  create: (data) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/single`, data),
  
  getById: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/single/${id}`),
  
  delete: (id) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/postgres/single/${id}`),
  
  start: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/single/${id}/start`),
  
  stop: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/single/${id}/stop`),
  
  restart: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/postgres/single/${id}/restart`),
  
  getLogs: (id, lines = 100) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/single/${id}/logs?tail=${lines}`),
  
  getStats: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/postgres/single/${id}/stats`),
};

// Nginx API
export const nginxAPI = {
  create: (data) => 
    createAuthAxios().post(`${PROVISIONING_URL}/nginx`, data),
  
  // NOTE: Backend không có endpoint GET /nginx để list all nginx instances
  // Để lấy danh sách nginx, sử dụng stackAPI.getAll() và filter resources có type 'NGINX'
  // getAll: () => 
  //   createAuthAxios().get(`${PROVISIONING_URL}/nginx`),
  
  getById: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/${id}`),
  
  delete: (id) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/nginx/${id}`),
  
  start: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/${id}/start`),
  
  stop: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/${id}/stop`),
  
  restart: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/${id}/restart`),
  
  reload: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/${id}/reload`),
  
  // Config management
  updateConfig: (id, config) => 
    createAuthAxios().put(`${PROVISIONING_URL}/nginx/${id}/config`, config),
  
  getConfig: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/${id}/config`),
  
  // Domain management
  addDomain: (id, domain) => 
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/${id}/domains`, domain),
  
  deleteDomain: (id, domain) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/nginx/${id}/domains/${domain}`),
  
  // Route management
  addRoute: (id, route) => 
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/${id}/routes`, route),
  
  deleteRoute: (id, routeId) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/nginx/${id}/routes/${routeId}`),
  
  // Upstream management
  getUpstreams: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/${id}/upstreams`),
  
  addUpstream: (id, upstream) => 
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/${id}/upstreams`, upstream),
  
  updateUpstream: (id, upstreamName, upstream) => 
    createAuthAxios().put(`${PROVISIONING_URL}/nginx/${id}/upstreams/${upstreamName}`, upstream),
  
  deleteUpstream: (id, upstreamName) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/nginx/${id}/upstreams/${upstreamName}`),
  
  // Security settings
  setSecurityPolicy: (id, policy) => 
    createAuthAxios().put(`${PROVISIONING_URL}/nginx/${id}/security`, policy),
  
  getSecurityPolicy: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/${id}/security`),
  
  // SSL management
  enableSSL: (id, sslConfig) => 
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/${id}/ssl`, sslConfig),
  
  disableSSL: (id) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/nginx/${id}/ssl`),
  
  // Logs and stats
  getLogs: (id, lines = 100) => 
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/${id}/logs?lines=${lines}`),
  
  getStats: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/${id}/stats`),
};

// Nginx Cluster API
// NOTE: Backend không có endpoint GET /nginx/cluster để list all clusters
// Để lấy danh sách clusters, sử dụng stackAPI.getAll() và filter resources có type 'NGINX_CLUSTER'
// Xem ví dụ trong NginxClusterDashboard.js
export const nginxClusterAPI = {
  create: (data) =>
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/cluster`, data),

  getById: (id) =>
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/cluster/${id}`),

  delete: (id) =>
    createAuthAxios().delete(`${PROVISIONING_URL}/nginx/cluster/${id}`),

  start: (id) =>
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/cluster/${id}/start`),

  stop: (id) =>
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/cluster/${id}/stop`),

  restart: (id) =>
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/cluster/${id}/restart`),

  addNode: (id, nodeData) =>
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/cluster/${id}/nodes`, nodeData),

  getNode: (id, nodeId) =>
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/cluster/${id}/nodes/${nodeId}`),

  removeNode: (id, nodeId) =>
    createAuthAxios().delete(`${PROVISIONING_URL}/nginx/cluster/${id}/nodes/${nodeId}`),

  updateConfig: (id, config) =>
    createAuthAxios().put(`${PROVISIONING_URL}/nginx/cluster/${id}/config`, config),

  syncConfig: (id) =>
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/cluster/${id}/sync-config`),

  listUpstreams: (id) =>
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/cluster/${id}/upstreams`),

  addUpstream: (id, upstream) =>
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/cluster/${id}/upstreams`, upstream),

  updateUpstream: (id, upstreamId, upstream) =>
    createAuthAxios().put(`${PROVISIONING_URL}/nginx/cluster/${id}/upstreams/${upstreamId}`, upstream),

  deleteUpstream: (id, upstreamId) =>
    createAuthAxios().delete(`${PROVISIONING_URL}/nginx/cluster/${id}/upstreams/${upstreamId}`),

  listServerBlocks: (id) =>
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/cluster/${id}/server-blocks`),

  addServerBlock: (id, block) =>
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/cluster/${id}/server-blocks`, block),

  deleteServerBlock: (id, blockId) =>
    createAuthAxios().delete(`${PROVISIONING_URL}/nginx/cluster/${id}/server-blocks/${blockId}`),

  getHealth: (id) =>
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/cluster/${id}/health`),

  getMetrics: (id) =>
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/cluster/${id}/metrics`),

  testConnection: (id) =>
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/cluster/${id}/test-connection`),

  getConnectionInfo: (id) =>
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/cluster/${id}/connection-info`),

  triggerFailover: (id, targetNodeId, reason = '') =>
    createAuthAxios().post(`${PROVISIONING_URL}/nginx/cluster/${id}/failover`, {
      target_node_id: targetNodeId,
      reason,
    }),

  getFailoverHistory: (id) =>
    createAuthAxios().get(`${PROVISIONING_URL}/nginx/cluster/${id}/failover-history`),
};

// Docker-in-Docker (DinD) API - Isolated Docker Sandbox
export const dinDAPI = {
  // Environment Management
  createEnvironment: (data) => 
    createAuthAxios().post(`${PROVISIONING_URL}/dind/environments`, data),
  
  listEnvironments: () => 
    createAuthAxios().get(`${PROVISIONING_URL}/dind/environments`),
  
  getEnvironment: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/dind/environments/${id}`),
  
  deleteEnvironment: (id) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/dind/environments/${id}`),
  
  startEnvironment: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/dind/environments/${id}/start`),
  
  stopEnvironment: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/dind/environments/${id}/stop`),
  
  // Docker Operations inside DinD
  execCommand: (id, command, timeout = 60) => 
    createAuthAxios().post(`${PROVISIONING_URL}/dind/environments/${id}/exec`, { command, timeout }),
  
  buildImage: (id, dockerfile, imageName, tag = 'latest', noCache = false) => 
    createAuthAxios().post(`${PROVISIONING_URL}/dind/environments/${id}/build`, { 
      dockerfile, 
      image_name: imageName, 
      tag,
      no_cache: noCache 
    }),
  
  runCompose: (id, composeContent, action, serviceName = '', detach = true) => 
    createAuthAxios().post(`${PROVISIONING_URL}/dind/environments/${id}/compose`, { 
      compose_content: composeContent, 
      action, 
      service_name: serviceName,
      detach 
    }),
  
  pullImage: (id, image) => 
    createAuthAxios().post(`${PROVISIONING_URL}/dind/environments/${id}/pull`, { image }),
  
  // Info Retrieval
  listContainers: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/dind/environments/${id}/containers`),
  
  listImages: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/dind/environments/${id}/images`),
  
  getLogs: (id, tail = 100) => 
    createAuthAxios().get(`${PROVISIONING_URL}/dind/environments/${id}/logs?tail=${tail}`),
  
  getStats: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/dind/environments/${id}/stats`),
};

// Stack API
export const stackAPI = {
  getAll: () => 
    createAuthAxios().get(`${PROVISIONING_URL}/stacks`),
  
  getById: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/stacks/${id}`),
  
  create: (data) => 
    createAuthAxios().post(`${PROVISIONING_URL}/stacks`, data),
  
  delete: (id) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/stacks/${id}`),
  
  start: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/stacks/${id}/start`),
  
  stop: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/stacks/${id}/stop`),
  
  restart: (id) => 
    createAuthAxios().post(`${PROVISIONING_URL}/stacks/${id}/restart`),
  
  clone: (id, newName) => 
    createAuthAxios().post(`${PROVISIONING_URL}/stacks/clone`, { stack_id: id, name: newName }),
  
  exportTemplate: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/stacks/${id}/export`),
  
  getLogs: (id, lines = 100) => 
    createAuthAxios().get(`${PROVISIONING_URL}/stacks/${id}/logs?lines=${lines}`),
  
  getMetrics: (id, timeRange = '24h') => 
    createAuthAxios().get(`${PROVISIONING_URL}/stacks/${id}/metrics?range=${timeRange}`),
};

// Auto Deploy API - Deploy containers with auto-created infrastructure
export const deployAPI = {
  // Deploy a container with auto-provisioned infrastructure
  deploy: (data) => 
    createAuthAxios().post(`${PROVISIONING_URL}/deploy`, data),
  
  // List all deployments
  list: () => 
    createAuthAxios().get(`${PROVISIONING_URL}/deploy`),
  
  // Get deployment by ID
  getById: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/deploy/${id}`),
  
  // Delete deployment and its infrastructure
  delete: (id) => 
    createAuthAxios().delete(`${PROVISIONING_URL}/deploy/${id}`),
  
  // Get deployment logs
  getLogs: (id, lines = 100) => 
    createAuthAxios().get(`${PROVISIONING_URL}/deploy/${id}/logs?lines=${lines}`),
  
  // Health check for deployment
  getHealth: (id) => 
    createAuthAxios().get(`${PROVISIONING_URL}/deploy/${id}/health`),
};

// Monitoring API (Elasticsearch-based metrics & logs)
export const monitoringAPI = {
  // Get current metrics for an instance
  getCurrentMetrics: (instanceId) => 
    createAuthAxios().get(`${MONITORING_URL}/monitoring/metrics/${instanceId}`),
  
  // Get historical metrics with pagination
  getHistoricalMetrics: (instanceId, from = 0, size = 100) => 
    createAuthAxios().get(`${MONITORING_URL}/monitoring/metrics/${instanceId}/history?from=${from}&size=${size}`),
  
  // Get aggregated metrics (avg/min/max) for time range
  getAggregatedMetrics: (instanceId, timeRange = '1h') => 
    createAuthAxios().get(`${MONITORING_URL}/monitoring/metrics/${instanceId}/aggregate?range=${timeRange}`),
  
  // Get logs for an instance
  getLogs: (instanceId, from = 0, size = 100) => 
    createAuthAxios().get(`${MONITORING_URL}/monitoring/logs/${instanceId}?from=${from}&size=${size}`),
  
  // Get health status from Redis
  getHealthStatus: (instanceId) => 
    createAuthAxios().get(`${MONITORING_URL}/monitoring/health/${instanceId}`),
  
  // List all monitored infrastructure
  listInfrastructure: () => 
    createAuthAxios().get(`${MONITORING_URL}/monitoring/infrastructure`),

  // Get uptime status
  getUptime: (infraId, period = '24h') => 
    createAuthAxios().get(`${MONITORING_URL}/uptime/${infraId}?period=${period}`),
};
