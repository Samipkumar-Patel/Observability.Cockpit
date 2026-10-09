package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Config struct {
	Port       string            `json:"port"`
	DBPath     string            `json:"db_path"`
	MaxRecords int               `json:"max_records"`
	Radar      []SoftwareVersion `json:"radar"`
}

var cfg Config
var customRadar []SoftwareVersion
var radarMu sync.Mutex

func loadConfig() {
	cfg = Config{
		Port:       "8080",
		DBPath:     "telemetry.db",
		MaxRecords: 100,
	}

	if p := os.Getenv("PORT"); p != "" {
		cfg.Port = p
	}
	if dbp := os.Getenv("DB_PATH"); dbp != "" {
		cfg.DBPath = dbp
	}

	if file, err := os.Open("config.json"); err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		_ = decoder.Decode(&cfg)
	}

	if len(cfg.Radar) > 0 {
		customRadar = cfg.Radar
	}
}

type TraceSpan struct {
	TraceID       string    `json:"trace_id"`
	SpanID        string    `json:"span_id"`
	ServiceName   string    `json:"service_name"`
	OperationName string    `json:"operation_name"`
	DurationMs    int64     `json:"duration_ms"`
	StatusCode    int       `json:"status_code"`
	Timestamp     time.Time `json:"timestamp"`
}

type MetricPoint struct {
	ServiceName string    `json:"service_name"`
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	Timestamp   time.Time `json:"timestamp"`
}

type LogEntry struct {
	ServiceName string    `json:"service_name"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

type AIInsight struct {
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	ServiceName string `json:"service_name"`
	Description string `json:"description"`
}

type SoftwareVersion struct {
	Name           string `json:"name"`
	Category       string `json:"category"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Status         string `json:"status"`
	ReleaseNotes   string `json:"release_notes"`
	ReleasedAt     string `json:"released_at"`
}

var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return err
	}

	spansTable := `
	CREATE TABLE IF NOT EXISTS spans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		trace_id TEXT,
		span_id TEXT,
		service_name TEXT,
		operation_name TEXT,
		duration_ms INTEGER,
		status_code INTEGER,
		timestamp DATETIME
	);
	`
	if _, err := db.Exec(spansTable); err != nil {
		return err
	}

	metricsTable := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_name TEXT,
		metric_name TEXT,
		value REAL,
		timestamp DATETIME
	);
	`
	if _, err := db.Exec(metricsTable); err != nil {
		return err
	}

	logsTable := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_name TEXT,
		level TEXT,
		message TEXT,
		timestamp DATETIME
	);
	`
	_, err = db.Exec(logsTable)
	return err
}

func addSpan(span TraceSpan) error {
	if span.Timestamp.IsZero() {
		span.Timestamp = time.Now()
	}
	query := `INSERT INTO spans (trace_id, span_id, service_name, operation_name, duration_ms, status_code, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, span.TraceID, span.SpanID, span.ServiceName, span.OperationName, span.DurationMs, span.StatusCode, span.Timestamp)
	return err
}

func addMetric(m MetricPoint) error {
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}
	query := `INSERT INTO metrics (service_name, metric_name, value, timestamp) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, m.ServiceName, m.MetricName, m.Value, m.Timestamp)
	return err
}

