# LLM & Agent Observability Dashboards

Grafana dashboards for LLM and AI agent telemetry, built on the [OpenTelemetry GenAI semantic conventions](https://github.com/open-telemetry/semantic-conventions-genai). They use the Prometheus data source, so no custom Grafana plugin is needed.

Status: **early (v0.1)**. The GenAI metric names are still marked "Development" in the spec and may change.

## What you get

- **Token usage:** input and output tokens per model, per minute.
- **Requests and errors:** request rate, error rate, and errors by type (for example `rate_limit`, `timeout`).
- **Latency:** p50 and p95 per model.
- **Agents:** run duration p95, and average LLM and tool calls per run.
- **Tools:** execution p95 per tool.

Dashboard variables filter by **service** and **model**.

## Quick start

Requirements: Docker with Compose.

```powershell
cd llm-observability-dashboards
docker compose up -d --build
```

Then open:
- Grafana: http://localhost:3000 (anonymous login is disabled; log in with `admin` / `admin`)
- Dashboard: **Dashboards > LLM Observability > LLM & Agent Observability**

Wait about a minute for the example app's metrics to appear.

## Architecture

```
example app (OTel SDK)  --OTLP/HTTP-->  OpenTelemetry Collector  --scrape-->  Prometheus  <--query--  Grafana
```

| Folder | Contents |
|---|---|
| `app/` | Example agent that emits GenAI metrics. It simulates LLM and tool calls, so no API key is needed. |
| `otel/` | Collector config: OTLP receiver, Prometheus exporter. |
| `prometheus/` | Scrape config. |
| `grafana/provisioning/` | Data source and dashboard provisioning. |
| `grafana/dashboards/` | Dashboard JSON. |

## Metrics reference

Prometheus names are derived from the OpenTelemetry names (dots become underscores, and counters get `_total`, histograms get `_bucket`, `_sum`, `_count`).

| OpenTelemetry metric | Type | Prometheus name | Key labels |
|---|---|---|---|
| `gen_ai.client.operation.duration` | Histogram (s) | `gen_ai_client_operation_duration_seconds_*` | `gen_ai_operation_name`, `gen_ai_provider_name`, `gen_ai_request_model`, `error_type` |
| `gen_ai.client.inference.usage.input_tokens` | Counter | `gen_ai_client_inference_usage_input_tokens_total` | `gen_ai_request_model`, `gen_ai_provider_name` |
| `gen_ai.client.inference.usage.output_tokens` | Counter | `gen_ai_client_inference_usage_output_tokens_total` | `gen_ai_request_model`, `gen_ai_provider_name` |
| `gen_ai.client.inference.usage.cache_read.input_tokens` | Counter | `gen_ai_client_inference_usage_cache_read_input_tokens_total` | `gen_ai_request_model`, `gen_ai_provider_name` |
| `gen_ai.invoke_agent.duration` | Histogram (s) | `gen_ai_invoke_agent_duration_seconds_*` | `gen_ai_agent_name` |
| `gen_ai.invoke_agent.inference_calls` | Histogram | `gen_ai_invoke_agent_inference_calls_*` | `gen_ai_agent_name` |
| `gen_ai.invoke_agent.tool_calls` | Histogram | `gen_ai_invoke_agent_tool_calls_*` | `gen_ai_agent_name` |
| `gen_ai.execute_tool.duration` | Histogram (s) | `gen_ai_execute_tool_duration_seconds_*` | `gen_ai_tool_name` |

The service name appears as the `exported_job` label, because the collector's own `job` label takes precedence.

## Instrumenting your own app

Emit the same metrics from your app with the OpenTelemetry SDK, and point the exporter at the collector (`OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318`). The example in `app/app.py` shows the instrument names, units, and histogram buckets to use.

## Known limitations

- **No cost panel yet.** Cost needs a price table per model, which is planned for a later version.
- **Example data is simulated.** Latency, tokens, and errors are random, so the dashboards show the layout, not real behavior.
- **Metric names may change.** The GenAI conventions are still in development.
- **Histogram estimates.** Percentiles are estimated from bucket boundaries, so they are approximate.

## License

To be decided before the first public release.
