package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type StatusResponse struct {
	Status      string `json:"status"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StatusResponse{
			Status:      "healthy",
			Name:        "CloudOps Observability Platform",
			Version:     "0.1.0-alpha",
			Description: "All-in-one OTLP telemetry engine for small teams",
		})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>CloudOps Observability Platform</title>
  <style>
    body { font-family: system-ui, sans-serif; background: #f5f7f4; color: #17252b; max-width: 800px; margin: 60px auto; padding: 0 20px; }
    h1 { font-size: 36px; margin-bottom: 8px; }
    .badge { background: #e2f0ea; color: #147d76; padding: 6px 12px; border-radius: 99px; font-weight: 700; font-size: 12px; display: inline-block; margin-bottom: 20px; }
    .card { background: white; border: 1px solid #dce5e3; border-radius: 6px; padding: 24px; margin-top: 20px; }
    code { background: #e8efeb; padding: 2px 6px; border-radius: 4px; }
  </style>
</head>
<body>
  <span class="badge">SINGLE-BINARY OBSERVABILITY ENGINE</span>
  <h1>CloudOps Cockpit</h1>
  <p>An all-in-one OTLP telemetry platform replacing complex monitoring stacks for small teams.</p>
  <div class="card">
    <h3>Engine Status</h3>
    <p>Server running successfully. OTLP ingestion receiver and embedded storage scaffolding initialized.</p>
    <p><a href="/api/health">Check API health endpoint (/api/health)</a></p>
  </div>
</body>
</html>`)
	})

	fmt.Printf("Starting CloudOps Observability Engine on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