func addLog(l LogEntry) error {
	if l.Timestamp.IsZero() {
		l.Timestamp = time.Now()
	}
	if l.Level == "" {
		l.Level = "INFO"
	}
	query := `INSERT INTO logs (service_name, level, message, timestamp) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, l.ServiceName, l.Level, l.Message, l.Timestamp)
	return err
}

func getSpans() ([]TraceSpan, error) {
	query := fmt.Sprintf(`SELECT trace_id, span_id, service_name, operation_name, duration_ms, status_code, timestamp FROM spans ORDER BY timestamp DESC LIMIT %d`, cfg.MaxRecords)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spans []TraceSpan
	for rows.Next() {
		var s TraceSpan
		var ts string
		if err := rows.Scan(&s.TraceID, &s.SpanID, &s.ServiceName, &s.OperationName, &s.DurationMs, &s.StatusCode, &ts); err != nil {
			continue
		}
		if t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", ts); err == nil {
			s.Timestamp = t
		} else if t, err := time.Parse(time.RFC3339, ts); err == nil {
			s.Timestamp = t
		} else {
			s.Timestamp = time.Now()
		}
		spans = append(spans, s)
	}
	return spans, nil
}

func getMetrics() ([]MetricPoint, error) {
	query := fmt.Sprintf(`SELECT service_name, metric_name, value, timestamp FROM metrics ORDER BY timestamp DESC LIMIT %d`, cfg.MaxRecords)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []MetricPoint
	for rows.Next() {
		var m MetricPoint
		var ts string
		if err := rows.Scan(&m.ServiceName, &m.MetricName, &m.Value, &ts); err != nil {
			continue
		}
		if t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", ts); err == nil {
			m.Timestamp = t
		} else if t, err := time.Parse(time.RFC3339, ts); err == nil {
			m.Timestamp = t
		} else {
			m.Timestamp = time.Now()
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func getLogs() ([]LogEntry, error) {
	query := fmt.Sprintf(`SELECT service_name, level, message, timestamp FROM logs ORDER BY timestamp DESC LIMIT %d`, cfg.MaxRecords)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var l LogEntry
		var ts string
		if err := rows.Scan(&l.ServiceName, &l.Level, &l.Message, &ts); err != nil {
			continue
		}
		if t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", ts); err == nil {
			l.Timestamp = t
		} else if t, err := time.Parse(time.RFC3339, ts); err == nil {
			l.Timestamp = t
		} else {
			l.Timestamp = time.Now()
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func getVersionRadar() []SoftwareVersion {
	radarMu.Lock()
	defer radarMu.Unlock()

	if len(customRadar) > 0 {
		return customRadar
	}

	return []SoftwareVersion{
		{
			Name:           "kube-prometheus-stack",
			Category:       "Monitoring",
			CurrentVersion: "v56.2.1",
			LatestVersion:  "v68.4.0",
			Status:         "upgrade-available",
			ReleaseNotes:   "Major CRD updates, Prometheus v3.0 support, and enhanced Kubernetes 1.32 compatibility.",
			ReleasedAt:     "2 days ago",
		},
		{
			Name:           "Grafana",
			Category:       "Dashboard & Visualization",
			CurrentVersion: "10.4.0",
			LatestVersion:  "11.4.0",
			Status:         "security-update",
			ReleaseNotes:   "Critical security patch for datasource permission handling, new panel visualizations, and improved trace correlation.",
			ReleasedAt:     "5 days ago",
		},
		{
			Name:           "Loki",
			Category:       "Log Aggregation",
			CurrentVersion: "2.9.3",
			LatestVersion:  "3.3.2",
			Status:         "upgrade-available",
			ReleaseNotes:   "Performance optimizations for chunk caching, reduced memory footprint, and native structured metadata queries.",
			ReleasedAt:     "1 week ago",
		},
		{
			Name:           "Tempo",
			Category:       "Distributed Tracing",
			CurrentVersion: "2.4.1",
			LatestVersion:  "2.6.1",
			Status:         "upgrade-available",
			ReleaseNotes:   "Improved block compaction speeds, lower CPU utilization during high ingestion loads, and OTLP native metrics export.",
			ReleasedAt:     "2 weeks ago",
		},
		{
			Name:           "n8n",
			Category:       "Workflow Automation",
			CurrentVersion: "1.38.2",
			LatestVersion:  "1.75.1",
			Status:         "upgrade-available",
			ReleaseNotes:   "Advanced AI agent node integrations, improved execution error handling, and faster workflow execution engine.",
			ReleasedAt:     "3 days ago",
		},
		{
			Name:           "Mimir",
			Category:       "Metrics Long-term Storage",
			CurrentVersion: "2.11.0",
			LatestVersion:  "2.14.0",
			Status:         "up-to-date",
			ReleaseNotes:   "Hadoop/S3 multi-tenant storage enhancements and reduced TSDB indexing overhead.",
			ReleasedAt:     "3 weeks ago",
		},
	}
}

func upsertRadarVersion(sv SoftwareVersion) {
	radarMu.Lock()
	defer radarMu.Unlock()

	// If empty, initialize with default radar first
	if len(customRadar) == 0 {
		customRadar = getVersionRadar()
	}

	found := false
	for i, item := range customRadar {
		if item.Name == sv.Name {
			if sv.CurrentVersion != "" {
				customRadar[i].CurrentVersion = sv.CurrentVersion
			}
			if sv.LatestVersion != "" {
				customRadar[i].LatestVersion = sv.LatestVersion
			}
			if sv.Category != "" {
				customRadar[i].Category = sv.Category
			}
			if sv.Status != "" {
				customRadar[i].Status = sv.Status
			}
			if sv.ReleaseNotes != "" {
				customRadar[i].ReleaseNotes = sv.ReleaseNotes
			}
			if sv.ReleasedAt != "" {
				customRadar[i].ReleasedAt = sv.ReleasedAt
			}
			found = true
			break
		}
	}

	if !found {
		customRadar = append(customRadar, sv)
	}
}

func generateAIInsights() []AIInsight {
	var insights []AIInsight
	spans, _ := getSpans()

	serviceStats := make(map[string]struct {
		total  int
		errors int
		maxDur int64
	})

	for _, s := range spans {
		stats := serviceStats[s.ServiceName]
		stats.total++
		if s.StatusCode >= 400 {
			stats.errors++
		}
		if s.DurationMs > stats.maxDur {
			stats.maxDur = s.DurationMs
		}
		serviceStats[s.ServiceName] = stats
	}

	for svc, stats := range serviceStats {
		if stats.total > 0 && float64(stats.errors)/float64(stats.total) >= 0.2 {
			insights = append(insights, AIInsight{
				Severity:    "critical",
				ServiceName: svc,
				Title:       fmt.Sprintf("High Error Rate in %s", svc),
				Description: fmt.Sprintf("AI detected an elevated error rate of %.1f%% (%d/%d requests failed).", (float64(stats.errors)/float64(stats.total))*100, stats.errors, stats.total),
			})
		}
	}

	if len(insights) == 0 {
		insights = append(insights, AIInsight{
			Severity:    "info",
			ServiceName: "Cluster",
			Title:       "All Systems Operating Normally",
			Description: "AI telemetry analysis shows stable latencies and nominal error rates.",
		})
	}
	return insights
}

func seedDemoData() {
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM spans`).Scan(&count)
	if count > 0 {
		return
	}

	addSpan(TraceSpan{
		TraceID:       "a1b2c3d4e5f6",
		SpanID:        "11223344",
		ServiceName:   "auth-service",
		OperationName: "POST /api/login",
		DurationMs:    45,
		StatusCode:    200,
		Timestamp:     time.Now().Add(-10 * time.Second),
	})
	addSpan(TraceSpan{
		TraceID:       "f6e5d4c3b2a1",
		SpanID:        "55667788",
		ServiceName:   "payment-api",
		OperationName: "GET /api/checkout",
		DurationMs:    210,
		StatusCode:    500,
		Timestamp:     time.Now().Add(-35 * time.Second),
	})

	addMetric(MetricPoint{
		ServiceName: "auth-service",
		MetricName:  "http_requests_total",
		Value:       1280,
		Timestamp:   time.Now().Add(-5 * time.Minute),
	})

	addLog(LogEntry{
		ServiceName: "payment-api",
		Level:       "ERROR",
		Message:     "Database connection timeout while executing checkout transaction",
		Timestamp:   time.Now().Add(-30 * time.Second),
	})
}

