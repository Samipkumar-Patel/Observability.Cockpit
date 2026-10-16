# LLM & AI Agent Observability Data Source for Grafana

[![GitHub License](https://img.shields.io/github/license/Samipkumar-Patel/Observability.Cockpit)](LICENSE)
[![GitHub Go Version](https://img.shields.io/github/go-mod/v/Samipkumar-Patel/Observability.Cockpit)](go.mod)
[![GitHub Stars](https://img.shields.io/github/stars/Samipkumar-Patel/Observability.Cockpit?style=social)](https://github.com/Samipkumar-Patel/Observability.Cockpit)

A specialized Grafana backend data source plugin built in Go to monitor AI agents and LLM telemetry. 

As engineering teams scale AI features built with LLMs (OpenAI, LangChain, CrewAI, n8n, etc.), this plugin bridges the gap by letting you visualize **token consumption, prompt latencies, and agent error rates** directly inside your Grafana dashboards.

---

## Project Structure

```text
├── pkg/                 # Go backend plugin (Grafana SDK query handler, health check)
├── src/                 # TypeScript frontend plugin (Query editor config)
├── plugin.json          # Grafana plugin metadata definition
└── build.go             # Go build automation script
```

---

## Key Features

1. **LLM Telemetry Ingestion:** Query token usage, prompt execution duration, and success/error rates.
2. **Grafana Data Frame Integration:** Converts backend query results directly into standard Grafana Data Frames for graphs, stat panels, and tables.
3. **Health Check API:** Built-in connection verification for Grafana data source settings.

---

## Getting Started & Local Development

### Prerequisites
- Go (1.21+)
- Node.js & npm (for frontend plugin assets)

### Build the Plugin Backend
```bash
go run build.go
```
The compiled plugin binary will be placed in the `dist/` directory.

---

## Maintainer & Creator

- **Samipkumar Patel** – Open-source contributor focused on Grafana plugins and cloud-native observability tooling.
