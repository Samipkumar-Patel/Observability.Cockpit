# CloudOps Observability Platform

[![GitHub License](https://img.shields.io/github/license/Samipkumar-Patel/Observability.Cockpit)](LICENSE)
[![GitHub Go Version](https://img.shields.io/github/go-mod/v/Samipkumar-Patel/Observability.Cockpit)](go.mod)
[![GitHub Stars](https://img.shields.io/github/stars/Samipkumar-Patel/Observability.Cockpit?style=social)](https://github.com/Samipkumar-Patel/Observability.Cockpit)
[![GHCR Container Image](https://img.shields.io/badge/container-ghcr.io%2Fsamipkumar--patel%2Fcloudops--cockpit-blue?logo=github)](https://ghcr.io/samipkumar-patel/cloudops-cockpit)

An all-in-one, zero-configuration observability engine built in Go for small teams without dedicated DevOps engineers.

Instead of managing Prometheus, Grafana, Loki, Tempo, and OpenTelemetry collectors, **CloudOps Cockpit** provides a single compiled binary that receives OTLP telemetry (metrics, logs, and traces), stores them in persistent SQLite, features an **AI Root-Cause Assistant** and **Open-Source Version Radar**, and serves a built-in multi-page dashboard.

---

## Key Features

1. **Unified OTLP Telemetry Ingestion:**
   - Native endpoints for Traces (`/v1/traces`), Metrics (`/v1/metrics`), and Structured Logs (`/v1/logs`).
2. **Persistent SQLite Storage:**
   - Lightweight, zero-dependency embedded database (`telemetry.db`) ensuring historical telemetry survives server restarts.
3. **Multi-Page Web Console:**
   - Clean, tabbed navigation separating **Version Radar**, **Live Traces & Latency**, **Metrics Stream**, **Structured Logs**, and **AI Assistant**.
4. **AI Observability Assistant & Root-Cause Analyzer:**
   - Automatically inspects stored telemetry for error-rate spikes, slow execution spans, and recurring log errors to generate actionable diagnostic cards.
5. **Open-Source Version Radar & Upgrade Assistant:**
   - Discovers deployed Helm releases specifically inside your cluster's `monitoring` namespace (using read-only RBAC), checking current versions against the latest releases with security patch advisories and release notes.
6. **Pro-Grade GUI Utilities:**
   - KPI Summary Cards, interactive search/filtering, visual latency sparkline distribution charts, customizable refresh intervals, built-in telemetry simulation, and one-click JSON data export.

---

## Official Docker Images (Container Registry)

You can pull pre-built production container images directly from GitHub Container Registry (GHCR):

```bash
docker pull ghcr.io/samipkumar-patel/cloudops-cockpit:latest
```

Or run it directly with Docker:
```bash
docker run -d -p 8080:8080 -v cockpit-data:/data ghcr.io/samipkumar-patel/cloudops-cockpit:latest
```

---

## Modular Helm Charts & Deployment Options

CloudOps Cockpit offers separate modular Helm charts so you can deploy only the features you need.

### 1. Modular Helm Charts (AKS & Kubernetes)
- **Version Radar Chart (with read-only RBAC):**
  ```bash
  helm install cloudops-radar ./deploy/helm/radar --namespace observability --create-namespace
  ```
- **Telemetry Chart (with SQLite persistence):**
  ```bash
  helm install cloudops-telemetry ./deploy/helm/telemetry --namespace observability --create-namespace
  ```

### 2. Local Development (Go Binary)
```bash
go run main.go
```
Open `http://localhost:8080`.

### 3. Docker & Docker Compose
```bash
docker compose up --build -d
```

---

## How Applications Send Data

CloudOps Cockpit exposes standard OpenTelemetry (OTLP) HTTP JSON endpoints out of the box (`/v1/traces`, `/v1/metrics`, `/v1/logs`). 

### 1. Standard OpenTelemetry Environment Variables (Zero Code Changes)
If your application uses OpenTelemetry SDKs (Go, Node.js, Python, Java, C#, etc.):
```bash
export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:8080"
export OTEL_SERVICE_NAME="my-web-app"
```

### 2. OpenTelemetry Collector / Agent (Recommended for Production)
If you run an OpenTelemetry Collector or Agent across your infrastructure, configure its `exporter` block to route traces, metrics, and logs directly to CloudOps Cockpit:

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

### 3. Direct HTTP POST (Custom Apps & Scripts)
```powershell
# Push a Trace Span
Invoke-RestMethod -Uri "http://localhost:8080/v1/traces" -Method Post -ContentType "application/json" -Body '{"service_name":"checkout-svc","operation_name":"POST /pay","duration_ms":142,"status_code":200}'

# Push a Metric Point
Invoke-RestMethod -Uri "http://localhost:8080/v1/metrics" -Method Post -ContentType "application/json" -Body '{"service_name":"checkout-svc","metric_name":"orders_processed","value":15}'

# Push a Log Entry
Invoke-RestMethod -Uri "http://localhost:8080/v1/logs" -Method Post -ContentType "application/json" -Body '{"service_name":"checkout-svc","level":"ERROR","message":"Database timeout during transaction"}'
```

---

## Maintainer & Creator

- **Samipkumar Patel** – Architect and core maintainer of CloudOps Cockpit.
- Passionate about cloud observability, high-performance Go backend architecture, and simplifying DevOps tools for small teams.
