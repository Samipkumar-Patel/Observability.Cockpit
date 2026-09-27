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

var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("sqlite", "telemetry.db")
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
	_, err = db.Exec(metricsTable)
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

func getSpans() ([]TraceSpan, error) {
	query := `SELECT trace_id, span_id, service_name, operation_name, duration_ms, status_code, timestamp FROM spans ORDER BY timestamp DESC LIMIT 50`
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
	query := `SELECT service_name, metric_name, value, timestamp FROM metrics ORDER BY timestamp DESC LIMIT 50`
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
}

func getDashboardHTML() string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>CloudOps Observability Platform</title>
  <style>
    :root { --ink: #17252b; --muted: #718087; --line: #dce5e3; --paper: #f5f7f4; --panel: #fff; --teal: #147d76; --orange: #d66c3c; }
    * { box-sizing: border-box; }
    body { margin: 0; color: var(--ink); background: var(--paper); font: 15px/1.5 system-ui, sans-serif; }
    .shell { max-width: 1100px; margin: 0 auto; padding: 40px 20px; }
    .topbar { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 30px; }
    h1, h2, p { margin: 0; } h1 { font-size: 32px; letter-spacing: -.03em; }
    .eyebrow { color: var(--teal); font-size: 11px; font-weight: 800; letter-spacing: .12em; margin-bottom: 6px; }
    .badge { background: #e2f0ea; color: var(--teal); border-radius: 99px; padding: 6px 12px; font-size: 11px; font-weight: 800; }
    .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 20px; }
    .panel { background: var(--panel); border: 1px solid var(--line); border-radius: 6px; padding: 24px; margin-bottom: 20px; }
    .panel-heading { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
    table { width: 100%; border-collapse: collapse; text-align: left; }
    th { color: var(--muted); font-size: 11px; text-transform: uppercase; letter-spacing: .08em; font-weight: 700; padding: 10px; border-bottom: 1px solid var(--line); }
    td { padding: 12px 10px; border-bottom: 1px solid var(--line); font-size: 14px; }
    .tag { background: #e8efeb; color: var(--teal); padding: 3px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
    .status-error { color: var(--orange); font-weight: 700; }
    .status-ok { color: var(--teal); font-weight: 700; }
    .code-box { background: var::ink; color: #fff; padding: 16px; border-radius: 4px; font-family: monospace; font-size: 13px; overflow-x: auto; margin-top: 10px; }
  </style>
</head>
<body>
  <main class="shell">
    <header class="topbar">
      <div>
        <p class="eyebrow">SINGLE-BINARY ENGINE + SQLITE</p>
        <h1>CloudOps Cockpit</h1>
        <p class="muted">Zero-config OTLP traces & metrics ingestion engine.</p>
      </div>
      <span class="badge">RUNNING</span>
    </header>

    <section class="grid">
      <article class="panel" style="margin-bottom:0;">
        <div class="panel-heading">
          <h2>Live Traces</h2>
          <span class="muted">Refreshes every 3s</span>
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
          <span class="muted">Counters & Gauges</span>
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

    <section class="panel">
      <h2>Push OpenTelemetry Data</h2>
      <p class="muted" style="margin-top: 6px;">Send Spans or Metrics using PowerShell:</p>
      <div class="code-box"># Push Trace Span
Invoke-RestMethod -Uri "http://localhost:8080/v1/traces" -Method Post -ContentType "application/json" -Body '{"service_name":"checkout-svc","operation_name":"POST /pay","duration_ms":142,"status_code":200}'

# Push Metric Point
Invoke-RestMethod -Uri "http://localhost:8080/v1/metrics" -Method Post -ContentType "application/json" -Body '{"service_name":"checkout-svc","metric_name":"orders_processed","value":15}'</div>
    </section>
  </main>

  <script>
    async function fetchData() {
      try {
        const [tracesRes, metricsRes] = await Promise.all([
          fetch("/api/traces"),
          fetch("/api/metrics")
        ]);
        const traces = await tracesRes.json();
        const metrics = await metricsRes.json();

        const spansTbody = document.querySelector("#spans-table");
        if (!traces || traces.length === 0) {
          spansTbody.innerHTML = '<tr><td colspan="4" style="text-align:center; color: var(--muted);">No traces yet.</td></tr>';
        } else {
          spansTbody.innerHTML = traces.slice(0, 10).map(function(s) {
            return '<tr>' +
              '<td><span class="tag">' + s.service_name + '</span></td>' +
              '<td><strong>' + s.operation_name + '</strong></td>' +
              '<td>' + s.duration_ms + ' ms</td>' +
              '<td><span class="' + (s.status_code >= 400 ? 'status-error' : 'status-ok') + '">' + s.status_code + '</span></td>' +
            '</tr>';
          }).join("");
        }

        const metricsTbody = document.querySelector("#metrics-table");
        if (!metrics || metrics.length === 0) {
          metricsTbody.innerHTML = '<tr><td colspan="3" style="text-align:center; color: var(--muted);">No metrics yet.</td></tr>';
        } else {
          metricsTbody.innerHTML = metrics.slice(0, 10).map(function(m) {
            return '<tr>' +
              '<td><span class="tag">' + m.service_name + '</span></td>' +
              '<td>' + m.metric_name + '</td>' +
              '<td><strong>' + m.value + '</strong></td>' +
            '</tr>';
          }).join("");
        }
      } catch (err) {
        console.error("Failed to load telemetry data", err);
      }
    }

    fetchData();
    setInterval(fetchData, 3000);
  </script>
</body>
</html>`
}

func main() {
	if err := initDB(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	seedDemoData()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		spans, _ := getSpans()
		metrics, _ := getMetrics()
		json.NewEncoder(w).Encode(map[string]any{
			"status":        "healthy",
			"name":          "CloudOps Observability Platform",
			"version":       "0.4.0-metrics",
			"active_spans":  len(spans),
			"active_metric": len(metrics),
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDashboardHTML())
	})

	fmt.Printf("Starting CloudOps Observability Engine (Metrics + Traces) on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
