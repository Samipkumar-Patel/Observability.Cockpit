# LLM & AI Agent Observability Data Source for Grafana

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
- Docker (optional, for local Grafana testing)

### Build the Plugin Backend
```bash
go run build.go
```
The compiled plugin binary will be placed in the `dist/` directory.

---

## Local Testing Guide with Docker

To test your custom data source plugin locally inside Grafana:

1. Build the plugin binary:
   ```bash
   go run build.go
   ```
2. Run a local Grafana container and mount your repository directory into Grafana's plugin folder:
   ```bash
   docker run -d -p 3000:3000 \
     --name=grafana \
     -e "GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=samipkumar-llm-datasource" \
     -v "$(pwd):/var/lib/grafana/plugins/llm-datasource" \
     grafana/grafana:latest
   ```
3. Open `http://localhost:3000` (Login with `admin` / `admin`).
4. Navigate to **Connections** ➔ **Data sources** ➔ **Add data source**.
5. Search for **LLM Observability Data Source**, select it, and test your connection!

---

## Maintainer & Creator

- **Samipkumar Patel** – Open-source contributor focused on Grafana plugins and cloud-native observability tooling.
