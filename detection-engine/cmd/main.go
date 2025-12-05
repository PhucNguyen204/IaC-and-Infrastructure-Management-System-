package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
)

// Config from environment
type Config struct {
	// ClickHouse config (for reading logs)
	CHHost     string
	CHPort     string
	CHUser     string
	CHPassword string
	CHDatabase string

	// PostgreSQL config (for storing alerts)
	PGHost     string
	PGPort     string
	PGUser     string
	PGPassword string
	PGDatabase string

	// Engine config
	LogLevel         string
	LogTable         string
	RulesStoragePath string
	LogFilePath      string
}

// Rule structure
type Rule struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Severity    string   `yaml:"severity" json:"severity"`
	Keywords    []string `yaml:"keywords" json:"keywords"` // Simple keyword matching
	Enabled     bool     `yaml:"enabled" json:"enabled"`
}

// LogEntry from ClickHouse
type LogEntry struct {
	ID        string    `json:"id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
}

// Alert structure
type Alert struct {
	ID          string    `json:"id"`
	RuleID      string    `json:"rule_id"`
	RuleName    string    `json:"rule_name"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	LogID       string    `json:"log_id"`
	MatchedText string    `json:"matched_text"`
	CreatedAt   time.Time `json:"created_at"`
}

// DetectionEngine main struct
type DetectionEngine struct {
	config      Config
	chConn      driver.Conn // ClickHouse connection
	pgDB        *sql.DB     // PostgreSQL connection
	rules       []Rule
	memAlerts   []Alert // In-memory alerts cache
	logFile     *os.File
	isHealthy   bool
	startTime   time.Time
	lastCheckAt time.Time
}

var engine *DetectionEngine

func main() {
	log.Println("===========================================")
	log.Println("   Detection Engine v2.0.0 Starting...")
	log.Println("   ClickHouse -> Rule Match -> PostgreSQL")
	log.Println("===========================================")

	engine = &DetectionEngine{
		startTime: time.Now(),
		memAlerts: make([]Alert, 0),
	}

	// Load config from environment
	engine.loadConfig()

	// Initialize log file
	if err := engine.initLogFile(); err != nil {
		log.Printf("Warning: Could not init log file: %v", err)
	}

	// Connect to ClickHouse (source for logs) with retry
	log.Println("Connecting to ClickHouse...")
	engine.connectClickHouseWithRetry(10, 3*time.Second)

	// Connect to PostgreSQL (destination for alerts) with retry
	// PostgreSQL might take longer to be ready (Patroni cluster initialization)
	log.Println("Connecting to PostgreSQL (with retry for cluster initialization)...")
	engine.connectPostgresWithRetry(30, 2*time.Second) // Up to 60 seconds of retries

	// Load rules
	if err := engine.loadRules(); err != nil {
		log.Printf("Warning: Could not load rules: %v", err)
	}

	engine.isHealthy = true

	// Start background reconnection loop for PostgreSQL
	go engine.backgroundPostgresReconnect()

	// Start background detection loop
	go engine.detectionLoop()

	// Start API server
	engine.startAPI()
}

func (e *DetectionEngine) loadConfig() {
	e.config = Config{
		// ClickHouse (logs source)
		CHHost:     getEnv("CH_HOST", getEnv("DB_HOST", "localhost")),
		CHPort:     getEnv("CH_PORT", getEnv("DB_PORT", "9000")),
		CHUser:     getEnv("CH_USER", getEnv("DB_USER", "default")),
		CHPassword: getEnv("CH_PASSWORD", getEnv("DB_PASSWORD", "")),
		CHDatabase: getEnv("CH_DATABASE", getEnv("DB_NAME", "detection_db")),

		// PostgreSQL (alerts destination)
		PGHost:     getEnv("PG_HOST", "localhost"),
		PGPort:     getEnv("PG_PORT", "5432"),
		PGUser:     getEnv("PG_USER", "postgres"),
		PGPassword: getEnv("PG_PASSWORD", "postgres"),
		PGDatabase: getEnv("PG_DATABASE", "alerts_db"),

		// Engine
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		LogTable:         getEnv("LOG_TABLE", "logs"),
		RulesStoragePath: getEnv("RULES_STORAGE_PATH", "/opt/rules_storage"),
		LogFilePath:      getEnv("LOG_FILE_PATH", "/var/log/detection-engine/engine.log"),
	}

	log.Printf("Config loaded:")
	log.Printf("  ClickHouse (logs): %s:%s/%s", e.config.CHHost, e.config.CHPort, e.config.CHDatabase)
	log.Printf("  PostgreSQL (alerts): %s:%s/%s", e.config.PGHost, e.config.PGPort, e.config.PGDatabase)
	log.Printf("  Rules Path: %s", e.config.RulesStoragePath)
}

