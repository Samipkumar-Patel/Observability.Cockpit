package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	_ "modernc.org/sqlite"
)

const defaultDBPath = "telemetry.db"

// maxRows caps how many rows a single query returns.
const maxRows = 10000

// settingsModel is the jsonData stored by the config editor.
type settingsModel struct {
	DBPath string `json:"dbPath"`
}

// queryModel is the JSON sent by the query editor.
type queryModel struct {
	Metric string `json:"metric"`
}

// LLMObservabilityDatasource reads metrics from a SQLite telemetry database.
type LLMObservabilityDatasource struct {
	db   *sql.DB
	path string
}

// NewLLMObservabilityDatasourceInstance opens the telemetry database configured
// for this data source instance.
func NewLLMObservabilityDatasourceInstance(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var s settingsModel
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &s); err != nil {
			return nil, fmt.Errorf("invalid data source settings: %w", err)
		}
	}
	path := strings.TrimSpace(s.DBPath)
	if path == "" {
		path = defaultDBPath
	}

	// sql.Open would silently create an empty database, so check the file first.
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("telemetry database not found at %q: %w", path, err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open telemetry database %q: %w", path, err)
	}
	return &LLMObservabilityDatasource{db: db, path: path}, nil
}

// Dispose closes the database connection when Grafana discards the instance.
func (d *LLMObservabilityDatasource) Dispose() {
	if d.db != nil {
		_ = d.db.Close()
	}
}

// QueryData answers each query in the request.
func (d *LLMObservabilityDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	for _, q := range req.Queries {
		response.Responses[q.RefID] = d.query(ctx, q)
	}
	return response, nil
}

func (d *LLMObservabilityDatasource) query(ctx context.Context, q backend.DataQuery) backend.DataResponse {
	var model queryModel
	if err := json.Unmarshal(q.JSON, &model); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("invalid query: %v", err))
	}
	if model.Metric == "" {
		return backend.ErrDataResponse(backend.StatusBadRequest, "select a metric")
	}
	if d.db == nil {
		return backend.ErrDataResponse(backend.StatusInternal, "telemetry database is not open")
	}

	rows, err := d.db.QueryContext(ctx,
		`SELECT timestamp, value FROM metrics WHERE metric_name = ? ORDER BY id ASC LIMIT ?`,
		model.Metric, maxRows)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("query failed: %v", err))
	}
	defer rows.Close()

	times := []time.Time{}
	values := []float64{}
	for rows.Next() {
		var rawTS string
		var val float64
		if err := rows.Scan(&rawTS, &val); err != nil {
			return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("scan failed: %v", err))
		}
		ts, err := parseTimestamp(rawTS)
		if err != nil {
			log.DefaultLogger.Warn("Skipping row with unparsable timestamp", "timestamp", rawTS)
			continue
		}
		if !ts.Before(q.TimeRange.From) && !ts.After(q.TimeRange.To) {
			times = append(times, ts)
			values = append(values, val)
		}
	}
	if err := rows.Err(); err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("read failed: %v", err))
	}

	frame := data.NewFrame("metrics",
		data.NewField("time", nil, times),
		data.NewField(model.Metric, nil, values),
	)
	return backend.DataResponse{Frames: data.Frames{frame}}
}

// parseTimestamp accepts RFC3339 and the Go time.String() format that the
// telemetry writer stores, e.g. "2026-09-27 20:03:39.3668848 +0200 CEST m=+56.6".
func parseTimestamp(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t, nil
	}
	// Drop the monotonic clock suffix " m=+..." if present.
	if i := strings.Index(raw, " m="); i >= 0 {
		raw = raw[:i]
	}
	if t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", raw); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unrecognized timestamp %q", raw)
}

// CheckHealth verifies the database file is readable and contains the metrics table.
func (d *LLMObservabilityDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if d.db == nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "telemetry database is not open"}, nil
	}
	var count int
	if err := d.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM metrics`).Scan(&count); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("cannot read metrics table in %q: %v", d.path, err),
		}, nil
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: fmt.Sprintf("LLM Observability data source connected (%d metric rows in %s)", count, d.path),
	}, nil
}
