# Deployment & Execution Guide

CloudOps Cockpit can be deployed locally, via Docker, or onto Azure Kubernetes Service (AKS).

---

## 1. Local Development

1. Run the server directly:
   ```bash
   go run main.go
   ```
2. Open `http://localhost:8080`.

---

## 2. Docker & Docker Compose

Run the platform as a container with persistent storage in a single command:

```bash
docker compose up --build -d
```

Access the UI at `http://localhost:8080`. To stop the container:
```bash
docker compose down
```

---

## 3. Azure Kubernetes Service (AKS) Deployment

Deploy CloudOps Cockpit to AKS with persistent storage and Azure Application Gateway Ingress.

### Apply the Kubernetes Manifests
```bash
kubectl apply -f deploy/kubernetes.yaml
```

This creates:
- Namespace `observability`
- A PersistentVolumeClaim (`cloudops-pvc`) for SQLite storage
- Deployment (`cloudops-cockpit`) with liveness and readiness probes
- ClusterIP Service (`cloudops-service`)
- Ingress (`cloudops-ingress`) configured for Azure Application Gateway (`azure/application-gateway`)

### Check Deployment Status
```bash
kubectl get all -n observability
```

### Port Forwarding (Quick Test without Ingress)
```bash
kubectl port-forward svc/cloudops-service 8080:80 -n observability
```
Open `http://localhost:8080`.
