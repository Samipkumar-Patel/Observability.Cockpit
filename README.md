# Azure Cost Cockpit

An open-source Azure-first cloud cost and reliability cockpit for small teams.

The first product question is:

> Why did our Azure bill change, and what can we safely do about it?

This repository contains a runnable demo vertical slice and an optional Azure Cost Management data source. Demo mode uses safe local sample data and does not require Azure credentials.

## Run locally

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r requirements.txt
uvicorn app.main:app --reload
```

Open <http://127.0.0.1:8000>.

Edit `cockpit.config.json` to choose demo or Azure mode. The dashboard reads this configuration when the backend starts.

## Use live Azure data

The application uses `DefaultAzureCredential`, so it supports Azure CLI login locally and managed identity or service principal credentials when deployed.

```json
{
  "data_source": "azure",
  "azure_subscription_id": "00000000-0000-0000-0000-000000000000"
}
```

```powershell
az login
uvicorn app.main:app --reload
```

The signed-in identity needs permission to query costs at the subscription scope. Assign the **Cost Management Reader** role, or another role that includes cost query permissions. The live integration currently provides monthly cost trend and service spend. Resource-level cost attribution and utilization data will be added through separate, lighter integrations.

To return to local sample data, set `data_source` back to `demo` in `cockpit.config.json`.

## Current MVP

- Azure subscription and monthly spend summary
- Cost trend by month
- Spend grouped by service
- Resources with estimated monthly cost
- Cost optimization recommendations with estimated savings and risk
- Demo mode that works without Azure credentials
- Optional live Azure Cost Management data source

## Planned Azure integrations

The current Azure integration uses Azure Identity and Azure Cost Management APIs to collect:

- Actual cost by subscription, resource group, service, and resource

Future integrations will add:

- Resource inventory and utilization metadata
- Budget and forecast information
- Change events for deployment and cost correlation

The project is intentionally starting with one narrow workflow instead of trying to replace Azure Cost Management or build a complete observability platform.

## Project layout

```text
app/
  main.py              FastAPI application and API routes
  models.py            Product domain models
  config.py            Environment-based data source configuration
  services.py          Provider selection and recommendation logic
  azure_provider.py    Azure Cost Management provider
  static/              Lightweight dashboard
```
