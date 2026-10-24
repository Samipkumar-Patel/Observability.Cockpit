package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	if err := datasource.Manage("samipkumar-llm-datasource", NewLLMObservabilityDatasourceInstance, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error("Error managing LLM datasource plugin", "error", err.Error())
		os.Exit(1)
	}
}