func (e *DetectionEngine) initLogFile() error {
	dir := filepath.Dir(e.config.LogFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(e.config.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	e.logFile = f
	return nil
}

func (e *DetectionEngine) writeLog(message string) {
	logEntry := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), message)
	log.Print(message)
	if e.logFile != nil {
		e.logFile.WriteString(logEntry)
	}
}

// ============== ClickHouse Connection ==============

func (e *DetectionEngine) connectClickHouse() error {
	ctx := context.Background()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", e.config.CHHost, e.config.CHPort)},
		Auth: clickhouse.Auth{
			Database: e.config.CHDatabase,
			Username: e.config.CHUser,
			Password: e.config.CHPassword,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:     10 * time.Second,
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})

	if err != nil {
		return fmt.Errorf("failed to open CH connection: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping CH: %w", err)
	}

	e.chConn = conn
	e.writeLog(fmt.Sprintf("Connected to ClickHouse at %s:%s", e.config.CHHost, e.config.CHPort))
	return nil
}

func (e *DetectionEngine) createClickHouseTables() error {
	ctx := context.Background()

	// Create database
	if err := e.chConn.Exec(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", e.config.CHDatabase)); err != nil {
		return err
	}

	// Create logs table
	logTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s (
			id String,
			level String,
			message String,
			source String,
			timestamp DateTime DEFAULT now()
		) ENGINE = MergeTree()
		ORDER BY (timestamp, id)
	`, e.config.CHDatabase, e.config.LogTable)

	if err := e.chConn.Exec(ctx, logTableSQL); err != nil {
		return fmt.Errorf("failed to create logs table: %w", err)
	}

	e.writeLog("ClickHouse tables created/verified")
	return nil
}

// connectClickHouseWithRetry attempts to connect to ClickHouse with retries
func (e *DetectionEngine) connectClickHouseWithRetry(maxRetries int, retryInterval time.Duration) {
	for i := 0; i < maxRetries; i++ {
		if err := e.connectClickHouse(); err == nil {
			// Successfully connected, create tables
			if err := e.createClickHouseTables(); err != nil {
				e.writeLog(fmt.Sprintf("Warning: Could not create CH tables: %v", err))
			}
			return
		} else {
			e.writeLog(fmt.Sprintf("ClickHouse connection attempt %d/%d failed: %v", i+1, maxRetries, err))
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
			}
		}
	}
	e.writeLog("ClickHouse connection failed after all retries. Engine will run without ClickHouse.")
}

// ============== PostgreSQL Connection ==============

func (e *DetectionEngine) connectPostgres() error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=10",
		e.config.PGHost, e.config.PGPort, e.config.PGUser, e.config.PGPassword, e.config.PGDatabase,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open PG connection: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping PG: %w", err)
	}

	e.pgDB = db
	e.writeLog(fmt.Sprintf("Connected to PostgreSQL at %s:%s", e.config.PGHost, e.config.PGPort))
	return nil
}

// connectPostgresWithRetry attempts to connect to PostgreSQL with retries
func (e *DetectionEngine) connectPostgresWithRetry(maxRetries int, retryInterval time.Duration) {
	for i := 0; i < maxRetries; i++ {
		if err := e.connectPostgres(); err == nil {
			// Successfully connected, create tables
			if err := e.createPostgresTables(); err != nil {
				e.writeLog(fmt.Sprintf("Warning: Could not create PG tables: %v", err))
			}
			return
		} else {
			e.writeLog(fmt.Sprintf("PostgreSQL connection attempt %d/%d failed: %v", i+1, maxRetries, err))
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
			}
		}
	}
	e.writeLog("PostgreSQL connection failed after all retries. Alerts will be stored in memory only.")
}

// backgroundPostgresReconnect continuously tries to reconnect to PostgreSQL if disconnected
func (e *DetectionEngine) backgroundPostgresReconnect() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Check if PostgreSQL is connected
		if e.pgDB == nil {
			e.writeLog("Attempting to reconnect to PostgreSQL...")
			if err := e.connectPostgres(); err == nil {
				if err := e.createPostgresTables(); err != nil {
					e.writeLog(fmt.Sprintf("Warning: Could not create PG tables: %v", err))
				} else {
					e.writeLog("PostgreSQL reconnected successfully!")
				}
			}
		} else {
			// Check if connection is still alive
			if err := e.pgDB.Ping(); err != nil {
				e.writeLog(fmt.Sprintf("PostgreSQL connection lost: %v. Reconnecting...", err))
				e.pgDB.Close()
				e.pgDB = nil
			}
		}
	}
}

func (e *DetectionEngine) createPostgresTables() error {
	// Create alerts table
	alertTableSQL := `
		CREATE TABLE IF NOT EXISTS alerts (
			id VARCHAR(36) PRIMARY KEY,
			rule_id VARCHAR(50) NOT NULL,
			rule_name VARCHAR(200) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			message TEXT NOT NULL,
			log_id VARCHAR(36),
			matched_text TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
		CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);
	`

	if _, err := e.pgDB.Exec(alertTableSQL); err != nil {
		return fmt.Errorf("failed to create alerts table: %w", err)
	}

	e.writeLog("PostgreSQL tables created/verified")
	return nil
}

// ============== Rules Loading ==============

func (e *DetectionEngine) loadRules() error {
	rulesPath := e.config.RulesStoragePath
	e.rules = make([]Rule, 0)

	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		e.writeLog(fmt.Sprintf("Rules path %s does not exist, creating with sample rules", rulesPath))
		os.MkdirAll(rulesPath, 0755)

		// Create sample rules with keyword matching
		sampleRules := []Rule{
			{
				ID:          "RULE-001",
				Name:        "Error Detection",
				Description: "Detect error messages in logs",
				Severity:    "high",
				Keywords:    []string{"error", "ERROR", "failed", "FAILED", "exception"},
				Enabled:     true,
			},
			{
				ID:          "RULE-002",
				Name:        "Security Alert",
				Description: "Detect security-related events",
				Severity:    "critical",
				Keywords:    []string{"unauthorized", "attack", "intrusion", "malware", "breach"},
				Enabled:     true,
			},
			{
				ID:          "RULE-003",
				Name:        "Warning Detection",
				Description: "Detect warning messages",
				Severity:    "medium",
				Keywords:    []string{"warning", "WARN", "deprecated"},
				Enabled:     true,
			},
			{
				ID:          "RULE-004",
				Name:        "Connection Issues",
				Description: "Detect connection problems",
				Severity:    "high",
				Keywords:    []string{"connection refused", "timeout", "unreachable", "disconnected"},
				Enabled:     true,
			},
		}

		data, _ := yaml.Marshal(sampleRules)
		os.WriteFile(filepath.Join(rulesPath, "default_rules.yaml"), data, 0644)
	}

	// Load all YAML files
	files, err := os.ReadDir(rulesPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		filePath := filepath.Join(rulesPath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var fileRules []Rule
		if err := yaml.Unmarshal(data, &fileRules); err != nil {
			continue
		}
		e.rules = append(e.rules, fileRules...)
	}

	e.writeLog(fmt.Sprintf("Loaded %d rules from %s", len(e.rules), rulesPath))
	return nil
}

// ============== Detection Logic ==============

func (e *DetectionEngine) detectionLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		e.runDetection()
	}
}

func (e *DetectionEngine) runDetection() {
	if e.chConn == nil {
		return
	}

	e.writeLog("Running detection cycle...")

	// Get logs since last check
	since := e.lastCheckAt
	if since.IsZero() {
		since = time.Now().Add(-5 * time.Minute)
	}

	logs, err := e.fetchLogsFromClickHouse(since)
	if err != nil {
		e.writeLog(fmt.Sprintf("Failed to fetch logs: %v", err))
		return
	}

	e.lastCheckAt = time.Now()

	if len(logs) == 0 {
		return
	}

	e.writeLog(fmt.Sprintf("Processing %d logs", len(logs)))

	// Match each log against rules
	for _, logEntry := range logs {
		for _, rule := range e.rules {
			if !rule.Enabled {
				continue
			}

			if matched, keyword := e.matchRule(logEntry, rule); matched {
				alert := Alert{
					ID:          uuid.New().String(),
					RuleID:      rule.ID,
					RuleName:    rule.Name,
					Severity:    rule.Severity,
					Message:     fmt.Sprintf("Rule '%s' matched: %s", rule.Name, rule.Description),
					LogID:       logEntry.ID,
					MatchedText: keyword,
					CreatedAt:   time.Now(),
				}

				// Store in PostgreSQL
				if err := e.storeAlertInPostgres(alert); err != nil {
					e.writeLog(fmt.Sprintf("Failed to store alert in PG: %v", err))
				}

				// Also keep in memory cache
				e.memAlerts = append(e.memAlerts, alert)
				if len(e.memAlerts) > 1000 {
					e.memAlerts = e.memAlerts[len(e.memAlerts)-1000:]
				}

				e.writeLog(fmt.Sprintf("ALERT [%s]: %s (matched: %s)", alert.Severity, alert.Message, keyword))
			}
		}
	}
}

func (e *DetectionEngine) fetchLogsFromClickHouse(since time.Time) ([]LogEntry, error) {
	ctx := context.Background()

	query := fmt.Sprintf(`
		SELECT id, level, message, source, timestamp 
		FROM %s.%s 
		WHERE timestamp >= ? 
		ORDER BY timestamp ASC
		LIMIT 1000
	`, e.config.CHDatabase, e.config.LogTable)

	rows, err := e.chConn.Query(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var log LogEntry
		if err := rows.Scan(&log.ID, &log.Level, &log.Message, &log.Source, &log.Timestamp); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (e *DetectionEngine) matchRule(logEntry LogEntry, rule Rule) (bool, string) {
	// Simple keyword matching
	text := strings.ToLower(logEntry.Message + " " + logEntry.Level)

	for _, keyword := range rule.Keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true, keyword
		}
	}
	return false, ""
}

func (e *DetectionEngine) storeAlertInPostgres(alert Alert) error {
	if e.pgDB == nil {
		return nil
	}

	query := `
		INSERT INTO alerts (id, rule_id, rule_name, severity, message, log_id, matched_text, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := e.pgDB.Exec(query,
		alert.ID, alert.RuleID, alert.RuleName, alert.Severity,
		alert.Message, alert.LogID, alert.MatchedText, alert.CreatedAt,
	)
	return err
}

// ============== API Server ==============

func (e *DetectionEngine) startAPI() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Health check
	r.GET("/health", e.healthHandler)

	// Config
	r.GET("/config", e.configHandler)

	// Rules
	r.GET("/rules", e.getRulesHandler)
	r.POST("/rules/reload", e.reloadRulesHandler)

	// Alerts - Main feature
	r.GET("/alerts", e.getAlertsHandler)
	r.GET("/alerts/count", e.getAlertsCountHandler)

	// Logs - Insert test logs to ClickHouse
	r.POST("/logs", e.insertLogHandler)
	r.GET("/logs", e.getLogsHandler)

	// Manual detection trigger
	r.POST("/detect", e.triggerDetectionHandler)

	// Query endpoints
	r.POST("/query/clickhouse", e.queryClickHouseHandler)
	r.POST("/query/postgres", e.queryPostgresHandler)

	port := getEnv("API_PORT", "8000")
	e.writeLog(fmt.Sprintf("Starting API server on port %s", port))

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start API: %v", err)
	}
}

