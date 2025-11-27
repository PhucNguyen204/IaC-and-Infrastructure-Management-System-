import React, { useState } from 'react';
import { dinDAPI } from '../../api';
import { 
  Layers, 
  Play, 
  Square,
  RefreshCw,
  FileText,
  Copy,
  CheckCircle,
  XCircle,
  AlertCircle,
  List
} from 'lucide-react';
import './DinDCompose.css';

const DinDCompose = ({ environmentId, environmentStatus, onRefresh }) => {
  const [composeContent, setComposeContent] = useState(`version: '3.8'

services:
  web:
    image: nginx:alpine
    ports:
      - "8080:80"
    
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"`);
  const [executing, setExecuting] = useState(false);
  const [result, setResult] = useState(null);
  const [activeAction, setActiveAction] = useState(null);
  const [activeTemplate, setActiveTemplate] = useState(null);

  const templates = [
    {
      name: 'Web + Redis',
      content: `version: '3.8'

services:
  web:
    image: nginx:alpine
    ports:
      - "8080:80"
    
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"`
    },
    {
      name: 'LAMP Stack',
      content: `version: '3.8'

services:
  web:
    image: php:8.2-apache
    ports:
      - "80:80"
    volumes:
      - ./html:/var/www/html
    depends_on:
      - db

  db:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: app
    ports:
      - "3306:3306"`
    },
    {
      name: 'Node + MongoDB',
      content: `version: '3.8'

services:
  app:
    image: node:18-alpine
    working_dir: /app
    command: sh -c "npm install && npm start"
    ports:
      - "3000:3000"
    depends_on:
      - mongo
    environment:
      MONGO_URL: mongodb://mongo:27017/app

  mongo:
    image: mongo:6
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db

volumes:
  mongo_data:`
    },
    {
      name: 'WordPress',
      content: `version: '3.8'

services:
  wordpress:
    image: wordpress:latest
    ports:
      - "8080:80"
    environment:
      WORDPRESS_DB_HOST: db
      WORDPRESS_DB_USER: wordpress
      WORDPRESS_DB_PASSWORD: wordpress
      WORDPRESS_DB_NAME: wordpress
    depends_on:
      - db

  db:
    image: mysql:5.7
    environment:
      MYSQL_DATABASE: wordpress
      MYSQL_USER: wordpress
      MYSQL_PASSWORD: wordpress
      MYSQL_ROOT_PASSWORD: secret
    volumes:
      - db_data:/var/lib/mysql

volumes:
  db_data:`
    },
    {
      name: 'ELK Stack',
      content: `version: '3.8'

services:
  elasticsearch:
    image: elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"

  kibana:
    image: kibana:8.11.0
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch

  logstash:
    image: logstash:8.11.0
    depends_on:
      - elasticsearch`
    }
  ];

  const actions = [
    { id: 'up', label: 'Up (Start)', icon: Play, color: 'success' },
    { id: 'down', label: 'Down (Stop)', icon: Square, color: 'warning' },
    { id: 'restart', label: 'Restart', icon: RefreshCw, color: 'info' },
    { id: 'ps', label: 'Status', icon: List, color: 'secondary' },
    { id: 'logs', label: 'Logs', icon: FileText, color: 'secondary' },
  ];

  const handleAction = async (action) => {
    if (!composeContent.trim()) {
      alert('Please provide docker-compose content');
      return;
    }

    setExecuting(true);
    setActiveAction(action);
    setResult(null);

    try {
      const response = await dinDAPI.runCompose(
        environmentId,
        composeContent,
        action,
        '',
        action === 'up' // detach for 'up' action
      );
      
      if (response.data.success) {
        setResult({
          success: response.data.data.success,
          action,
          ...response.data.data
        });
        onRefresh?.();
      }
    } catch (error) {
      setResult({
        success: false,
        action,
        error: error.response?.data?.message || error.message
      });
    } finally {
      setExecuting(false);
      setActiveAction(null);
    }
  };

  const handleTemplateSelect = (template) => {
    setComposeContent(template.content);
    setActiveTemplate(template.name);
    setResult(null);
  };

  const copyContent = () => {
    navigator.clipboard.writeText(composeContent);
  };

  const isDisabled = environmentStatus !== 'running';

  return (
    <div className="dind-compose">
      {isDisabled && (
        <div className="compose-warning">
          <AlertCircle size={16} />
          Environment is not running. Start it to use Docker Compose.
        </div>
      )}

      <div className="compose-layout">
        {/* Left Panel - Compose Editor */}
        <div className="compose-editor-panel">
          <div className="editor-header">
            <h3><Layers size={18} /> docker-compose.yml</h3>
            <button className="btn-icon-sm" onClick={copyContent} title="Copy">
              <Copy size={14} />
            </button>
          </div>

          {/* Templates */}
          <div className="compose-templates">
            <span className="templates-label">Templates:</span>
            <div className="templates-list">
              {templates.map((t, i) => (
                <button
                  key={i}
                  className={`template-btn ${activeTemplate === t.name ? 'active' : ''}`}
                  onClick={() => handleTemplateSelect(t)}
                >
                  {t.name}
                </button>
              ))}
            </div>
          </div>

          {/* Editor */}
          <textarea
            className="compose-editor"
            value={composeContent}
            onChange={(e) => { setComposeContent(e.target.value); setActiveTemplate(null); }}
            placeholder="Enter your docker-compose.yml content..."
            disabled={isDisabled}
            spellCheck={false}
          />
        </div>

        {/* Right Panel - Actions & Results */}
        <div className="compose-actions-panel">
          <div className="actions-section">
            <h3>Actions</h3>
            <div className="action-buttons">
              {actions.map(action => (
                <button
                  key={action.id}
                  className={`action-btn ${action.color} ${activeAction === action.id ? 'executing' : ''}`}
                  onClick={() => handleAction(action.id)}
                  disabled={isDisabled || executing}
                >
                  {executing && activeAction === action.id ? (
                    <RefreshCw className="spinning" size={16} />
                  ) : (
                    <action.icon size={16} />
                  )}
                  {action.label}
                </button>
              ))}
            </div>
          </div>

          {/* Result */}
          {result && (
            <div className={`compose-result ${result.success ? 'success' : 'error'}`}>
              <div className="result-header">
                {result.success ? (
                  <><CheckCircle size={18} /> {result.action} Successful</>
                ) : (
                  <><XCircle size={18} /> {result.action} Failed</>
                )}
              </div>

              {result.services && result.services.length > 0 && (
                <div className="result-services">
                  <h4>Services</h4>
                  <div className="services-list">
                    {result.services.map((s, i) => (
                      <span key={i} className="service-tag">{s}</span>
                    ))}
                  </div>
                </div>
              )}

              {result.output && (
                <div className="result-output">
                  <h4>Output</h4>
                  <pre>{result.output}</pre>
                </div>
              )}

              {result.error && (
                <div className="result-error">
                  <pre>{result.error}</pre>
                </div>
              )}
            </div>
          )}

          {/* Help Section */}
          <div className="compose-help">
            <h4>Docker Compose Commands</h4>
            <ul>
              <li><strong>Up:</strong> Create and start containers</li>
              <li><strong>Down:</strong> Stop and remove containers</li>
              <li><strong>Restart:</strong> Restart all services</li>
              <li><strong>Status:</strong> List containers status</li>
              <li><strong>Logs:</strong> View container logs</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DinDCompose;

