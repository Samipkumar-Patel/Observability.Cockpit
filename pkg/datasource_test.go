package main

import (
	"context"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckHealth(t *testing.T) {
	ds := &LLMObservabilityDatasource{}
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusOk, res.Status)
	assert.Contains(t, res.Message, "LLM Observability data source connected")
}
