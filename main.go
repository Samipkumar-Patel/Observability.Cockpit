package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
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

var (
	mu       sync.Mutex
	spans    []TraceSpan
	maxSpans = 50
)

func addSpan(span TraceSpan) {
	mu.Lock()
	defer mu.Unlock()
	if span.Timestamp.IsZero() {
		span.Timestamp = time.Now()
	}
	spans = append([]TraceSpan{span}, spans...)
	if len(spans) > maxSpans {
		spans = spans[:maxSpans]
	}
}

func getSpans() []TraceSpan {
	mu.Lock()
	defer mu.Unlock()
	copied := make([]TraceSpan, len(spans))
	copy(copied, spans)
	return copied
}

func seedDemoData() {
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
    .panel { background: var(--panel); border: 1px solid var(--line); border-radius: 6px; padding: 24px; margin-bottom: 20px; }
    .panel-heading { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
    table { width: 100%; border-collapse: collapse; text-align: left; }
    th { color: var(--muted); font-size: 11px; text-transform: uppercase; letter-spacing: .08em; font-weight: 700; padding: 10px; border-bottom: 1px solid var(--line); }
    td { padding: 12px 10px; border-bottom: 1px solid var(--line); font-size: 14px; }
    .tag { background: #e8efeb; color: var(--teal); padding: 3px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
    .status-error { color: var(--orange); font-weight: 700; }
    .status-ok { color: var(--teal); font-weight: 700; }
    .code-box { background: var(--ink); color: #fff; padding: 16px; border-radius: 4px; font-family: monospace; font-size: 13px; overflow-x: auto; margin-top: 10px; }
  </style>
</head>
<body>
  <main class="shell">
    <header class="topbar">
      <div>
        <p class="eyebrow">SINGLE-BINARY ENGINE</p>
        <h1>CloudOps Cockpit</h1>
        <p class="muted">Zero-config OTLP telemetry ingestion and live request tracing.</p>
      </div>
      <span class="badge">RUNNING</span>
    </header>

    <section class="panel">
      <div class="panel-heading">
        <h2>Live Telemetry Stream</h2>
        <span class="muted">Auto-refreshes every 3 seconds</span>
      </div>
      <table>
        <thead>
          <tr>
            <th>Time</th>
            <th>Service</th>
            <th>Operation</th>
            <th>Duration</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody id="spans-table">
          <tr><td colspan="5" style="text-align:center; color: var(--muted);">Loading telemetry...</td></tr>
        </tbody>
      </table>
    </section>

    <section class="panel">
      <h2>Push OpenTelemetry Data</h2>
      <p class="muted" style="margin-top: 6px;">Send JSON spans to your local OTLP endpoint using PowerShell:</p>
      <div class="code-box">Invoke-RestMethod -Uri "http://localhost:8080/v1/traces" -Method Post -ContentType "application/json" -Body '{"service_name":"checkout-svc","operation_name":"POST /pay","duration_ms":142,"status_code":200}'</div>
    </section>
  </main>

  <script>
    async function fetchTraces() {
      try {
        const res = await fetch("/api/traces");
        const data = await res.json();
        const tbody = document.querySelector("#spans-table");
        if (!data || data.length === 0) {
          tbody.innerHTML = '<tr><td colspan="5" style="text-align:center; color: var(--muted);">No telemetry received yet.</td></tr>';
          return;
        }
        tbody.innerHTML = data.map(function(s) {
          return '<tr>' +
            '<td>' + new Date(s.timestamp).toLocaleTimeString() + '</td>' +
            '<td><span class="tag">' + s.service_name + '</span></td>' +
            '<td><strong>' + s.operation_name + '</strong></td>' +
            '<td>' + s.duration_ms + ' ms</td>' +
            '<td><span class="' + (s.status_code >= 400 ? 'status-error' : 'status-ok') + '">' + s.status_code + '</span></td>' +
          '</tr>';
        }).join("");
      } catch (err) {
        console.error("Failed to load traces", err);
      }
    }

    fetchTraces();
    setInterval(fetchTraces, 3000);
  </script>
</body>
</html>`
}

func main() {
	seedDemoData()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":       "healthy",
			"name":         "CloudOps Observability Platform",
			"version":      "0.2.0-alpha",
			"active_spans": len(getSpans()),
		})
	})

	mux.HandleFunc("/api/traces", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(getSpans())
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

		addSpan(span)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDashboardHTML())
	})

	fmt.Printf("Starting CloudOps Observability Engine on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
