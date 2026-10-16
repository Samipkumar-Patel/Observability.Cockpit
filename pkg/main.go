package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	// Start the datasource plugin backend server
	if err := datasource.Manage("samipkumar-custom-datasource", NewCustomDatasource, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error("Error managing datasource plugin", "error", err.Error())
		os.Exit(1)
	}
}