func (e *DetectionEngine) healthHandler(c *gin.Context) {
	chStatus := "disconnected"
	if e.chConn != nil {
		if err := e.chConn.Ping(context.Background()); err == nil {
			chStatus = "connected"
		}
	}

	pgStatus := "disconnected"
	if e.pgDB != nil {
		if err := e.pgDB.Ping(); err == nil {
			pgStatus = "connected"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "healthy",
		"uptime":     time.Since(e.startTime).String(),
		"clickhouse": chStatus,
		"postgres":   pgStatus,
		"rules":      len(e.rules),
		"alerts":     len(e.memAlerts),
	})
}

func (e *DetectionEngine) configHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"clickhouse": gin.H{
			"host":     e.config.CHHost,
			"port":     e.config.CHPort,
			"database": e.config.CHDatabase,
		},
		"postgres": gin.H{
			"host":     e.config.PGHost,
			"port":     e.config.PGPort,
			"database": e.config.PGDatabase,
		},
		"rules_path": e.config.RulesStoragePath,
	})
}

func (e *DetectionEngine) getRulesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"count": len(e.rules),
		"rules": e.rules,
	})
}

func (e *DetectionEngine) reloadRulesHandler(c *gin.Context) {
	if err := e.loadRules(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rules reloaded", "count": len(e.rules)})
}

// GET /alerts - Get alerts from PostgreSQL
func (e *DetectionEngine) getAlertsHandler(c *gin.Context) {
	limit := c.DefaultQuery("limit", "100")
	severity := c.Query("severity")

	// Try to get from PostgreSQL first
	if e.pgDB != nil {
		alerts, err := e.fetchAlertsFromPostgres(limit, severity)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"source": "postgres",
				"count":  len(alerts),
				"alerts": alerts,
			})
			return
		}
		e.writeLog(fmt.Sprintf("Failed to fetch from PG: %v", err))
	}

	// Fallback to memory
	filtered := e.memAlerts
	if severity != "" {
		filtered = make([]Alert, 0)
		for _, a := range e.memAlerts {
			if a.Severity == severity {
				filtered = append(filtered, a)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"source": "memory",
		"count":  len(filtered),
		"alerts": filtered,
	})
}

