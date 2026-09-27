# CloudOps Observability Platform

An all-in-one, zero-configuration observability engine built in Go for small teams without dedicated DevOps engineers.

Instead of managing Prometheus, Grafana, Loki, Tempo, and OpenTelemetry collectors, **CloudOps Cockpit** provides a single compiled binary that receives OTLP telemetry (metrics, logs, and traces), stores them in persistent SQLite, and serves a built-in dashboard.

---

## Deployment Options

CloudOps Cockpit is designed for flexibility, offering four ways to deploy depending on your infrastructure:

### 1. Local Development (Go Binary)
Great for local testing or lightweight scripts.
```bash
go run main.go
```
Open `http://localhost:8080`.

### 2. Docker & Docker Compose
Spin up the container with persistent local volume storage in a single command.
```bash
docker compose up --build -d
```
Open `http://localhost:8080`. To stop: `docker compose down`.

### 3. Kubernetes / AKS (Raw Manifests)
Deploy to any Kubernetes cluster or Azure Kubernetes Service (AKS) using our pre-configured manifests (includes Namespace, PVC for SQLite, Deployment, Service, and Ingress for Azure Application Gateway):
```bash
kubectl apply -f deploy/kubernetes.yaml
```

### 4. Helm Chart (Production Kubernetes & AKS)
For robust cluster management, use our Helm chart located in `deploy/helm/`:
```bash
helm install cloudops-cockpit ./deploy/helm --namespace observability --create-namespace
```
To customize values (such as Ingress hosts or storage size):
```bash
helm install cloudops-cockpit ./deploy/helm --namespace observability --set ingress.host="cockpit.yourdomain.com"
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