func getDashboardHTML() string {
	return `<!doctype html>
<html lang="en" data-theme="light">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>CloudOps Observability Platform</title>
  <style>
    :root {
      --ink: #17252b; --muted: #718087; --line: #dce5e3; --paper: #f5f7f4; --panel: #fff;
      --teal: #147d76; --orange: #d66c3c; --hover: #f1f5f3; --btn-bg: #e8efeb;
      --ai-bg: #e2f0ea; --ai-border: #b8d9cc; --ai-text: #0f514c;
    }
    [data-theme="dark"] {
      --ink: #f0f4f8; --muted: #9aa5b1; --line: #2d3748; --paper: #111822; --panel: #1a2332;
      --teal: #319795; --orange: #ed8936; --hover: #222d3f; --btn-bg: #222d3f;
      --ai-bg: #162c28; --ai-border: #234e48; --ai-text: #4fd1c5;
    }
    * { box-sizing: border-box; }
    body { margin: 0; color: var(--ink); background: var(--paper); font: 15px/1.5 system-ui, sans-serif; transition: background 0.2s, color 0.2s; }
    .shell { max-width: 1280px; margin: 0 auto; padding: 40px 20px; }
    .topbar { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 24px; }
    h1, h2, p { margin: 0; } h1 { font-size: 32px; letter-spacing: -.03em; }
    .eyebrow { color: var(--teal); font-size: 11px; font-weight: 800; letter-spacing: .12em; margin-bottom: 6px; }
    
    /* Navigation Bar */
    .nav-bar { display: flex; gap: 10px; margin-bottom: 24px; border-bottom: 1px solid var(--line); padding-bottom: 12px; }
    .nav-btn { background: var(--panel); border: 1px solid var(--line); color: var(--muted); border-radius: 6px; padding: 10px 18px; font-size: 13px; font-weight: 700; cursor: pointer; transition: all 0.2s; }
    .nav-btn.active { background: var(--teal); color: #fff; border-color: var(--teal); }

    .badge-group { display: flex; gap: 10px; align-items: center; }
    .badge { background: var(--panel); border: 1px solid var(--line); color: var(--teal); border-radius: 99px; padding: 6px 14px; font-size: 11px; font-weight: 800; }
    .btn { background: var(--btn-bg); border: 1px solid var(--line); color: var(--ink); border-radius: 99px; padding: 6px 14px; font-size: 11px; font-weight: 800; cursor: pointer; display: inline-flex; align-items: center; gap: 6px; }
    .btn-primary { background: var(--teal); color: #fff; border-color: var(--teal); }

    /* Page Sections */
    .page-section { display: none; }
    .page-section.active { display: block; }

    /* AI Insights Panel */
    .ai-panel { background: var(--ai-bg); border: 1px solid var(--ai-border); border-radius: 6px; padding: 20px; margin-bottom: 24px; }
    .ai-header { display: flex; align-items: center; gap: 8px; font-weight: 800; color: var(--ai-text); font-size: 14px; margin-bottom: 12px; letter-spacing: .06em; text-transform: uppercase; }
    .ai-insights-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 12px; }
    .ai-card { background: var(--panel); border: 1px solid var(--line); border-radius: 6px; padding: 14px; }
    .ai-card.critical { border-left: 4px solid var(--orange); }
    .ai-card.warning { border-left: 4px solid #d69e2e; }
    .ai-card.info { border-left: 4px solid var(--teal); }
    .ai-card-title { font-weight: 700; font-size: 14px; margin-bottom: 4px; display: flex; justify-content: space-between; }
    .ai-card-desc { color: var(--muted); font-size: 13px; }

    /* KPI Summary Cards */
    .kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px; }
    .kpi-card { background: var(--panel); border: 1px solid var(--line); border-radius: 6px; padding: 18px; }
    .kpi-card .label { color: var(--muted); font-size: 11px; text-transform: uppercase; font-weight: 700; letter-spacing: .08em; }
    .kpi-card .value { font-size: 26px; font-weight: 700; margin-top: 6px; color: var(--ink); }

    /* Toolbar */
    .toolbar { background: var(--panel); border: 1px solid var(--line); border-radius: 6px; padding: 14px 20px; margin-bottom: 24px; display: flex; gap: 16px; align-items: center; flex-wrap: wrap; }
    .search-input { flex: 1; min-width: 240px; padding: 10px 14px; border: 1px solid var(--line); border-radius: 4px; background: var(--paper); color: var(--ink); font: inherit; outline: none; }
    .controls { display: flex; gap: 12px; align-items: center; }

    /* Chart Panel */
    .chart-panel { background: var(--panel); border: 1px solid var(--line); border-radius: 6px; padding: 20px 24px; margin-bottom: 24px; }
    .chart-bars { display: flex; align-items: flex-end; gap: 12px; height: 110px; border-bottom: 1px solid var(--line); padding-bottom: 4px; margin-top: 14px; overflow-x: auto; }
    .chart-bar-wrap { flex: 1; min-width: 32px; display: flex; flex-direction: column; align-items: center; height: 100%; justify-content: flex-end; }
    .chart-bar { width: 100%; background: var(--teal); border-radius: 3px 3px 0 0; transition: height 0.3s; }
    .chart-bar.error { background: var(--orange); }
    .chart-label { font-size: 10px; color: var(--muted); margin-top: 6px; white-space: nowrap; }

    .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 20px; }
    .panel { background: var(--panel); border: 1px solid var(--line); border-radius: 6px; padding: 24px; margin-bottom: 20px; }
    .panel-heading { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
    table { width: 100%; border-collapse: collapse; text-align: left; }
    th { color: var(--muted); font-size: 11px; text-transform: uppercase; letter-spacing: .08em; font-weight: 700; padding: 10px; border-bottom: 1px solid var(--line); }
    td { padding: 12px 10px; border-bottom: 1px solid var(--line); font-size: 14px; cursor: pointer; }
    tbody tr:hover { background: var(--hover); }
    .tag { background: var(--paper); border: 1px solid var(--line); color: var(--teal); padding: 3px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
    .status-error { color: var(--orange); font-weight: 700; }
    .status-ok { color: var(--teal); font-weight: 700; }
    .log-error { color: var(--orange); font-weight: 700; }
    .log-warn { color: #d69e2e; font-weight: 700; }
    .log-info { color: #3182ce; font-weight: 700; }
    
    .badge-status-up { background: #e2f0ea; color: #147d76; padding: 3px 8px; border-radius: 4px; font-size: 11px; font-weight: 700; }
    .badge-status-upgrade { background: #fcebdd; color: #d66c3c; padding: 3px 8px; border-radius: 4px; font-size: 11px; font-weight: 700; }
    .badge-status-security { background: #fed7d7; color: #c53030; padding: 3px 8px; border-radius: 4px; font-size: 11px; font-weight: 700; }

    /* Modal */
    .modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.6); display: none; align-items: center; justify-content: center; z-index: 100; }
    .modal-overlay.active { display: flex; }
    .modal { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; width: 650px; max-width: 90%; padding: 24px; box-shadow: 0 10px 30px rgba(0,0,0,0.3); }
    .modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
    .modal-close { background: none; border: none; font-size: 20px; color: var(--ink); cursor: pointer; }
    .code-view { background: var(--paper); border: 1px solid var(--line); color: var(--ink); padding: 14px; border-radius: 4px; font-family: monospace; font-size: 13px; white-space: pre-wrap; word-break: break-all; max-height: 350px; overflow-y: auto; }
    
    @media (max-width: 900px) { .grid { grid-template-columns: 1fr; } .kpi-grid { grid-template-columns: repeat(2, 1fr); } }
  </style>
</head>
<body>
  <main class="shell">
    <header class="topbar">
      <div>
        <p class="eyebrow">ENTERPRISE OBSERVABILITY ENGINE</p>
        <h1>CloudOps Cockpit</h1>
        <p class="muted">Unified OTLP Traces, Metrics, Logs, & Version Radar Console.</p>
      </div>
      <div class="badge-group">
        <button class="btn" onclick="toggleTheme()">🌓 Theme</button>
        <span class="badge">PRO EDITION</span>
      </div>
    </header>

    <!-- Navigation Bar -->
    <nav class="nav-bar">
      <button class="nav-btn active" onclick="switchPage('radar', event)">🛰️ Version Radar</button>
      <button class="nav-btn" onclick="switchPage('traces', event)">📊 Live Traces & Latency</button>
      <button class="nav-btn" onclick="switchPage('metrics', event)">📈 Metrics Stream</button>
      <button class="nav-btn" onclick="switchPage('logs', event)">📝 Structured Logs</button>
      <button class="nav-btn" onclick="switchPage('ai', event)">🤖 AI Assistant</button>
    </nav>

    <!-- PAGE 1: VERSION RADAR (DEFAULT) -->
    <section id="page-radar" class="page-section active">
      <article class="panel">
        <div class="panel-heading">
          <h2>Open-Source Version Radar & Upgrade Assistant</h2>
          <span class="muted">Tracking Helm charts, images & community tools</span>
        </div>
        <table>
          <thead>
            <tr><th>Software / Stack</th><th>Category</th><th>Current</th><th>Latest</th><th>Status</th><th>Action</th></tr>
          </thead>
          <tbody id="radar-table">
            <tr><td colspan="6" style="text-align:center; color: var(--muted);">Loading version radar...</td></tr>
          </tbody>
        </table>
      </article>
    </section>

    <!-- PAGE 2: LIVE TRACES -->
    <section id="page-traces" class="page-section">
      <section class="toolbar">
        <input type="text" id="search-traces" class="search-input" placeholder="🔍 Filter traces by service name or operation..." oninput="filterData()">
        <div class="controls">
          <button class="btn btn-primary" onclick="simulateTelemetry()">🧪 Test Telemetry</button>
        </div>
      </section>
      <section class="chart-panel">
        <div class="panel-heading" style="margin-bottom:0;">
          <h2>Trace Latency Distribution</h2>
          <span class="muted">Last requests (ms)</span>
        </div>
        <div class="chart-bars" id="latency-chart">
          <span class="muted" style="margin: auto; font-size: 13px;">No trace latency data available</span>
        </div>
      </section>
      <article class="panel">
        <div class="panel-heading">
          <h2>Live Traces Stream</h2>
          <span class="muted" id="trace-count">0 items</span>
        </div>
        <table>
          <thead>
            <tr><th>Service</th><th>Operation</th><th>Duration</th><th>Status</th></tr>
          </thead>
          <tbody id="spans-table">
            <tr><td colspan="4" style="text-align:center; color: var(--muted);">Loading...</td></tr>
          </tbody>
        </table>
      </article>
    </section>

    <!-- PAGE 3: METRICS STREAM -->
    <section id="page-metrics" class="page-section">
      <section class="toolbar">
        <input type="text" id="search-metrics" class="search-input" placeholder="🔍 Filter metrics by service name or metric name..." oninput="filterData()">
      </section>
      <article class="panel">
        <div class="panel-heading">
          <h2>Metrics Stream</h2>
          <span class="muted" id="metric-count">0 items</span>
        </div>
        <table>
          <thead>
            <tr><th>Service</th><th>Metric</th><th>Value</th></tr>
          </thead>
          <tbody id="metrics-table">
            <tr><td colspan="3" style="text-align:center; color: var(--muted);">Loading...</td></tr>
          </tbody>
        </table>
      </article>
    </section>

    <!-- PAGE 4: STRUCTURED LOGS -->
    <section id="page-logs" class="page-section">
      <section class="toolbar">
        <input type="text" id="search-logs" class="search-input" placeholder="🔍 Filter logs by service name, level, or keyword..." oninput="filterData()">
      </section>
      <article class="panel">
        <div class="panel-heading">
          <h2>Structured Logs</h2>
          <span class="muted" id="log-count">0 items</span>
        </div>
        <table>
          <thead>
            <tr><th>Time</th><th>Service</th><th>Level</th><th>Message</th></tr>
          </thead>
          <tbody id="logs-table">
            <tr><td colspan="4" style="text-align:center; color: var(--muted);">Loading logs...</td></tr>
          </tbody>
        </table>
      </article>
    </section>

    <!-- PAGE 5: AI ASSISTANT -->
    <section id="page-ai" class="page-section">
      <section class="ai-panel">
        <div class="ai-header">🤖 AI Observability Assistant & Root-Cause Analyzer</div>
        <div class="ai-insights-grid" id="ai-insights-container">
          <div class="ai-card info">
            <div class="ai-card-title">Analyzing Telemetry...</div>
            <div class="ai-card-desc">AI engine is evaluating traces, logs, and error rates.</div>
          </div>
        </div>
      </section>

      <section class="kpi-grid">
        <div class="kpi-card">
          <div class="label">Total Traces</div>
          <div class="value" id="kpi-traces">0</div>
        </div>
        <div class="kpi-card">
          <div class="label">Error Rate</div>
          <div class="value" id="kpi-errors">0%</div>
        </div>
        <div class="kpi-card">
          <div class="label">Active Services</div>
          <div class="value" id="kpi-services">0</div>
        </div>
        <div class="kpi-card">
          <div class="label">Total Logs</div>
          <div class="value" id="kpi-logs">0</div>
        </div>
      </section>
    </section>

  </main>

  <!-- Interactive Detail Modal -->
  <div class="modal-overlay" id="detail-modal" onclick="closeModal(event)">
    <div class="modal" onclick="event.stopPropagation()">
      <div class="modal-header">
        <h3 id="modal-title">Telemetry Details</h3>
        <button class="modal-close" onclick="closeModal()">×</button>
      </div>
      <div class="code-view" id="modal-content"></div>
    </div>
  </div>

  <script>
    let globalData = { traces: [], metrics: [], logs: [], insights: [], radar: [] };
    let refreshTimer = null;

    function toggleTheme() {
      const html = document.documentElement;
      const current = html.getAttribute("data-theme");
      const next = current === "dark" ? "light" : "dark";
      html.setAttribute("data-theme", next);
      localStorage.setItem("cloudops-theme", next);
    }

    const savedTheme = localStorage.getItem("cloudops-theme") || "light";
    document.documentElement.setAttribute("data-theme", savedTheme);

    function switchPage(pageId, evt) {
      document.querySelectorAll(".page-section").forEach(function(sec) {
        sec.classList.remove("active");
      });
      document.querySelectorAll(".nav-btn").forEach(function(btn) {
        btn.classList.remove("active");
      });
      document.querySelector("#page-" + pageId).classList.add("active");
      if (evt && evt.currentTarget) {
        evt.currentTarget.classList.add("active");
      }
    }

    async function fetchData() {
      try {
        const [tracesRes, metricsRes, logsRes, insightsRes, radarRes] = await Promise.all([
          fetch("/api/traces"),
          fetch("/api/metrics"),
          fetch("/api/logs"),
          fetch("/api/insights"),
          fetch("/api/radar")
        ]);
        globalData.traces = await tracesRes.json() || [];
        globalData.metrics = await metricsRes.json() || [];
        globalData.logs = await logsRes.json() || [];
        globalData.insights = await insightsRes.json() || [];
        globalData.radar = await radarRes.json() || [];
        
        updateKPIs();
        renderAIInsights();
        renderRadar();
        renderChart();
        filterData();
      } catch (err) {
        console.error("Failed to load telemetry data", err);
      }
    }

    function updateKPIs() {
      const totalTraces = globalData.traces.length;
      const errorTraces = globalData.traces.filter(s => s.status_code >= 400).length;
      const errorRate = totalTraces > 0 ? ((errorTraces / totalTraces) * 100).toFixed(1) + "%" : "0%";
      
      const services = new Set([
        ...globalData.traces.map(s => s.service_name),
        ...globalData.metrics.map(m => m.service_name),
        ...globalData.logs.map(l => l.service_name)
      ]);

      document.querySelector("#kpi-traces").textContent = totalTraces;
      document.querySelector("#kpi-errors").textContent = errorRate;
      document.querySelector("#kpi-services").textContent = services.size;
      document.querySelector("#kpi-logs").textContent = globalData.logs.length;
    }

    function renderAIInsights() {
      const container = document.querySelector("#ai-insights-container");
      if (!globalData.insights || globalData.insights.length === 0) {
        container.innerHTML = '<div class="ai-card info"><div class="ai-card-title">All Systems Normal</div><div class="ai-card-desc">No anomalies detected by AI.</div></div>';
        return;
      }
      container.innerHTML = globalData.insights.map(function(i) {
        return '<div class="ai-card ' + i.severity + '">' +
          '<div class="ai-card-title"><span>' + i.title + '</span> <span class="tag">' + i.service_name + '</span></div>' +
          '<div class="ai-card-desc">' + i.description + '</div>' +
        '</div>';
      }).join("");
    }

    function renderRadar() {
      const tbody = document.querySelector("#radar-table");
      if (!globalData.radar || globalData.radar.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center; color: var(--muted);">No software versions tracked.</td></tr>';
        return;
      }
      tbody.innerHTML = globalData.radar.map(function(r) {
        let badgeHtml = '<span class="badge-status-up">Up to Date</span>';
        if (r.status === 'upgrade-available') {
          badgeHtml = '<span class="badge-status-upgrade">Upgrade Available</span>';
        } else if (r.status === 'security-update') {
          badgeHtml = '<span class="badge-status-security">Security Patch</span>';
        }
        return '<tr>' +
          '<td><strong>' + r.name + '</strong></td>' +
          '<td><span class="tag">' + r.category + '</span></td>' +
          '<td>' + r.current_version + '</td>' +
          '<td><strong>' + r.latest_version + '</strong></td>' +
          '<td>' + badgeHtml + '</td>' +
          '<td><button class="btn" onclick=\'showReleaseNotes(' + JSON.stringify(r) + ')\'>📖 Release Notes</button></td>' +
        '</tr>';
      }).join("");
    }

    function renderChart() {
      const chartContainer = document.querySelector("#latency-chart");
      const recentTraces = [...globalData.traces].reverse().slice(-15);
      if (recentTraces.length === 0) {
        chartContainer.innerHTML = '<span class="muted" style="margin: auto; font-size: 13px;">No trace latency data available</span>';
        return;
      }
      const maxDuration = Math.max(...recentTraces.map(function(t) { return t.duration_ms; }), 50);
      chartContainer.innerHTML = recentTraces.map(function(t) {
        const heightPct = Math.max(Math.round((t.duration_ms / maxDuration) * 100), 8);
        const isError = t.status_code >= 400;
        return '<div class="chart-bar-wrap" title="' + t.service_name + ' | ' + t.operation_name + ' | ' + t.duration_ms + 'ms">' +
          '<div class="chart-bar ' + (isError ? 'error' : '') + '" style="height: ' + heightPct + '%"></div>' +
          '<span class="chart-label">' + t.duration_ms + 'ms</span>' +
        '</div>';
      }).join("");
    }

    function filterData() {
      const qTraces = (document.querySelector("#search-traces")?.value || "").toLowerCase().trim();
      const qMetrics = (document.querySelector("#search-metrics")?.value || "").toLowerCase().trim();
      const qLogs = (document.querySelector("#search-logs")?.value || "").toLowerCase().trim();
      
      const filteredTraces = globalData.traces.filter(function(s) {
        if (!qTraces) return true;
        return (s.service_name || "").toLowerCase().includes(qTraces) || (s.operation_name || "").toLowerCase().includes(qTraces);
      });

      const filteredMetrics = globalData.metrics.filter(function(m) {
        if (!qMetrics) return true;
        return (m.service_name || "").toLowerCase().includes(qMetrics) || (m.metric_name || "").toLowerCase().includes(qMetrics);
      });

      const filteredLogs = globalData.logs.filter(function(l) {
        if (!qLogs) return true;
        return (l.service_name || "").toLowerCase().includes(qLogs) || (l.message || "").toLowerCase().includes(qLogs) || (l.level || "").toLowerCase().includes(qLogs);
      });
      
      renderTables({ traces: filteredTraces, metrics: filteredMetrics, logs: filteredLogs });
    }

    function renderTables(data) {
      document.querySelector("#trace-count").textContent = data.traces.length + " items";
      document.querySelector("#metric-count").textContent = data.metrics.length + " items";
      document.querySelector("#log-count").textContent = data.logs.length + " items";

      const spansTbody = document.querySelector("#spans-table");
      if (data.traces.length === 0) {
        spansTbody.innerHTML = '<tr><td colspan="4" style="text-align:center; color: var(--muted);">No matching traces.</td></tr>';
      } else {
        spansTbody.innerHTML = data.traces.slice(0, 10).map(function(s) {
          return '<tr onclick=\'showDetails("Trace Span", ' + JSON.stringify(s) + ')\'>' +
            '<td><span class="tag">' + s.service_name + '</span></td>' +
            '<td><strong>' + s.operation_name + '</strong></td>' +
            '<td>' + s.duration_ms + ' ms</td>' +
            '<td><span class="' + (s.status_code >= 400 ? 'status-error' : 'status-ok') + '">' + s.status_code + '</span></td>' +
          '</tr>';
        }).join("");
      }

      const metricsTbody = document.querySelector("#metrics-table");
      if (data.metrics.length === 0) {
        metricsTbody.innerHTML = '<tr><td colspan="3" style="text-align:center; color: var(--muted);">No matching metrics.</td></tr>';
      } else {
        metricsTbody.innerHTML = data.metrics.slice(0, 10).map(function(m) {
          return '<tr onclick=\'showDetails("Metric Point", ' + JSON.stringify(m) + ')\'>' +
            '<td><span class="tag">' + m.service_name + '</span></td>' +
            '<td>' + m.metric_name + '</td>' +
            '<td><strong>' + m.value + '</strong></td>' +
          '</tr>';
        }).join("");
      }

      const logsTbody = document.querySelector("#logs-table");
      if (data.logs.length === 0) {
        logsTbody.innerHTML = '<tr><td colspan="4" style="text-align:center; color: var(--muted);">No matching logs.</td></tr>';
      } else {
        logsTbody.innerHTML = data.logs.slice(0, 10).map(function(l) {
          var lvlClass = l.level === 'ERROR' ? 'log-error' : (l.level === 'WARN' ? 'log-warn' : 'log-info');
          return '<tr onclick=\'showDetails("Log Entry", ' + JSON.stringify(l) + ')\'>' +
            '<td>' + new Date(l.timestamp).toLocaleTimeString() + '</td>' +
            '<td><span class="tag">' + l.service_name + '</span></td>' +
            '<td><span class="' + lvlClass + '">' + l.level + '</span></td>' +
            '<td>' + l.message + '</td>' +
          '</tr>';
        }).join("");
      }
    }

    async function simulateTelemetry() {
      const services = ["orders-api", "billing-svc", "users-service", "inventory-db"];
      const svc = services[Math.floor(Math.random() * services.length)];
      const isError = Math.random() < 0.25;

      try {
        await fetch("/v1/traces", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            service_name: svc,
            operation_name: isError ? "GET /api/fail" : "POST /api/process",
            duration_ms: Math.floor(Math.random() * 300) + 20,
            status_code: isError ? 500 : 200
          })
        });

        await fetch("/v1/metrics", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            service_name: svc,
            metric_name: "request_duration_ms",
            value: Math.floor(Math.random() * 250) + 10
          })
        });

        await fetch("/v1/logs", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            service_name: svc,
            level: isError ? "ERROR" : "INFO",
            message: isError ? "Simulated database connection timeout error" : "Simulated successful request processing"
          })
        });

        fetchData();
      } catch (err) {
        console.error("Simulation failed", err);
      }
    }

    function showReleaseNotes(r) {
      document.querySelector("#modal-title").textContent = r.name + " (" + r.latest_version + ") Release Notes";
      document.querySelector("#modal-content").textContent = "Category: " + r.category + "\nCurrent Version: " + r.current_version + "\nLatest Version: " + r.latest_version + "\nReleased: " + r.released_at + "\n\nRelease Highlights:\n" + r.release_notes;
      document.querySelector("#detail-modal").classList.add("active");
    }

    function showDetails(title, obj) {
      document.querySelector("#modal-title").textContent = title + " Details";
      document.querySelector("#modal-content").textContent = JSON.stringify(obj, null, 2);
      document.querySelector("#detail-modal").classList.add("active");
    }

    function closeModal() {
      document.querySelector("#detail-modal").classList.remove("active");
    }

    fetchData();
    refreshTimer = setInterval(fetchData, 3000);
  </script>
</body>
</html>`
}

