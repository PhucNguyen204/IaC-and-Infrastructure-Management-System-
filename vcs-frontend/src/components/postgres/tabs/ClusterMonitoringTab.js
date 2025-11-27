import React from 'react';
import MonitoringPanel from '../../monitoring/MonitoringPanel';

const ClusterMonitoringTab = ({ clusterId, clusterName }) => {
  return (
    <div style={{ maxWidth: '1400px' }}>
      <MonitoringPanel 
        instanceId={clusterId}
        instanceName={clusterName || 'PostgreSQL Cluster'}
        showLogs={true}
      />
    </div>
  );
};

export default ClusterMonitoringTab;
