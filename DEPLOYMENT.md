# Deployment & Execution Guide

CloudOps Cockpit can be deployed locally, via Docker, or onto Azure Kubernetes Service (AKS) using modular Helm charts.

---

## 1. Local Development

1. Run the server directly:
   ```bash
   go run main.go
   ```
2. Open `http://localhost:8080`.

---

## 2. Docker & Docker Compose

Run the platform as a container with persistent storage in a single command:

```bash
docker compose up --build -d
```

Access the UI at `http://localhost:8080`. To stop the container:
```bash
docker compose down
```

---

## 3. Modular Helm Charts (Production Kubernetes & AKS)

CloudOps Cockpit features separate modular Helm charts so you can install only the features you need.

### Option A: Install Version Radar Chart
Tracks and audits Helm release versions specifically inside your `monitoring` namespace (includes RBAC read-only reader role).
```bash
helm install cloudops-radar ./deploy/helm/radar --namespace observability --create-namespace
```

### Option B: Install Telemetry Chart
Receives OTLP Traces, Metrics, and Logs with persistent SQLite storage (`/data/telemetry.db`).
```bash
helm install cloudops-telemetry ./deploy/helm/telemetry --namespace observability --create-namespace
```

### Customizing Values
You can override default settings (such as Ingress hosts or target namespaces) using `--set`:
```bash
helm install cloudops-radar ./deploy/helm/radar --namespace observability --set monitoringNamespace="my-custom-monitoring"
```
