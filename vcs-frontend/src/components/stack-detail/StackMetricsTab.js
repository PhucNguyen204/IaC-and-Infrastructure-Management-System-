import React, { useState, useEffect } from 'react';
import { Activity, Database, Globe, Container, RefreshCw } from 'lucide-react';
import { LineChart, Line, AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { stackAPI, monitoringAPI } from '../../api';
import MonitoringPanel from '../monitoring/MonitoringPanel';
import './StackMetricsTab.css';

const StackMetricsTab = ({ stackId, resources = [] }) => {
  const [selectedResource, setSelectedResource] = useState(null);
  const [infrastructureList, setInfrastructureList] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadInfrastructure();
  }, []);

  const loadInfrastructure = async () => {
    try {
      const response = await monitoringAPI.listInfrastructure();
      if (response.data?.data) {
        setInfrastructureList(response.data.data);
      }
    } catch (error) {
      console.error('Error loading infrastructure:', error);
    } finally {
      setLoading(false);
    }
  };

  const getResourceIcon = (type) => {
    switch (type?.toUpperCase()) {
      case 'POSTGRES_CLUSTER':
        return <Database size={18} className="resource-icon postgres" />;
      case 'NGINX_GATEWAY':
        return <Globe size={18} className="resource-icon nginx" />;
      case 'DOCKER_SERVICE':
        return <Container size={18} className="resource-icon docker" />;
      default:
        return <Activity size={18} className="resource-icon" />;
    }
  };

  // Combine resources from stack with infrastructure list
  const monitoredResources = resources.map(r => ({
    ...r,
    monitored: infrastructureList.some(i => i.id === r.resource_id || i.id === r.infrastructure_id)
  }));

  if (loading) {
    return (
      <div className="stack-metrics-tab loading">
        <RefreshCw className="spin" size={32} />
        <p>Loading monitoring data...</p>
      </div>
    );
  }

  return (
    <div className="stack-metrics-tab">
      <div className="metrics-header">
        <h3>
          <Activity size={20} />
          Stack Monitoring
        </h3>
        <p className="metrics-subtitle">
          Real-time metrics and logs from Elasticsearch
        </p>
      </div>

      {/* Resource Selector */}
      <div className="resource-selector">
        <h4>Select Resource to Monitor</h4>
        <div className="resource-list">
          {monitoredResources.length > 0 ? (
            monitoredResources.map((resource, index) => (
              <button
                key={index}
                className={`resource-btn ${selectedResource?.resource_id === resource.resource_id ? 'active' : ''}`}
                onClick={() => setSelectedResource(resource)}
              >
                {getResourceIcon(resource.resource_type)}
                <div className="resource-info">
                  <span className="resource-name">{resource.resource_name || resource.resource_type}</span>
                  <span className="resource-type">{resource.resource_type?.replace(/_/g, ' ')}</span>
                </div>
                {resource.monitored && (
                  <span className="monitored-badge">Live</span>
                )}
              </button>
            ))
          ) : (
            <div className="no-resources">
              <p>No resources in this stack</p>
            </div>
          )}

          {/* Show monitored infrastructure that might not be in stack */}
          {infrastructureList.length > 0 && (
            <>
              <div className="resource-divider">
                <span>All Monitored Infrastructure</span>
              </div>
              {infrastructureList.map((infra, index) => (
                <button
                  key={`infra-${index}`}
                  className={`resource-btn ${selectedResource?.id === infra.id ? 'active' : ''}`}
                  onClick={() => setSelectedResource({ 
                    resource_id: infra.id, 
                    resource_name: infra.name,
                    resource_type: infra.type 
                  })}
                >
                  <Activity size={18} className="resource-icon" />
                  <div className="resource-info">
                    <span className="resource-name">{infra.name}</span>
                    <span className={`resource-status ${infra.status}`}>{infra.status}</span>
                  </div>
                </button>
              ))}
            </>
          )}
        </div>
      </div>

      {/* Monitoring Panel for Selected Resource */}
      {selectedResource ? (
        <MonitoringPanel 
          instanceId={selectedResource.resource_id || selectedResource.infrastructure_id}
          instanceName={selectedResource.resource_name}
          showLogs={true}
        />
      ) : (
        <div className="select-resource-prompt">
          <Activity size={48} />
          <h3>Select a resource to view monitoring data</h3>
          <p>Choose from the resources above to see real-time metrics, historical data, and logs.</p>
        </div>
      )}
    </div>
  );
};

export default StackMetricsTab;
