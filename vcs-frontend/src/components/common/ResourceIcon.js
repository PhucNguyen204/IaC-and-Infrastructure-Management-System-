import React from 'react';
import { Database, Globe, Layers, Server, Box } from 'lucide-react';

const ResourceIcon = ({ type, size = 20 }) => {
  const getIcon = () => {
    switch (type?.toLowerCase()) {
      case 'postgres_cluster':
      case 'postgres_instance':
      case 'postgres_database':
      case 'postgresql':
      case 'database':
        return <Database size={size} />;
      case 'nginx_gateway':
      case 'nginx_cluster':
      case 'nginx':
      case 'gateway':
        return <Globe size={size} />;
      case 'dind_environment':
      case 'docker_sandbox':
      case 'sandbox':
        return <Box size={size} />;
      case 'stack':
        return <Layers size={size} />;
      case 'server':
        return <Server size={size} />;
      default:
        return <Box size={size} />;
    }
  };

  return <span className="resource-icon">{getIcon()}</span>;
};

export default ResourceIcon;
