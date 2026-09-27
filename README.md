# CloudOps Observability Platform

An all-in-one, zero-configuration observability engine built in Go for small teams without dedicated DevOps engineers.

Instead of managing Prometheus, Grafana, Loki, Tempo, and OpenTelemetry collectors, **CloudOps Cockpit** provides a single compiled binary that receives OTLP telemetry (metrics, logs, and traces), stores them in persistent SQLite, and serves a built-in dashboard.

---

## How Applications Send Data to CloudOps Cockpit

CloudOps Cockpit exposes standard OpenTelemetry (OTLP) HTTP JSON endpoints out of the box (`/v1/traces`, `/v1/metrics`, `/v1/logs`). Teams can send monitoring data in three ways:

### 1. Standard OpenTelemetry Environment Variables (Zero Code Changes)
If your application already uses OpenTelemetry SDKs (Go, Node.js, Python, Java, C#, etc.), simply set these environment variables where your app runs:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:8080"
export OTEL_SERVICE_NAME="my-web-app"
```
The standard OTel SDK automatically packages and streams your traces, metrics, and logs to CloudOps Cockpit.

### 2. OpenTelemetry Collector / Agent
For containerized or Kubernetes environments, run an OpenTelemetry Collector and configure its exporter to point to your CloudOps instance:

```yaml
exporters:
  otlphttp:
    endpoint: "http://localhost:8080"

service:
  pipelines:
    traces:
      exporters: [otlphttp]
    metrics:
      exporters: [otlphttp]
    logs:
      exporters: [otlphttp]
```

### 3. Direct HTTP POST (Easiest for Custom Apps & Scripts)
You can push JSON payloads directly from your application code using standard HTTP requests:

```powershell
# Push a Trace Span
Invoke-RestMethod -Uri "http://localhost:8080/v1/traces" -Method Post -ContentType "application/json" -Body '{"service_name":"checkout-svc","operation_name":"POST /pay","duration_ms":142,"status_code":200}'

# Push a Metric Point
Invoke-RestMethod -Uri "http://localhost:8080/v1/metrics" -Method Post -ContentType "application/json" -Body '{"service_name":"checkout-svc","metric_name":"orders_processed","value":15}'

# Push a Log Entry
Invoke-RestMethod -Uri "http://localhost:8080/v1/logs" -Method Post -ContentType "application/json" -Body '{"service_name":"checkout-svc","level":"ERROR","message":"Database timeout during transaction"}'
```

---

## Quick Start

1. Build the binary:
   ```bash
   go build -o cloudops-cockpit main.go
   ```
2. Run it:
   ```bash
   ./cloudops-cockpit
   ```
3. Open `http://localhost:8080`.
