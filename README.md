# Grafana Custom Data Source Plugin

A lightweight, production-ready backend data source plugin skeleton for Grafana, built in Go and TypeScript. 

This repository serves as a starting point for developers wanting to build and publish custom plugins to the official Grafana Plugin Catalog.

---

## Project Structure

```text
├── pkg/                 # Go backend plugin (query data, health check)
├── src/                 # TypeScript frontend plugin (config editor, query editor)
├── plugin.json          # Grafana plugin metadata definition
└── MAGEFILE.go          # Grafana plugin build and packaging automation
```

---

## Getting Started

### Prerequisites
- Go (1.21+)
- Node.js & npm (for frontend UI)
- [Mage](https://magefile.org/) (Grafana's build tool)

### Build the Plugin
```bash
# Build backend and frontend assets
go run build.go
```

---

## Maintainer & Creator

- **Samipkumar Patel** – Open-source contributor focused on Grafana plugins and cloud-native observability tooling.
