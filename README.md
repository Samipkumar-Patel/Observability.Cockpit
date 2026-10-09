# CloudOps Observability Platform

An all-in-one, zero-configuration observability engine built in Go for small teams without dedicated DevOps engineers.

Instead of managing Prometheus, Grafana, Loki, Tempo, and OpenTelemetry collectors, **CloudOps Cockpit** provides a single compiled binary that receives OTLP telemetry (metrics, logs, and traces), stores them in persistent SQLite, features an **AI Root-Cause Assistant** and **Open-Source Version Radar**, and serves a built-in multi-page dashboard.

---

## Official Docker Images (Container Registry)

You can pull pre-built production container images directly from GitHub Container Registry (GHCR) or Docker Hub without building from source:

```bash
docker pull ghcr.io/samipkumar-patel/cloudops-cockpit:latest
```

Or run it directly with Docker:
```bash
docker run -d -p 8080:8080 -v cockpit-data:/data ghcr.io/samipkumar-patel/cloudops-cockpit:latest
```

---

## Deployment Options

CloudOps Cockpit offers four ways to deploy depending on your infrastructure:

### 1. Local Development (Go Binary)
```bash
go run main.go
```
Open `http://localhost:8080`.

### 2. Docker & Docker Compose
```bash
docker compose up --build -d
```

### 3. Kubernetes / AKS (Raw Manifests)
```bash
kubectl apply -f deploy/kubernetes.yaml
```

### 4. Helm Chart (Production Kubernetes & AKS)
```bash
helm install cloudops-cockpit ./deploy/helm --namespace observability --create-namespace
```
