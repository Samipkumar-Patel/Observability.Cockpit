from calendar import monthrange
from datetime import date, datetime, time, timedelta, timezone
from time import sleep

from azure.core.exceptions import HttpResponseError
from azure.identity import DefaultAzureCredential
from azure.mgmt.costmanagement import CostManagementClient
from azure.mgmt.costmanagement.models import (
    QueryAggregation,
    QueryDataset,
    QueryDefinition,
    QueryGrouping,
    QueryTimePeriod,
)
from azure.mgmt.resource import SubscriptionClient

from .models import CostPoint, Overview, ResourceCost, ServiceCost
from .services import _recommendations


class AzureCostProvider:
    def __init__(self, subscription_id: str):
        self.subscription_id = subscription_id
        self.scope = f"/subscriptions/{subscription_id}"
        credential = DefaultAzureCredential()
        self.cost_client = CostManagementClient(credential)
        self.subscription_client = SubscriptionClient(credential)

    def get_overview(self) -> Overview:
        today = datetime.now(timezone.utc).date()
        current_start = today.replace(day=1)
        previous_start = _previous_month(current_start)
        trend_start = _previous_month(_previous_month(_previous_month(current_start)))

        rows = self._query(
            trend_start,
            today,
            granularity="Monthly",
            grouping=["ServiceName"],
        )
        current_rows = [row for row in rows if current_start <= _usage_date(row) <= today]
        previous_rows = [
            row for row in rows if previous_start <= _usage_date(row) < current_start
        ]

        current_cost = _total(current_rows)
        previous_cost = _total(previous_rows)
        change_percent = ((current_cost - previous_cost) / previous_cost * 100) if previous_cost else 0
        elapsed_days = max(today.day, 1)
        days_in_month = monthrange(today.year, today.month)[1]
        forecast = current_cost / elapsed_days * days_in_month

        service_totals = _group_total(current_rows, "ServiceName")
        service_costs = [
            ServiceCost(
                service=service,
                amount=round(amount, 2),
                share=round(amount / current_cost * 100, 1) if current_cost else 0,
            )
            for service, amount in sorted(service_totals.items(), key=lambda item: item[1], reverse=True)
        ]
        # Add resource-level costs through a separate, lighter integration.
        resources: list[ResourceCost] = []
        subscription = self.subscription_client.subscriptions.get(self.subscription_id)
        currency = next((str(row.get("Currency")) for row in rows if row.get("Currency")), "USD")

        return Overview(
            data_source="azure",
            subscription_name=subscription.display_name or self.subscription_id,
            currency=currency,
            current_month_cost=round(current_cost, 2),
            previous_month_cost=round(previous_cost, 2),
            change_percent=round(change_percent, 1),
            forecasted_month_cost=round(forecast, 2),
            cost_trend=_trend_points(rows),
            service_costs=service_costs,
            resources=resources,
            recommendations=_recommendations(resources),
        )

    def _query(self, start: date, end: date, *, grouping: list[str], granularity: str | None = None) -> list[dict]:
        dataset = QueryDataset(
            granularity=granularity,
            aggregation={"totalCost": QueryAggregation(name="PreTaxCost", function="Sum")},
            grouping=[QueryGrouping(type="Dimension", name=name) for name in grouping],
        )
        definition = QueryDefinition(
            type="Usage",
            timeframe="Custom",
            time_period=QueryTimePeriod(
                from_property=datetime.combine(start, time.min, tzinfo=timezone.utc),
                to=datetime.combine(end, time.max, tzinfo=timezone.utc),
            ),
            dataset=dataset,
        )
        result = self._run_query(definition)
        columns = [column.name for column in result.columns]
        return [dict(zip(columns, row)) for row in result.rows]

    def _run_query(self, definition: QueryDefinition):
        for attempt in range(3):
            try:
                return self.cost_client.query.usage(self.scope, definition)
            except HttpResponseError as error:
                if error.status_code != 429 or attempt == 2:
                    raise
                headers = getattr(error.response, "headers", {})
                retry_after = next(
                    (
                        value
                        for key, value in headers.items()
                        if key.lower()
                        in {
                            "retry-after",
                            "x-ms-ratelimit-microsoft.costmanagement-clienttype-retry-after",
                        }
                    ),
                    None,
                )
                try:
                    delay = min(float(retry_after), 60) if retry_after else 2**attempt
                except (TypeError, ValueError):
                    delay = 2**attempt
                sleep(delay)
def _previous_month(month_start: date) -> date:
    return (month_start - timedelta(days=1)).replace(day=1)


def _total(rows: list[dict]) -> float:
    return sum(float(row.get("PreTaxCost", row.get("totalCost", 0)) or 0) for row in rows)


def _usage_date(row: dict) -> date:
    value = row.get("UsageDate") or row.get("BillingMonth")
    if isinstance(value, datetime):
        return value.date()
    text = str(value)
    if len(text) == 8 and text.isdigit():
        return datetime.strptime(text, "%Y%m%d").date()
    return datetime.fromisoformat(text.replace("Z", "+00:00")).date()


def _group_total(rows: list[dict], key: str) -> dict[str, float]:
    totals: dict[str, float] = {}
    for row in rows:
        name = str(row.get(key) or "Unknown")
        totals[name] = totals.get(name, 0) + float(row.get("PreTaxCost", row.get("totalCost", 0)) or 0)
    return totals


def _resource_costs(rows: list[dict]) -> list[ResourceCost]:
    totals: dict[str, float] = {}
    metadata: dict[str, dict] = {}
    for row in rows:
        resource_id = str(row.get("ResourceId") or "unknown-resource")
        totals[resource_id] = totals.get(resource_id, 0) + _total([row])
        resource_group, resource_type = _resource_metadata(resource_id)
        metadata[resource_id] = {
            "resource_type": resource_type,
            "resource_group": resource_group,
        }
    resources = [
        ResourceCost(
            name=resource_id.rsplit("/", 1)[-1],
            resource_type=details["resource_type"],
            resource_group=details["resource_group"],
            monthly_cost=round(totals[resource_id], 2),
            utilization=None,
            environment="Unknown",
        )
        for resource_id, details in metadata.items()
    ]
    return sorted(resources, key=lambda resource: resource.monthly_cost, reverse=True)


def _resource_metadata(resource_id: str) -> tuple[str, str]:
    parts = resource_id.strip("/").split("/")
    lowered = [part.lower() for part in parts]
    try:
        group_index = lowered.index("resourcegroups")
        resource_group = parts[group_index + 1]
    except (ValueError, IndexError):
        resource_group = "Unknown"
    try:
        provider_index = lowered.index("providers")
        resource_type = "/".join(parts[provider_index + 1 : -1])
    except (ValueError, IndexError):
        resource_type = "Unknown"
    return resource_group, resource_type or "Unknown"


def _trend_points(rows: list[dict]) -> list[CostPoint]:
    date_key = "UsageDate" if any(row.get("UsageDate") for row in rows) else "BillingMonth"
    totals = _group_total(rows, date_key)
    if not totals:
        return []
    return [CostPoint(month=_format_month(key), amount=round(value, 2)) for key, value in sorted(totals.items())]


def _format_month(value: object) -> str:
    if isinstance(value, datetime):
        return value.strftime("%b")
    text = str(value)
    if len(text) == 8 and text.isdigit():
        try:
            return datetime.strptime(text, "%Y%m%d").strftime("%b")
        except ValueError:
            pass
    try:
        return datetime.fromisoformat(text.replace("Z", "+00:00")).strftime("%b")
    except ValueError:
        return text[:3]
