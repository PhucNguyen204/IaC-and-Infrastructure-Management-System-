import React, { useState, useEffect, useRef } from 'react';
import { 
  Play, 
  Database, 
  Table, 
  RefreshCw, 
  Download, 
  Copy, 
  Check,
  ChevronRight,
  ChevronDown,
  AlertCircle,
  Clock,
  Layers
} from 'lucide-react';
import { clusterAPI } from '../../api';
import toast from 'react-hot-toast';
import './SQLConsole.css';

const SQLConsole = ({ clusterId }) => {
  const [query, setQuery] = useState('SELECT version();');
  const [selectedDatabase, setSelectedDatabase] = useState('postgres');
  const [databases, setDatabases] = useState(['postgres']);
  const [tables, setTables] = useState([]);
  const [expandedTable, setExpandedTable] = useState(null);
  const [tableSchema, setTableSchema] = useState(null);
  const [result, setResult] = useState(null);
  const [executing, setExecuting] = useState(false);
  const [loadingTables, setLoadingTables] = useState(false);
  const [copiedCell, setCopiedCell] = useState(null);
  const [queryHistory, setQueryHistory] = useState([]);
  const textareaRef = useRef(null);

  useEffect(() => {
    loadDatabases();
  }, [clusterId]);

  useEffect(() => {
    if (selectedDatabase) {
      loadTables();
    }
  }, [selectedDatabase]);

  const loadDatabases = async () => {
    try {
      const response = await clusterAPI.listDatabases(clusterId);
      const dbList = response.data?.map(db => db.name) || ['postgres'];
      setDatabases(dbList);
    } catch (error) {
      console.error('Error loading databases:', error);
    }
  };

  const loadTables = async () => {
    try {
      setLoadingTables(true);
      const response = await clusterAPI.getTables(clusterId, selectedDatabase);
      setTables(response.data || []);
    } catch (error) {
      console.error('Error loading tables:', error);
      setTables([]);
    } finally {
      setLoadingTables(false);
    }
  };

  const loadTableSchema = async (tableName) => {
    try {
      const response = await clusterAPI.getTableSchema(clusterId, selectedDatabase, tableName);
      setTableSchema(response.data);
    } catch (error) {
      console.error('Error loading schema:', error);
      toast.error('Failed to load table schema');
    }
  };

  const executeQuery = async () => {
    if (!query.trim()) {
      toast.error('Please enter a query');
      return;
    }

    try {
      setExecuting(true);
      setResult(null);

      const response = await clusterAPI.executeQuery(clusterId, query, selectedDatabase);
      setResult(response.data);
      
      // Add to history
      setQueryHistory(prev => [
        { query: query.trim(), timestamp: new Date(), success: true },
        ...prev.slice(0, 9)
      ]);

      toast.success(`Query executed in ${response.data.duration}`);
    } catch (error) {
      const errorMsg = error.response?.data?.error || error.message;
      setResult({ error: errorMsg });
      setQueryHistory(prev => [
        { query: query.trim(), timestamp: new Date(), success: false, error: errorMsg },
        ...prev.slice(0, 9)
      ]);
      toast.error('Query failed');
    } finally {
      setExecuting(false);
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      executeQuery();
    }
  };

  const insertTableName = (tableName) => {
    const cursorPos = textareaRef.current?.selectionStart || query.length;
    const newQuery = query.slice(0, cursorPos) + tableName + query.slice(cursorPos);
    setQuery(newQuery);
    textareaRef.current?.focus();
  };

  const copyCell = async (value) => {
    try {
      await navigator.clipboard.writeText(String(value));
      setCopiedCell(value);
      setTimeout(() => setCopiedCell(null), 1500);
    } catch (err) {
      toast.error('Failed to copy');
    }
  };

  const exportResults = (format) => {
    if (!result || !result.rows) return;

    let content, filename, type;
    
    if (format === 'csv') {
      const headers = result.columns.join(',');
      const rows = result.rows.map(row => row.map(cell => `"${cell}"`).join(','));
      content = [headers, ...rows].join('\n');
      filename = 'query_results.csv';
      type = 'text/csv';
    } else {
      content = JSON.stringify({ columns: result.columns, rows: result.rows }, null, 2);
      filename = 'query_results.json';
      type = 'application/json';
    }

    const blob = new Blob([content], { type });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  };

  const quickQueries = [
    { label: 'Show Tables', query: "SELECT tablename FROM pg_tables WHERE schemaname = 'public';" },
    { label: 'Database Size', query: "SELECT pg_size_pretty(pg_database_size(current_database()));" },
    { label: 'Active Connections', query: "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';" },
    { label: 'Table Sizes', query: "SELECT relname, pg_size_pretty(pg_total_relation_size(relid)) FROM pg_catalog.pg_statio_user_tables ORDER BY pg_total_relation_size(relid) DESC LIMIT 10;" },
  ];

  return (
    <div className="sql-console">
      <div className="console-layout">
        {/* Left Sidebar - Schema Browser */}
        <div className="schema-browser">
          <div className="browser-header">
            <Database size={16} />
            <select 
              value={selectedDatabase}
              onChange={(e) => setSelectedDatabase(e.target.value)}
              className="db-select"
            >
              {databases.map(db => (
                <option key={db} value={db}>{db}</option>
              ))}
            </select>
            <button className="refresh-btn" onClick={loadTables} disabled={loadingTables}>
              <RefreshCw size={14} className={loadingTables ? 'spin' : ''} />
            </button>
          </div>

          <div className="tables-list">
            {loadingTables ? (
              <div className="loading-tables">
                <RefreshCw size={16} className="spin" />
                Loading...
              </div>
            ) : tables.length === 0 ? (
              <div className="no-tables">No tables found</div>
            ) : (
              tables.map((table, idx) => (
                <div key={idx} className="table-item">
                  <div 
                    className="table-header"
                    onClick={() => {
                      if (expandedTable === table.name) {
                        setExpandedTable(null);
                        setTableSchema(null);
                      } else {
                        setExpandedTable(table.name);
                        loadTableSchema(table.name);
                      }
                    }}
                  >
                    {expandedTable === table.name ? (
                      <ChevronDown size={14} />
                    ) : (
                      <ChevronRight size={14} />
                    )}
                    <Table size={14} />
                    <span className="table-name">{table.name}</span>
                    <button 
                      className="insert-btn"
                      onClick={(e) => {
                        e.stopPropagation();
                        insertTableName(table.name);
                      }}
                      title="Insert table name"
                    >
                      +
                    </button>
                  </div>
                  
                  {expandedTable === table.name && tableSchema && (
                    <div className="table-columns">
                      {tableSchema.columns?.map((col, cidx) => (
                        <div key={cidx} className="column-item">
                          <span className="col-name">{col.name}</span>
                          <span className="col-type">{col.data_type}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        </div>

        {/* Main Content */}
        <div className="console-main">
          {/* Query Editor */}
          <div className="query-editor">
            <div className="editor-header">
              <span className="editor-title">
                <Layers size={16} />
                SQL Query
              </span>
              <div className="quick-queries">
                {quickQueries.map((q, idx) => (
                  <button 
                    key={idx}
                    className="quick-query-btn"
                    onClick={() => setQuery(q.query)}
                  >
                    {q.label}
                  </button>
                ))}
              </div>
            </div>
            
            <textarea
              ref={textareaRef}
              className="query-input"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Enter your SQL query here... (Ctrl+Enter to execute)"
              spellCheck={false}
            />

            <div className="editor-footer">
              <span className="hint">
                Press <kbd>Ctrl</kbd>+<kbd>Enter</kbd> to execute
              </span>
              <button 
                className="execute-btn"
                onClick={executeQuery}
                disabled={executing}
              >
                {executing ? (
                  <>
                    <RefreshCw size={16} className="spin" />
                    Executing...
                  </>
                ) : (
                  <>
                    <Play size={16} />
                    Execute
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Results */}
          <div className="results-section">
            <div className="results-header">
              <span className="results-title">Results</span>
              {result && !result.error && (
                <div className="results-actions">
                  <span className="row-count">
                    {result.row_count} rows
                  </span>
                  <span className="duration">
                    <Clock size={14} />
                    {result.duration}
                  </span>
                  <button onClick={() => exportResults('csv')}>
                    <Download size={14} />
                    CSV
                  </button>
                  <button onClick={() => exportResults('json')}>
                    <Download size={14} />
                    JSON
                  </button>
                </div>
              )}
            </div>

            <div className="results-content">
              {!result ? (
                <div className="no-results">
                  <Play size={24} />
                  <span>Execute a query to see results</span>
                </div>
              ) : result.error ? (
                <div className="error-result">
                  <AlertCircle size={20} />
                  <span>{result.error}</span>
                </div>
              ) : (
                <div className="results-table-wrapper">
                  <table className="results-table">
                    <thead>
                      <tr>
                        {result.columns?.map((col, idx) => (
                          <th key={idx}>{col}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {result.rows?.map((row, ridx) => (
                        <tr key={ridx}>
                          {row.map((cell, cidx) => (
                            <td 
                              key={cidx}
                              onClick={() => copyCell(cell)}
                              title="Click to copy"
                            >
                              {copiedCell === cell ? (
                                <Check size={12} className="copied-icon" />
                              ) : null}
                              {cell === null ? (
                                <span className="null-value">NULL</span>
                              ) : (
                                String(cell)
                              )}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right Sidebar - History */}
        <div className="query-history">
          <div className="history-header">
            <Clock size={16} />
            History
          </div>
          <div className="history-list">
            {queryHistory.length === 0 ? (
              <div className="no-history">No queries yet</div>
            ) : (
              queryHistory.map((item, idx) => (
                <div 
                  key={idx}
                  className={`history-item ${item.success ? 'success' : 'error'}`}
                  onClick={() => setQuery(item.query)}
                >
                  <div className="history-query">{item.query.slice(0, 50)}...</div>
                  <div className="history-time">
                    {item.timestamp.toLocaleTimeString()}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SQLConsole;

