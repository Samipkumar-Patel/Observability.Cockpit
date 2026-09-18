from pydantic import BaseModel, Field


class CostPoint(BaseModel):
    month: str
    amount: float = Field(ge=0)


class ServiceCost(BaseModel):
    service: str
    amount: float = Field(ge=0)
    share: float = Field(ge=0, le=100)


class ResourceCost(BaseModel):
    name: str
    resource_type: str
    resource_group: str
    monthly_cost: float = Field(ge=0)
    utilization: int | None = Field(default=None, ge=0, le=100)
    environment: str


class Recommendation(BaseModel):
    id: str
    title: str
    summary: str
    resource_name: str
    estimated_monthly_savings: float = Field(ge=0)
    risk: str
    category: str


class Overview(BaseModel):
    subscription_name: str
    currency: str
    current_month_cost: float = Field(ge=0)
    previous_month_cost: float = Field(ge=0)
    change_percent: float
    forecasted_month_cost: float = Field(ge=0)
    cost_trend: list[CostPoint]
    service_costs: list[ServiceCost]
    resources: list[ResourceCost]
    recommendations: list[Recommendation]
