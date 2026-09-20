from time import monotonic

from .models import CostPoint, Overview, Recommendation, ResourceCost, ServiceCost

_azure_overview_cache: dict[str, tuple[float, Overview]] = {}
_AZURE_CACHE_SECONDS = 60


def _recommendations(resources: list[ResourceCost]) -> list[Recommendation]:
    recommendations = []
    for resource in resources:
        if resource.utilization is not None and resource.utilization < 10:
            savings = round(resource.monthly_cost * 0.45, 2)
            recommendations.append(
                Recommendation(
                    id=f"rightsizing-{resource.name}",
                    title="Review underutilized compute",
                    summary=f"Average utilization is {resource.utilization}%. Consider resizing or scheduling this resource.",
                    resource_name=resource.name,
                    estimated_monthly_savings=savings,
                    risk="Low",
                    category="Rightsizing",
                )
            )
    return recommendations


def get_demo_overview() -> Overview:
    resources = [
        ResourceCost(
            name="prod-api-vm",
            resource_type="Virtual Machine",
            resource_group="rg-production",
            monthly_cost=312.40,
            utilization=7,
            environment="Production",
        ),
        ResourceCost(
            name="orders-db",
            resource_type="Azure Database for PostgreSQL",
            resource_group="rg-production",
            monthly_cost=428.10,
            utilization=46,
            environment="Production",
        ),
        ResourceCost(
            name="staging-api-vm",
            resource_type="Virtual Machine",
            resource_group="rg-staging",
            monthly_cost=146.80,
            utilization=3,
            environment="Staging",
        ),
        ResourceCost(
            name="logs-workspace",
            resource_type="Log Analytics Workspace",
            resource_group="rg-platform",
            monthly_cost=188.70,
            utilization=None,
            environment="Shared",
        ),
    ]
    recommendations = _recommendations(resources)
    return Overview(
        data_source="demo",
        subscription_name="Contoso Engineering (demo)",
        currency="USD",
        current_month_cost=1076.00,
        previous_month_cost=934.00,
        change_percent=15.2,
        forecasted_month_cost=1218.00,
        cost_trend=[
            CostPoint(month="Oct", amount=812),
            CostPoint(month="Nov", amount=856),
            CostPoint(month="Dec", amount=934),
            CostPoint(month="Jan", amount=1076),
        ],
        service_costs=[
            ServiceCost(service="Compute", amount=459.20, share=42.7),
            ServiceCost(service="Databases", amount=428.10, share=39.8),
            ServiceCost(service="Monitoring", amount=188.70, share=17.5),
        ],
        resources=resources,
        recommendations=recommendations,
    )


def get_overview(data_source: str | None = None, subscription_id: str | None = None) -> Overview:
    from .azure_provider import AzureCostProvider
    from .config import get_settings

    settings = get_settings()
    selected_source = (data_source or settings.data_source).lower()
    if selected_source not in {"demo", "azure"}:
        raise ValueError("data_source must be either 'demo' or 'azure'")
    if selected_source == "demo":
        return get_demo_overview()
    selected_subscription = subscription_id or settings.subscription_id
    if not selected_subscription:
        raise ValueError("AZURE_SUBSCRIPTION_ID is required when COCKPIT_DATA_SOURCE=azure")
    cached = _azure_overview_cache.get(selected_subscription)
    if cached and monotonic() - cached[0] < _AZURE_CACHE_SECONDS:
        return cached[1]
    overview = AzureCostProvider(selected_subscription).get_overview()
    _azure_overview_cache[selected_subscription] = (monotonic(), overview)
    return overview
