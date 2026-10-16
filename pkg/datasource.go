package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type LLMObservabilityDatasource struct{}

func NewLLMObservabilityDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &LLMObservabilityDatasource{}, nil
}

func (d *LLMObservabilityDatasource) Dispose() {}

func (d *LLMObservabilityDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		res := d.query(ctx, q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type QueryModel struct {
	AgentName string `json:"agentName"`
	Metric    string `json:"metric"` // e.g., "token_usage", "prompt_latency"
}

func (d *LLMObservabilityDatasource) query(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	var model QueryModel
	if err := json.Unmarshal(query.JSON, &model); err != nil {
		response.Error = err
		return response
	}

	log.DefaultLogger.Info("Fetching LLM Observability metrics", "agent", model.AgentName, "metric", model.Metric)

	// Return telemetry data frame for AI agents (Token usage, Latency, Cost)
	frame := data.NewFrame("llm_telemetry",
		data.NewField("time", nil, []time.Time{time.Now().Add(-2 * time.Minute), time.Now()}),
		data.NewField("tokens_consumed", nil, []float64{450, 1250}),
		data.NewField("prompt_latency_ms", nil, []float64{310, 840}),
	)

	response.Frames = append(response.Frames, frame)
	return response
}

func (d *LLMObservabilityDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "LLM Observability data source connected successfully",
	}, nil
}