func (e *DetectionEngine) fetchAlertsFromPostgres(limit, severity string) ([]Alert, error) {
	query := "SELECT id, rule_id, rule_name, severity, message, log_id, matched_text, created_at FROM alerts"
	args := []interface{}{}

	if severity != "" {
		query += " WHERE severity = $1"
		args = append(args, severity)
	}

	query += " ORDER BY created_at DESC LIMIT " + limit

	rows, err := e.pgDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert
		var logID, matchedText sql.NullString
		if err := rows.Scan(&a.ID, &a.RuleID, &a.RuleName, &a.Severity, &a.Message, &logID, &matchedText, &a.CreatedAt); err != nil {
			continue
		}
		a.LogID = logID.String
		a.MatchedText = matchedText.String
		alerts = append(alerts, a)
	}

	return alerts, nil
}

func (e *DetectionEngine) getAlertsCountHandler(c *gin.Context) {
	count := 0

	if e.pgDB != nil {
		row := e.pgDB.QueryRow("SELECT COUNT(*) FROM alerts")
		row.Scan(&count)
	} else {
		count = len(e.memAlerts)
	}

	// Count by severity
	severityCounts := map[string]int{}
	if e.pgDB != nil {
		rows, _ := e.pgDB.Query("SELECT severity, COUNT(*) FROM alerts GROUP BY severity")
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var sev string
				var cnt int
				rows.Scan(&sev, &cnt)
				severityCounts[sev] = cnt
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":       count,
		"by_severity": severityCounts,
	})
}

