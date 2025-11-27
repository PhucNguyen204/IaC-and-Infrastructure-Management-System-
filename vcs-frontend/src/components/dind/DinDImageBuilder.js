import React, { useState } from 'react';
import { dinDAPI } from '../../api';
import { 
  Box, 
  Play, 
  FileText,
  RefreshCw,
  CheckCircle,
  XCircle,
  AlertCircle,
  Copy
} from 'lucide-react';
import './DinDImageBuilder.css';

const DinDImageBuilder = ({ environmentId, environmentStatus, onRefresh }) => {
  const [imageName, setImageName] = useState('');
  const [tag, setTag] = useState('latest');
  const [dockerfile, setDockerfile] = useState(`FROM alpine:latest

# Install packages
RUN apk add --no-cache curl

# Set working directory
WORKDIR /app

# Copy files (uncomment if needed)
# COPY . .

# Default command
CMD ["echo", "Hello from Docker!"]`);
  const [noCache, setNoCache] = useState(false);
  const [building, setBuilding] = useState(false);
  const [buildResult, setBuildResult] = useState(null);
  const [activeTemplate, setActiveTemplate] = useState(null);

  const templates = [
    {
      name: 'Alpine Base',
      dockerfile: `FROM alpine:latest

RUN apk add --no-cache curl wget

CMD ["sh"]`
    },
    {
      name: 'Node.js App',
      dockerfile: `FROM node:18-alpine

WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .

EXPOSE 3000
CMD ["npm", "start"]`
    },
    {
      name: 'Python Flask',
      dockerfile: `FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .

EXPOSE 5000
CMD ["python", "app.py"]`
    },
    {
      name: 'Nginx Static',
      dockerfile: `FROM nginx:alpine

COPY ./html /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]`
    },
    {
      name: 'Go App',
      dockerfile: `FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o main .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]`
    }
  ];

  const handleBuild = async () => {
    if (!imageName.trim() || !dockerfile.trim()) {
      alert('Please provide image name and Dockerfile content');
      return;
    }

    setBuilding(true);
    setBuildResult(null);

    try {
      const response = await dinDAPI.buildImage(
        environmentId, 
        dockerfile, 
        imageName, 
        tag, 
        noCache
      );
      
      if (response.data.success) {
        setBuildResult({
          success: response.data.data.success,
          ...response.data.data
        });
        onRefresh?.();
      }
    } catch (error) {
      setBuildResult({
        success: false,
        error: error.response?.data?.message || error.message
      });
    } finally {
      setBuilding(false);
    }
  };

  const handleTemplateSelect = (template) => {
    setDockerfile(template.dockerfile);
    setActiveTemplate(template.name);
  };

  const copyDockerfile = () => {
    navigator.clipboard.writeText(dockerfile);
  };

  const isDisabled = environmentStatus !== 'running';

  return (
    <div className="image-builder">
      {isDisabled && (
        <div className="builder-warning">
          <AlertCircle size={16} />
          Environment is not running. Start it to build images.
        </div>
      )}

      <div className="builder-layout">
        {/* Left Panel - Dockerfile Editor */}
        <div className="editor-panel">
          <div className="editor-header">
            <h3><FileText size={18} /> Dockerfile</h3>
            <div className="editor-actions">
              <button className="btn-icon-sm" onClick={copyDockerfile} title="Copy">
                <Copy size={14} />
              </button>
            </div>
          </div>

          {/* Templates */}
          <div className="templates-bar">
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
            className="dockerfile-editor"
            value={dockerfile}
            onChange={(e) => { setDockerfile(e.target.value); setActiveTemplate(null); }}
            placeholder="Enter your Dockerfile content here..."
            disabled={isDisabled}
            spellCheck={false}
          />
        </div>

        {/* Right Panel - Build Config & Results */}
        <div className="config-panel">
          <div className="config-section">
            <h3><Box size={18} /> Build Configuration</h3>
            
            <div className="form-group">
              <label>Image Name</label>
              <input
                type="text"
                placeholder="my-app"
                value={imageName}
                onChange={(e) => setImageName(e.target.value)}
                disabled={isDisabled}
              />
            </div>

            <div className="form-group">
              <label>Tag</label>
              <input
                type="text"
                placeholder="latest"
                value={tag}
                onChange={(e) => setTag(e.target.value)}
                disabled={isDisabled}
              />
            </div>

            <div className="form-group checkbox">
              <label>
                <input
                  type="checkbox"
                  checked={noCache}
                  onChange={(e) => setNoCache(e.target.checked)}
                  disabled={isDisabled}
                />
                Build without cache (--no-cache)
              </label>
            </div>

            <button
              className="build-btn"
              onClick={handleBuild}
              disabled={isDisabled || building || !imageName.trim()}
            >
              {building ? (
                <>
                  <RefreshCw className="spinning" size={16} />
                  Building...
                </>
              ) : (
                <>
                  <Play size={16} />
                  Build Image
                </>
              )}
            </button>
          </div>

          {/* Build Result */}
          {buildResult && (
            <div className={`build-result ${buildResult.success ? 'success' : 'error'}`}>
              <div className="result-header">
                {buildResult.success ? (
                  <><CheckCircle size={18} /> Build Successful</>
                ) : (
                  <><XCircle size={18} /> Build Failed</>
                )}
              </div>
              
              {buildResult.success ? (
                <div className="result-details">
                  <div className="detail-row">
                    <span>Image:</span>
                    <code>{buildResult.image_name}:{buildResult.tag}</code>
                  </div>
                  {buildResult.image_id && (
                    <div className="detail-row">
                      <span>Image ID:</span>
                      <code>{buildResult.image_id?.substring(0, 12)}</code>
                    </div>
                  )}
                  {buildResult.duration && (
                    <div className="detail-row">
                      <span>Duration:</span>
                      <code>{buildResult.duration}</code>
                    </div>
                  )}
                </div>
              ) : (
                <div className="result-error">
                  {buildResult.error}
                </div>
              )}

              {buildResult.build_logs && buildResult.build_logs.length > 0 && (
                <div className="build-logs">
                  <h4>Build Logs</h4>
                  <pre>
                    {buildResult.build_logs.slice(-50).join('\n')}
                  </pre>
                </div>
              )}
            </div>
          )}

          {/* Quick Run */}
          {buildResult?.success && (
            <div className="quick-run">
              <h4>Quick Run</h4>
              <code>docker run --rm {imageName}:{tag}</code>
              <p className="hint">Go to Terminal tab to run this command</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default DinDImageBuilder;

