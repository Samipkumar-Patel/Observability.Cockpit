package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

type Config struct {
	Port       string `json:"port"`
	DBPath     string `json:"db_path"`
	MaxRecords int    `json:"max_records"`
}

var cfg Config

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
	addMetric(MetricPoint{
		ServiceName: "payment-api",
		MetricName:  "http_requests_total",
		Value:       435,
		Timestamp:   time.Now().Add(-5 * time.Minute),
	})

	addLog(LogEntry{
		ServiceName: "auth-service",
		Level:       "INFO",
		Message:     "User authenticated successfully for admin@example.com",
		Timestamp:   time.Now().Add(-8 * time.Second),
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
      --teal: #147d76; --orange: #d66c3c; --hover: #f1f5f3;
    }
    [data-theme="dark"] {
      --ink: #f0f4f8; --muted: #9aa5b1; --line: #2d3748; --paper: #111822; --panel: #1a2332;
      --teal: #319795; --orange: #ed8936; --hover: #222d3f;
    }
    * { box-sizing: border-box; }
    body { margin: 0; color: var(--ink); background: var(--paper); font: 15px/1.5 system-ui, sans-serif; transition: background 0.2s, color 0.2s; }
    .shell { max-width: 1280px; margin: 0 auto; padding: 40px 20px; }
    .topbar { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 24px; }
    h1, h2, p { margin: 0; } h1 { font-size: 32px; letter-spacing: -.03em; }
    .eyebrow { color: var(--teal); font-size: 11px; font-weight: 800; letter-spacing: .12em; margin-bottom: 6px; }
    .badge-group { display: flex; gap: 10px; align-items: center; }
    .badge { background: var(--panel); border: 1px solid var(--line); color: var(--teal); border-radius: 99px; padding: 6px 14px; font-size: 11px; font-weight: 800; }
    .theme-btn { background: var(--panel); border: 1px solid var(--line); color: var(--ink); border-radius: 99px; padding: 6px 14px; font-size: 11px; font-weight: 800; cursor: pointer; }
    
    /* KPI Summary Cards */
    .kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px; }
    .kpi-card { background: var(--panel); border: 1px solid var(--line); border-radius: 6px; padding: 18px; }
    .kpi-card .label { color: var(--muted); font-size: 11px; text-transform: uppercase; font-weight: 700; letter-spacing: .08em; }
    .kpi-card .value { font-size: 26px; font-weight: 700; margin-top: 6px; color: var(--ink); }

    /* Search Bar Toolbar */
    .toolbar { background: var(--panel); border: 1px solid var(--line); border-radius: 6px; padding: 14px 20px; margin-bottom: 24px; display: flex; gap: 16px; align-items: center; }
    .search-input { flex: 1; padding: 10px 14px; border: 1px solid var(--line); border-radius: 4px; background: var(--paper); color: var(--ink); font: inherit; outline: none; }

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

    /* Modal */
    .modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.6); display: none; align-items: center; justify-content: center; z-index: 100; }
    .modal-overlay.active { display: flex; }
    .modal { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; width: 600px; max-width: 90%; padding: 24px; box-shadow: 0 10px 30px rgba(0,0,0,0.3); }
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
        <p class="muted">Unified OTLP Traces, Metrics, and Logs Console.</p>
      </div>
      <div class="badge-group">
        <button class="theme-btn" onclick="toggleTheme()">🌓 Theme</button>
        <span class="badge">LIVE ENGINE</span>
      </div>
    </header>

    <!-- KPI Summary Cards -->
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

    <!-- Search & Filter Toolbar -->
    <section class="toolbar">
      <input type="text" id="search-box" class="search-input" placeholder="🔍 Filter traces, metrics, or logs by service name or keyword..." oninput="filterData()">
    </section>

    <section class="grid">
      <article class="panel" style="margin-bottom:0;">
        <div class="panel-heading">
          <h2>Live Traces</h2>
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

      <article class="panel" style="margin-bottom:0;">
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
    let globalData = { traces: [], metrics: [], logs: [] };

    function toggleTheme() {
      const html = document.documentElement;
      const current = html.getAttribute("data-theme");
      const next = current === "dark" ? "light" : "dark";
      html.setAttribute("data-theme", next);
      localStorage.setItem("cloudops-theme", next);
    }

    const savedTheme = localStorage.getItem("cloudops-theme") || "light";
    document.documentElement.setAttribute("data-theme", savedTheme);

    async function fetchData() {
      try {
        const [tracesRes, metricsRes, logsRes] = await Promise.all([
          fetch("/api/traces"),
          fetch("/api/metrics"),
          fetch("/api/logs")
        ]);
        globalData.traces = await tracesRes.json() || [];
        globalData.metrics = await metricsRes.json() || [];
        globalData.logs = await logsRes.json() || [];
        
        updateKPIs();
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

    function filterData() {
      const q = (document.querySelector("#search-box").value || "").toLowerCase().trim();
      
      const filteredTraces = globalData.traces.filter(s => {
        if (!q) return true;
        return (s.service_name || "").toLowerCase().includes(q) || 
               (s.operation_name || "").toLowerCase().includes(q) ||
               String(s.status_code).includes(q);
      });

      const filteredMetrics = globalData.metrics.filter(m => {
        if (!q) return true;
        return (m.service_name || "").toLowerCase().includes(q) || 
               (m.metric_name || "").toLowerCase().includes(q) ||
               String(m.value).includes(q);
      });

      const filteredLogs = globalData.logs.filter(l => {
        if (!q) return true;
        return (l.service_name || "").toLowerCase().includes(q) || 
               (l.message || "").toLowerCase().includes(q) || 
               (l.level || "").toLowerCase().includes(q);
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

    function showDetails(title, obj) {
      document.querySelector("#modal-title").textContent = title + " Details";
      document.querySelector("#modal-content").textContent = JSON.stringify(obj, null, 2);
      document.querySelector("#detail-modal").classList.add("active");
    }

    function closeModal() {
      document.querySelector("#detail-modal").classList.remove("active");
    }

    fetchData();
    setInterval(fetchData, 3000);
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
			"version":        "0.7.1-fix-filter",
			"active_spans":   len(spans),
			"active_metrics": len(metrics),
			"active_logs":    len(logs),
		})
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
