import React, { useState, useRef, useEffect } from 'react';
import { dinDAPI } from '../../api';
import { 
  Terminal as TerminalIcon, 
  Send, 
  Trash2, 
  Download,
  Copy,
  Server,
  Layers,
  RefreshCw,
  ChevronRight,
  AlertCircle
} from 'lucide-react';
import './DinDTerminal.css';

const DinDTerminal = ({ environmentId, environmentStatus, onRefresh }) => {
  const [command, setCommand] = useState('');
  const [history, setHistory] = useState([]);
  const [executing, setExecuting] = useState(false);
  const [containers, setContainers] = useState([]);
  const [images, setImages] = useState([]);
  const [activePanel, setActivePanel] = useState('terminal');
  const [commandHistory, setCommandHistory] = useState([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const terminalRef = useRef(null);
  const inputRef = useRef(null);

  // Scroll to bottom when new output is added
  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [history]);

  // Load containers and images
  useEffect(() => {
    if (environmentStatus === 'running') {
      loadContainers();
      loadImages();
    }
  }, [environmentId, environmentStatus]);

  const loadContainers = async () => {
    try {
      const response = await dinDAPI.listContainers(environmentId);
      if (response.data.success) {
        setContainers(response.data.data.containers || []);
      }
    } catch (error) {
      console.error('Error loading containers:', error);
    }
  };

  const loadImages = async () => {
    try {
      const response = await dinDAPI.listImages(environmentId);
      if (response.data.success) {
        setImages(response.data.data.images || []);
      }
    } catch (error) {
      console.error('Error loading images:', error);
    }
  };

  const executeCommand = async (cmd) => {
    if (!cmd.trim() || executing || environmentStatus !== 'running') return;

    const commandToRun = cmd.trim();
    
    // Add to history display
    setHistory(prev => [...prev, { type: 'command', content: commandToRun }]);
    
    // Add to command history for navigation
    setCommandHistory(prev => [...prev, commandToRun]);
    setHistoryIndex(-1);
    
    setCommand('');
    setExecuting(true);

    try {
      const response = await dinDAPI.execCommand(environmentId, commandToRun);
      if (response.data.success) {
        const data = response.data.data;
        setHistory(prev => [...prev, { 
          type: data.exit_code === 0 ? 'output' : 'error',
          content: data.output || '(no output)',
          exitCode: data.exit_code,
          duration: data.duration
        }]);
      }
    } catch (error) {
      setHistory(prev => [...prev, { 
        type: 'error', 
        content: error.response?.data?.message || error.message 
      }]);
    } finally {
      setExecuting(false);
      loadContainers();
      loadImages();
      onRefresh?.();
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      executeCommand(command);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (commandHistory.length > 0) {
        const newIndex = historyIndex < commandHistory.length - 1 ? historyIndex + 1 : historyIndex;
        setHistoryIndex(newIndex);
        setCommand(commandHistory[commandHistory.length - 1 - newIndex] || '');
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIndex > 0) {
        const newIndex = historyIndex - 1;
        setHistoryIndex(newIndex);
        setCommand(commandHistory[commandHistory.length - 1 - newIndex] || '');
      } else if (historyIndex === 0) {
        setHistoryIndex(-1);
        setCommand('');
      }
    }
  };

  const quickCommands = [
    { label: 'docker ps', cmd: 'docker ps' },
    { label: 'docker images', cmd: 'docker images' },
    { label: 'docker ps -a', cmd: 'docker ps -a' },
    { label: 'docker network ls', cmd: 'docker network ls' },
    { label: 'docker volume ls', cmd: 'docker volume ls' },
    { label: 'docker system df', cmd: 'docker system df' },
  ];

  const clearTerminal = () => {
    setHistory([]);
  };

  const copyOutput = () => {
    const text = history
      .map(h => h.type === 'command' ? `$ ${h.content}` : h.content)
      .join('\n');
    navigator.clipboard.writeText(text);
  };

  const downloadLogs = () => {
    const text = history
      .map(h => h.type === 'command' ? `$ ${h.content}` : h.content)
      .join('\n');
    const blob = new Blob([text], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `dind-terminal-${new Date().toISOString()}.log`;
    a.click();
  };

  const isDisabled = environmentStatus !== 'running';

  return (
    <div className="dind-terminal-container">
      {/* Sub-tabs */}
      <div className="terminal-tabs">
        <button 
          className={`terminal-tab ${activePanel === 'terminal' ? 'active' : ''}`}
          onClick={() => setActivePanel('terminal')}
        >
          <TerminalIcon size={14} /> Terminal
        </button>
        <button 
          className={`terminal-tab ${activePanel === 'containers' ? 'active' : ''}`}
          onClick={() => { setActivePanel('containers'); loadContainers(); }}
        >
          <Server size={14} /> Containers ({containers.length})
        </button>
        <button 
          className={`terminal-tab ${activePanel === 'images' ? 'active' : ''}`}
          onClick={() => { setActivePanel('images'); loadImages(); }}
        >
          <Layers size={14} /> Images ({images.length})
        </button>
      </div>

      {activePanel === 'terminal' && (
        <div className="terminal-panel">
          {/* Quick Commands */}
          <div className="quick-commands">
            <span className="quick-label">Quick:</span>
            {quickCommands.map((qc, i) => (
              <button 
                key={i}
                className="quick-cmd-btn"
                onClick={() => executeCommand(qc.cmd)}
                disabled={isDisabled || executing}
              >
                {qc.label}
              </button>
            ))}
          </div>

          {/* Terminal Output */}
          <div className="terminal-output" ref={terminalRef}>
            {isDisabled && (
              <div className="terminal-warning">
                <AlertCircle size={16} />
                Environment is not running. Start it to use the terminal.
              </div>
            )}
            
            <div className="terminal-welcome">
              <span className="welcome-text">Docker-in-Docker Terminal</span>
              <span className="welcome-hint">Run any docker command in this isolated environment</span>
            </div>

            {history.map((item, index) => (
              <div key={index} className={`terminal-line ${item.type}`}>
                {item.type === 'command' ? (
                  <div className="command-line">
                    <span className="prompt"><ChevronRight size={14} /></span>
                    <span className="command-text">{item.content}</span>
                  </div>
                ) : (
                  <div className="output-block">
                    <pre>{item.content}</pre>
                    {item.duration && (
                      <div className="output-meta">
                        <span className={`exit-code ${item.exitCode === 0 ? 'success' : 'error'}`}>
                          Exit: {item.exitCode}
                        </span>
                        <span className="duration">{item.duration}</span>
                      </div>
                    )}
                  </div>
                )}
              </div>
            ))}

            {executing && (
              <div className="terminal-line executing">
                <RefreshCw className="spinning" size={14} />
                <span>Executing...</span>
              </div>
            )}
          </div>

          {/* Terminal Input */}
          <div className="terminal-input-container">
            <span className="input-prompt">$</span>
            <input
              ref={inputRef}
              type="text"
              value={command}
              onChange={(e) => setCommand(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={isDisabled ? "Environment not running" : "Enter docker command..."}
              disabled={isDisabled || executing}
              autoFocus
            />
            <button 
              className="send-btn"
              onClick={() => executeCommand(command)}
              disabled={isDisabled || executing || !command.trim()}
            >
              <Send size={16} />
            </button>
          </div>

          {/* Terminal Actions */}
          <div className="terminal-actions">
            <button className="action-btn" onClick={clearTerminal} title="Clear terminal">
              <Trash2 size={14} /> Clear
            </button>
            <button className="action-btn" onClick={copyOutput} title="Copy output">
              <Copy size={14} /> Copy
            </button>
            <button className="action-btn" onClick={downloadLogs} title="Download logs">
              <Download size={14} /> Download
            </button>
          </div>
        </div>
      )}

      {activePanel === 'containers' && (
        <div className="list-panel">
          <div className="list-header">
            <h3><Server size={18} /> Containers in this environment</h3>
            <button className="btn-icon-sm" onClick={loadContainers}>
              <RefreshCw size={14} />
            </button>
          </div>
          
          {containers.length === 0 ? (
            <div className="empty-list">
              <Server size={32} strokeWidth={1} />
              <p>No containers running</p>
              <span>Run <code>docker run</code> to start a container</span>
            </div>
          ) : (
            <table className="data-table">
              <thead>
                <tr>
                  <th>Container ID</th>
                  <th>Name</th>
                  <th>Image</th>
                  <th>Status</th>
                  <th>Ports</th>
                </tr>
              </thead>
              <tbody>
                {containers.map((c, i) => (
                  <tr key={i}>
                    <td className="mono">{c.container_id?.substring(0, 12)}</td>
                    <td>{c.name}</td>
                    <td className="mono">{c.image}</td>
                    <td>
                      <span className={`status-pill ${c.status?.includes('Up') ? 'up' : 'down'}`}>
                        {c.status}
                      </span>
                    </td>
                    <td className="mono">{c.ports || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {activePanel === 'images' && (
        <div className="list-panel">
          <div className="list-header">
            <h3><Layers size={18} /> Images in this environment</h3>
            <button className="btn-icon-sm" onClick={loadImages}>
              <RefreshCw size={14} />
            </button>
          </div>
          
          {images.length === 0 ? (
            <div className="empty-list">
              <Layers size={32} strokeWidth={1} />
              <p>No images available</p>
              <span>Pull an image with <code>docker pull</code> or build one</span>
            </div>
          ) : (
            <table className="data-table">
              <thead>
                <tr>
                  <th>Image ID</th>
                  <th>Repository</th>
                  <th>Tag</th>
                  <th>Size</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {images.map((img, i) => (
                  <tr key={i}>
                    <td className="mono">{img.image_id?.substring(0, 12)}</td>
                    <td>{img.repository}</td>
                    <td className="mono">{img.tag}</td>
                    <td>{img.size}</td>
                    <td>{img.created}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  );
};

export default DinDTerminal;

