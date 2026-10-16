package main

import (
	"context"
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	if err := datasource.Manage("samipkumar-llm-datasource", NewLLMObservabilityDatasourceInstance, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error("Error managing LLM datasource plugin", "error", err.Error())
		os.Exit(1)
	}
}

func NewLLMObservabilityDatasourceInstance(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &LLMObservabilityDatasource{}, nil
}
