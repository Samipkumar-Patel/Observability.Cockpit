let activeCurrency = "USD";
const money = value => new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: activeCurrency,
  maximumFractionDigits: 0,
}).format(value);

async function loadOverview() {
  const response = await fetch("/api/overview");
  const data = await response.json();
  if (!response.ok) throw new Error(data.detail || "Unable to load cockpit data");
  activeCurrency = data.currency || "USD";
  document.querySelector("#source-badge").textContent = data.data_source === "azure" ? "AZURE DATA" : "DEMO DATA";
  document.querySelector("#subscription").textContent = data.subscription_name;
  document.querySelector("#current-cost").textContent = money(data.current_month_cost);
  document.querySelector("#change").textContent = `↑ ${data.change_percent}%`;
  document.querySelector("#forecast").textContent = money(data.forecasted_month_cost);

  const max = Math.max(...data.cost_trend.map(point => point.amount));
  document.querySelector("#trend").innerHTML = data.cost_trend.map(point => `
    <div class="bar-column"><span>${money(point.amount)}</span><div class="bar" style="height: ${(point.amount / max) * 100}%"></div><small>${point.month}</small></div>
  `).join("");

  document.querySelector("#services").innerHTML = data.service_costs.map(item => `
    <div class="service-row"><div><strong>${item.service}</strong><span class="muted">${item.share}%</span></div><div class="service-track"><i style="width: ${item.share}%"></i></div><b>${money(item.amount)}</b></div>
  `).join("");

  const totalSavings = data.recommendations.reduce((sum, item) => sum + item.estimated_monthly_savings, 0);
  document.querySelector("#savings").textContent = `${money(totalSavings)} possible monthly savings`;
  document.querySelector("#recommendations").innerHTML = data.recommendations.map(item => `
    <div class="recommendation"><div class="rec-icon">↘</div><div><strong>${item.title}</strong><p>${item.summary}</p><small>${item.resource_name} · ${item.risk} risk · ${item.category}</small></div><b>${money(item.estimated_monthly_savings)}<small>/ month</small></b></div>
  `).join("");

  document.querySelector("#resources").innerHTML = data.resources.map(item => `
    <tr><td><strong>${item.name}</strong><small>${item.resource_group}</small></td><td>${item.resource_type}</td><td><span class="tag">${item.environment}</span></td><td>${item.utilization === null ? "N/A" : `${item.utilization}%`}</td><td><strong>${money(item.monthly_cost)}</strong></td></tr>
  `).join("");
}

loadOverview().catch(error => {
  document.querySelector("#subscription").textContent = error.message;
});
