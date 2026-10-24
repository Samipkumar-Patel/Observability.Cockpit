package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestDB creates a telemetry database with the same schema and row format
// as the telemetry writer, and returns its path.
func newTestDB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "telemetry.db")
	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_name TEXT,
		metric_name TEXT,
		value REAL,
		timestamp DATETIME)`)
	require.NoError(t, err)

	rows := []struct {
		svc, metric string
		value       float64
		ts          string
	}{
		{"orders-api", "request_duration_ms", 176, "2026-09-27 21:12:41.1011315 +0200 CEST m=+18.536230601"},
		{"orders-api", "request_duration_ms", 214, "2026-10-09 21:14:12.3883189 +0200 CEST m=+48.518411501"},
		{"checkout-svc", "orders_processed", 15, "2026-09-27 20:03:39.3668848 +0200 CEST m=+56.648364401"},
		{"orders-api", "request_duration_ms", 99, "2026-10-09T21:14:20Z"},
	}
	for _, r := range rows {
		_, err := db.Exec(`INSERT INTO metrics (service_name, metric_name, value, timestamp) VALUES (?, ?, ?, ?)`,
			r.svc, r.metric, r.value, r.ts)
		require.NoError(t, err)
	}
	return path
}

func newTestInstance(t *testing.T, dbPath string) *LLMObservabilityDatasource {
	t.Helper()
	settings := backend.DataSourceInstanceSettings{
		JSONData: []byte(`{"dbPath":"` + filepath.ToSlash(dbPath) + `"}`),
	}
	inst, err := NewLLMObservabilityDatasourceInstance(context.Background(), settings)
	require.NoError(t, err)
	ds := inst.(*LLMObservabilityDatasource)
	t.Cleanup(ds.Dispose)
	return ds
}

func TestNewInstanceMissingDB(t *testing.T) {
	settings := backend.DataSourceInstanceSettings{
		JSONData: []byte(`{"dbPath":"does-not-exist.db"}`),
	}
	_, err := NewLLMObservabilityDatasourceInstance(context.Background(), settings)
	assert.Error(t, err)
}

func TestCheckHealth(t *testing.T) {
	ds := newTestInstance(t, newTestDB(t))
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusOk, res.Status)
	assert.Contains(t, res.Message, "LLM Observability data source connected")
	assert.Contains(t, res.Message, "4 metric rows")
}

func TestQueryFiltersByMetric(t *testing.T) {
	ds := newTestInstance(t, newTestDB(t))

	q := backend.DataQuery{
		RefID: "A",
		JSON:  []byte(`{"metric":"request_duration_ms"}`),
		TimeRange: backend.TimeRange{
			From: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{Queries: []backend.DataQuery{q}})
	require.NoError(t, err)

	res := resp.Responses["A"]
	require.NoError(t, res.Error)
	require.Len(t, res.Frames, 1)
	assert.Equal(t, 3, res.Frames[0].Rows())
}

func TestQueryRespectsTimeRange(t *testing.T) {
	ds := newTestInstance(t, newTestDB(t))

	q := backend.DataQuery{
		RefID: "A",
		JSON:  []byte(`{"metric":"request_duration_ms"}`),
		TimeRange: backend.TimeRange{
			From: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),
		},
	}
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{Queries: []backend.DataQuery{q}})
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Responses["A"].Frames[0].Rows())
}

func TestQueryRequiresMetric(t *testing.T) {
	ds := newTestInstance(t, newTestDB(t))
	q := backend.DataQuery{RefID: "A", JSON: json.RawMessage(`{}`)}
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{Queries: []backend.DataQuery{q}})
	require.NoError(t, err)
	assert.Error(t, resp.Responses["A"].Error)
}

func TestParseTimestamp(t *testing.T) {
	cases := []string{
		"2026-10-09 21:14:12.3883189 +0200 CEST m=+48.518411501",
		"2026-10-09 21:14:12.3883189 +0200 CEST",
		"2026-10-09T21:14:12Z",
	}
	for _, c := range cases {
		_, err := parseTimestamp(c)
		assert.NoError(t, err, c)
	}
	_, err := parseTimestamp("not a timestamp")
	assert.Error(t, err)
}
