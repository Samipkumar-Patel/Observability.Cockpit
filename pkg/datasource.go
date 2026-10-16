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

type CustomDatasource struct{}

func NewCustomDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &CustomDatasource{}, nil
}

func (d *CustomDatasource) Dispose() {}

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

	frame := data.NewFrame("response",
		data.NewField("time", nil, []time.Time{time.Now()}),
		data.NewField("value", nil, []float64{42.0}),
	)

	response.Frames = append(response.Frames, frame)
	return response
}

func (d *CustomDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Custom data source is healthy and operational",
	}, nil
}