// POST /logs - Insert test log to ClickHouse
func (e *DetectionEngine) insertLogHandler(c *gin.Context) {
	var req struct {
		Level   string `json:"level"`
		Message string `json:"message"`
		Source  string `json:"source"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if e.chConn == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ClickHouse not connected"})
		return
	}

	logID := uuid.New().String()
	query := fmt.Sprintf(`
		INSERT INTO %s.%s (id, level, message, source, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`, e.config.CHDatabase, e.config.LogTable)

	if err := e.chConn.Exec(context.Background(), query, logID, req.Level, req.Message, req.Source, time.Now()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Log inserted",
		"log_id":  logID,
	})
}

func (e *DetectionEngine) getLogsHandler(c *gin.Context) {
	if e.chConn == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ClickHouse not connected"})
		return
	}

	limit := c.DefaultQuery("limit", "100")
	query := fmt.Sprintf(`
		SELECT id, level, message, source, timestamp 
		FROM %s.%s 
		ORDER BY timestamp DESC 
		LIMIT %s
	`, e.config.CHDatabase, e.config.LogTable, limit)

	rows, err := e.chConn.Query(context.Background(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var log LogEntry
		if err := rows.Scan(&log.ID, &log.Level, &log.Message, &log.Source, &log.Timestamp); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(logs),
		"logs":  logs,
	})
}

func (e *DetectionEngine) triggerDetectionHandler(c *gin.Context) {
	go e.runDetection()
	c.JSON(http.StatusOK, gin.H{"message": "Detection triggered"})
}

func (e *DetectionEngine) queryClickHouseHandler(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if e.chConn == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ClickHouse not connected"})
		return
	}

	rows, err := e.chConn.Query(context.Background(), req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	columns := rows.Columns()

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	c.JSON(http.StatusOK, gin.H{"columns": columns, "rows": results, "count": len(results)})
}

func (e *DetectionEngine) queryPostgresHandler(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if e.pgDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not connected"})
		return
	}

	rows, err := e.pgDB.Query(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	c.JSON(http.StatusOK, gin.H{"columns": columns, "rows": results, "count": len(results)})
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
