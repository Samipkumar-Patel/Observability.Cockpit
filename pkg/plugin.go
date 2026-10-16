package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// Ensure Datasource implements requisite interfaces
var _ datasource.InstanceSettingsMarshaler = (*CustomDatasource)(nil)

type CustomDatasource struct{}

func NewCustomDatasource(setting backend.DataSourceInstanceSettings) (datasource.InstanceSettings, error) {
	return &CustomDatasource{}, nil
}

func (d *CustomDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		res := d.query(ctx, q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type QueryModel struct {
	QueryText string `json:"queryText"`
}

func (d *CustomDatasource) query(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	var model QueryModel
	if err := json.Unmarshal(query.JSON, &model); err != nil {
		response.Error = err
		return response
	}

	log.DefaultLogger.Info("Executing custom data source query", "query", model.QueryText)

	// Return a simple mock data frame to Grafana
	// In production, connect this to your custom API, database, or telemetry source.
	frame := data.NewFrame("response")
	frame.AddField(data.NewField("time", nil, []time.Time{time.Now()}))
	frame.AddField(data.NewField("value", nil, []float64{42.0}))

	response.Frames = append(response.Frames, frame)
	return response
}

func (d *CustomDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Custom data source is healthy and operational",
	}, nil
}
