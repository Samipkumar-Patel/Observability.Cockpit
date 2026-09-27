## Helm Chart Deployment

A Helm chart is provided in `deploy/helm/` for deploying CloudOps Cockpit onto any Kubernetes cluster (including Azure Kubernetes Service).

### Install the Chart
```bash
helm install cloudops-cockpit ./deploy/helm --namespace observability --create-namespace
```

### Customize Values
You can override default settings (such as persistence size or Ingress host) by passing a custom values file or command-line flags:

```bash
helm install cloudops-cockpit ./deploy/helm --namespace observability --set ingress.host="cockpit.yourdomain.com"
```

### Uninstall the Chart
```bash
helm uninstall cloudops-cockpit --namespace observability
```
