# Azure Cost Cockpit

An open-source Azure-first cloud cost and reliability cockpit for small teams.

The first product question is:

> Why did our Azure bill change, and what can we safely do about it?

This repository currently contains a runnable demo vertical slice. It uses safe local sample data and does not require Azure credentials. The Azure Cost Management integration will be added behind the same API once the core workflow is validated.

## Run locally

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r requirements.txt
uvicorn app.main:app --reload
```

Open <http://127.0.0.1:8000>.

## Current MVP

- Azure subscription and monthly spend summary
- Cost trend by month
- Spend grouped by service
- Resources with estimated monthly cost
- Cost optimization recommendations with estimated savings and risk
- Demo mode that works without Azure credentials

## Planned Azure integration

The next integration will use Azure Identity and Azure Cost Management APIs to collect:

- Actual cost by subscription, resource group, service, and tag
- Resource inventory and utilization metadata
- Budget and forecast information
- Change events for deployment and cost correlation

The project is intentionally starting with one narrow workflow instead of trying to replace Azure Cost Management or build a complete observability platform.

## Project layout

```text
app/
  main.py              FastAPI application and API routes
  models.py            Product domain models
  services.py          Demo data and recommendation logic
  static/              Lightweight dashboard
```
