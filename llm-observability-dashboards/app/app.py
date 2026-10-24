"""Example LLM agent that emits OpenTelemetry GenAI metrics.

No real LLM is called. Each "chat" call is simulated with random latency,
token counts, and occasional errors, so the dashboards have realistic data
without an API key.
"""

import os
import random
import time

from opentelemetry import metrics
from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.sdk.metrics.view import ExplicitBucketHistogramAggregation, View

SERVICE_NAME = os.getenv("SERVICE_NAME", "support-agent")
AGENT_NAME = os.getenv("AGENT_NAME", "support-agent")
INTERVAL_SECONDS = float(os.getenv("RUN_INTERVAL_SECONDS", "2"))

# Models and their providers, as they would appear in gen_ai.* attributes.
MODELS = [
    ("gpt-4o-mini", "openai"),
    ("claude-3-5-haiku", "anthropic"),
    ("llama3", "ollama"),
]

# Spec-recommended buckets for gen_ai.client.operation.duration (seconds).
DURATION_BUCKETS = [0.01, 0.02, 0.04, 0.08, 0.16, 0.32, 0.64, 1.28, 2.56, 5.12, 10.24, 20.48, 40.96, 81.92]

TOOLS = ["search_docs", "get_order_status", "create_ticket"]


def setup_metrics():
    exporter = OTLPMetricExporter()  # reads OTEL_EXPORTER_OTLP_ENDPOINT
    reader = PeriodicExportingMetricReader(exporter, export_interval_millis=5000)
    views = [
        View(
            instrument_name="gen_ai.client.operation.duration",
            aggregation=ExplicitBucketHistogramAggregation(DURATION_BUCKETS),
        ),
        View(
            instrument_name="gen_ai.execute_tool.duration",
            aggregation=ExplicitBucketHistogramAggregation(DURATION_BUCKETS),
        ),
        View(
            instrument_name="gen_ai.invoke_agent.duration",
            aggregation=ExplicitBucketHistogramAggregation(DURATION_BUCKETS),
        ),
    ]
    provider = MeterProvider(
        metric_readers=[reader],
        views=views,
        resource=metrics_resource(),
    )
    metrics.set_meter_provider(provider)
    return metrics.get_meter("llm-observability-example")


def metrics_resource():
    from opentelemetry.sdk.resources import Resource

    return Resource.create({"service.name": SERVICE_NAME})


def build_instruments(meter):
    return {
        "operation_duration": meter.create_histogram(
            "gen_ai.client.operation.duration", unit="s",
            description="GenAI operation duration",
        ),
        "input_tokens": meter.create_counter(
            "gen_ai.client.inference.usage.input_tokens", unit="{token}",
            description="Input tokens used by inference calls",
        ),
        "output_tokens": meter.create_counter(
            "gen_ai.client.inference.usage.output_tokens", unit="{token}",
            description="Output tokens produced by inference calls",
        ),
        "cache_read_tokens": meter.create_counter(
            "gen_ai.client.inference.usage.cache_read.input_tokens", unit="{token}",
            description="Input tokens read from prompt cache",
        ),
        "agent_duration": meter.create_histogram(
            "gen_ai.invoke_agent.duration", unit="s",
            description="Duration of an agent invocation",
        ),
        "agent_inference_calls": meter.create_histogram(
            "gen_ai.invoke_agent.inference_calls", unit="{call}",
            description="LLM calls made during one agent invocation",
        ),
        "agent_tool_calls": meter.create_histogram(
            "gen_ai.invoke_agent.tool_calls", unit="{call}",
            description="Tool calls made during one agent invocation",
        ),
        "tool_duration": meter.create_histogram(
            "gen_ai.execute_tool.duration", unit="s",
            description="Duration of a tool execution",
        ),
    }


def simulate_chat(inst, model, provider):
    """One simulated LLM call. Returns True on success."""
    failed = random.random() < 0.05
    if failed:
        latency = random.uniform(0.2, 1.0)
        error_type = random.choice(["rate_limit", "timeout", "server_error"])
    else:
        latency = random.lognormvariate(0.2, 0.5)
        error_type = None

    attrs = {
        "gen_ai.operation.name": "chat",
        "gen_ai.provider.name": provider,
        "gen_ai.request.model": model,
    }
    if error_type:
        attrs["error.type"] = error_type

    time.sleep(min(latency, 0.5))  # keep the demo fast; metrics record the simulated latency
    inst["operation_duration"].record(latency, attributes=attrs)

    if not failed:
        input_tokens = random.randint(200, 2000)
        output_tokens = random.randint(50, 800)
        inst["input_tokens"].add(input_tokens, attributes=attrs)
        inst["output_tokens"].add(output_tokens, attributes=attrs)
        if random.random() < 0.3:
            inst["cache_read_tokens"].add(random.randint(100, 1000), attributes=attrs)
    return not failed


def simulate_tool(inst, tool_name):
    duration = random.uniform(0.05, 1.5)
    inst["tool_duration"].record(duration, attributes={"gen_ai.tool.name": tool_name})


def run_agent(inst):
    """One agent invocation: a few LLM calls and tool calls."""
    model, provider = random.choice(MODELS)
    inference_calls = random.randint(1, 4)
    tool_calls = random.randint(0, 3)

    agent_start = time.time()
    for _ in range(inference_calls):
        simulate_chat(inst, model, provider)
    for _ in range(tool_calls):
        simulate_tool(inst, random.choice(TOOLS))
    agent_duration = time.time() - agent_start + random.uniform(0.5, 3.0)

    agent_attrs = {"gen_ai.agent.name": AGENT_NAME}
    inst["agent_duration"].record(agent_duration, attributes=agent_attrs)
    inst["agent_inference_calls"].record(inference_calls, attributes=agent_attrs)
    inst["agent_tool_calls"].record(tool_calls, attributes=agent_attrs)


def main():
    meter = setup_metrics()
    inst = build_instruments(meter)
    print(f"Emitting GenAI metrics for agent '{AGENT_NAME}' every {INTERVAL_SECONDS}s", flush=True)
    while True:
        run_agent(inst)
        time.sleep(INTERVAL_SECONDS)


if __name__ == "__main__":
    main()
