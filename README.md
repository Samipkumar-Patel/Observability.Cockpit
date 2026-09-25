# CloudOps Observability Platform

An all-in-one, zero-configuration observability engine built in Go for small teams without dedicated DevOps engineers.

Instead of managing Prometheus, Grafana, Loki, Tempo, and OpenTelemetry collectors, **CloudOps Cockpit** provides a single compiled binary that receives OTLP telemetry (metrics, logs, and traces), stores them efficiently, and serves a built-in dashboard.

## Vision

- **Single Binary:** Download and run (`./cloudops-cockpit`). No complex container orchestration or external databases required for small workloads.
- **OpenTelemetry Native:** Accepts standard OTLP ingestion out of the box.
- **Unified UI:** Metrics, logs, traces, and service health in one coherent interface.

## Quick Start (Coming Soon)

```bash
go build -o cloudops-cockpit main.go
./cloudops-cockpit
```

Open `http://localhost:8080`.