func main() {
	loadConfig()

	if err := initDB(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	seedDemoData()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		spans, _ := getSpans()
		metrics, _ := getMetrics()
		logs, _ := getLogs()
		json.NewEncoder(w).Encode(map[string]any{
			"status":         "healthy",
			"name":           "CloudOps Observability Platform",
			"version":        "0.12.0-custom-radar",
			"active_spans":   len(spans),
			"active_metrics": len(metrics),
			"active_logs":    len(logs),
		})
	})

	mux.HandleFunc("/api/insights", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		insights := generateAIInsights()
		json.NewEncoder(w).Encode(insights)
	})

	mux.HandleFunc("/api/radar", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		radar := getVersionRadar()
		json.NewEncoder(w).Encode(radar)
	})

	mux.HandleFunc("/v1/radar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var sv SoftwareVersion
		if err := json.Unmarshal(body, &sv); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if sv.Name == "" {
			http.Error(w, "Software 'name' is required", http.StatusBadRequest)
			return
		}

		upsertRadarVersion(sv)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	})

	mux.HandleFunc("/api/traces", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		spans, err := getSpans()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(spans)
	})

	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		metrics, err := getMetrics()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(metrics)
	})

	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		logs, err := getLogs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(logs)
	})

	mux.HandleFunc("/v1/traces", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var span TraceSpan
		if err := json.Unmarshal(body, &span); err != nil {
			span = TraceSpan{
				TraceID:       fmt.Sprintf("otlp-%d", time.Now().UnixNano()),
				SpanID:        "otlp-span",
				ServiceName:   "external-app",
				OperationName: "OTLP Push",
				DurationMs:    12,
				StatusCode:    200,
				Timestamp:     time.Now(),
			}
		}

		if err := addSpan(span); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	})

	mux.HandleFunc("/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var m MetricPoint
		if err := json.Unmarshal(body, &m); err != nil {
			m = MetricPoint{
				ServiceName: "external-app",
				MetricName:  "custom_counter",
				Value:       1.0,
				Timestamp:   time.Now(),
			}
		}

		if err := addMetric(m); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	})

	mux.HandleFunc("/v1/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var l LogEntry
		if err := json.Unmarshal(body, &l); err != nil {
			l = LogEntry{
				ServiceName: "external-app",
				Level:       "INFO",
				Message:     "Sample log entry",
				Timestamp:   time.Now(),
			}
		}

		if err := addLog(l); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDashboardHTML())
	})

	fmt.Printf("Starting CloudOps Observability Platform on http://localhost:%s (db: %s)\n", cfg.Port, cfg.DBPath)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
