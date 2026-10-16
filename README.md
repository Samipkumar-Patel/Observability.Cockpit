# LLM & AI Agent Observability Data Source for Grafana

A specialized Grafana backend data source plugin built in Go to monitor AI agents and LLM telemetry. 

---

## Core Objective & Vision

As companies increasingly build features using Large Language Models (LLMs) and autonomous AI agents (OpenAI, Anthropic, LangChain, CrewAI, n8n, custom pipelines), they face a major blind spot: **traditional monitoring tools (like Prometheus or Datadog) were built for CPU, memory, and HTTP requests, not for AI.**

The objective of this plugin is to give engineering teams **native visibility into their AI applications directly inside Grafana**.

---

## What Problems Does It Solve?

When running AI workloads in production, engineering and SRE teams urgently need to track operational metrics that standard APM and monitoring tools completely miss:

1. **Token Consumption & Cost Control:** 
   * *Problem:* AI APIs bill strictly by token count (prompt + completion tokens). Without visibility, a single runaway loop in an AI agent can rack up unexpected bills overnight.
   * *Solution:* Tracks token usage per agent and model in real time so teams can set up alerts and dashboards.
2. **Prompt Latency & Performance Regressions:** 
   * *Problem:* LLM response times are inherently variable and often slow (taking seconds rather than milliseconds), impacting user experience.
   * *Solution:* Visualizes prompt execution duration and latency distributions to identify slow-performing model prompts or congested vector database lookups.
3. **Agent Failure Rates & Exceptions:** 
   * *Problem:* Autonomous agents frequently fail due to API rate limits (HTTP 429), context-window overflows, or JSON parsing errors when structuring LLM output.
   * *Solution:* Captures agent error rates and failure exceptions directly in Grafana alongside infrastructure health.

---

## Project Structure

```text
├── pkg/                 # Go backend plugin (Grafana SDK query handler, health check, SQLite integration, unit tests)
├── src/                 # TypeScript frontend plugin (Query editor config)
├── plugin.json          # Grafana plugin metadata definition
└── build.go             # Go build automation script
```

---

## Local Testing Guide with Docker (Easiest Method)

To test your custom data source plugin locally inside Grafana without building TypeScript frontend bundles:

1. **Build the Plugin Backend:**
   ```bash
   go run build.go
   ```
2. **Run Grafana with Pre-built Plugin Directory Mounted:**
   Open PowerShell and run this exact command to spin up Grafana with unsigned plugin loading enabled:
   ```powershell
   $cwd = (Get-Item .).FullName
   docker run -d -p 3000:3000 --name=grafana -e "GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=samipkumar-llm-datasource" -v "${cwd}:/var/lib/grafana/plugins/llm-datasource" grafana/grafana:11.4.0
   ```
3. **Access Grafana:**
   - Open **`http://localhost:3000`** in your browser.
   - Login with username **`admin`** and password **`admin`**.
4. **Add Your Data Source:**
   - Go to **Connections** ➔ **Data sources** ➔ **Add data source**.
   - Search for **LLM Observability Data Source**, select it, and click **Save & test**.

---

## Maintainer & Creator

- **Samipkumar Patel** – Open-source contributor focused on Grafana plugins and cloud-native observability tooling.
