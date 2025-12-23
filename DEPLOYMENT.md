# Deployment Guide - DBaaS Operator

## Prerequisites

1. **Kubernetes cluster** (v1.25+)
   - Minikube, Kind, K3s, hoặc cloud K8s cluster
   - kubectl configured và có quyền admin

2. **Tools required:**
   ```bash
   # Verify kubectl
   kubectl version

   # Verify cluster access
   kubectl cluster-info
   ```

3. **Dependencies:**
   - CloudNativePG operator (will be installed)
   - Cert-manager (optional, for webhooks)

## Step 1: Setup CloudNativePG Operator

DBaaS Operator cần CNPG làm underlying operator cho PostgreSQL.

```bash
# Install CNPG operator
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.23/releases/cnpg-1.23.0.yaml

# Verify CNPG installation
kubectl get deployment -n cnpg-system cnpg-controller-manager
kubectl get pods -n cnpg-system

# Wait until CNPG is ready
kubectl wait --for=condition=Available --timeout=300s \
  deployment/cnpg-controller-manager -n cnpg-system
```

## Step 2: Generate and Install CRDs

Generate CRDs from our API definitions:

```bash
# Generate CRDs
make manifests

# Or manually with controller-gen
controller-gen crd:crdVersions=v1 \
  rbac:roleName=manager-role \
  webhook \
  paths="./..." \
  output:crd:artifacts:config=config/crd/bases
```

Install CRDs:

```bash
# Install all CRDs
kubectl apply -f config/crd/bases/

# Verify CRDs are installed
kubectl get crds | grep dbaas.io
```

You should see:
```
backupstorages.dbaas.io
databaseclusters.dbaas.io
databaseengines.dbaas.io
monitoringconfigs.dbaas.io
opsrequests.dbaas.io
```

## Step 3: Build and Deploy Operator

### Option A: Run Locally (for development)

```bash
# Run operator locally (requires kubeconfig)
go run ./main.go
```

### Option B: Deploy to Cluster (recommended)

```bash
# Build Docker image
docker build -t dbaas-operator:latest .

# For Minikube, load image into cluster
minikube image load dbaas-operator:latest

# For Kind
kind load docker-image dbaas-operator:latest

# For remote registry
docker tag dbaas-operator:latest your-registry/dbaas-operator:latest
docker push your-registry/dbaas-operator:latest
```

Create namespace and RBAC:

```bash
# Create namespace
kubectl create namespace dbaas-system

# Apply RBAC
kubectl apply -f config/rbac/role.yaml

# Apply deployment
kubectl apply -f config/manager/manager.yaml
```

Verify operator is running:

```bash
# Check operator pod
kubectl get pods -n dbaas-system

# Check logs
kubectl logs -n dbaas-system deployment/dbaas-operator-controller-manager -f
```

## Step 4: Create Test Resources

### 4.1 Create DatabaseEngine for CNPG

```bash
kubectl apply -f config/samples/databaseengine-cnpg.yaml

# Verify
kubectl get databaseengine
```

### 4.2 Create PostgreSQL Cluster

```bash
kubectl apply -f config/samples/postgresql-cluster.yaml

# Watch the cluster being created
kubectl get databasecluster postgresql-demo -w

# Check CNPG cluster (child resource)
kubectl get cluster postgresql-demo

# Check pods
kubectl get pods -l dbaas.io/cluster=postgresql-demo
```

### 4.3 Check Status

```bash
# Get DatabaseCluster status
kubectl get databasecluster postgresql-demo -o yaml

# Check events
kubectl describe databasecluster postgresql-demo

# Check CNPG cluster status
kubectl get cluster postgresql-demo -o jsonpath='{.status.phase}'
```

## Step 5: Test Operations

### Test Horizontal Scaling

```bash
# Apply scaling operation
kubectl apply -f config/samples/ops-horizontal-scaling.yaml

# Watch operation status
kubectl get opsrequest scale-postgresql-demo -w

# Verify cluster scaled
kubectl get pods -l dbaas.io/cluster=postgresql-demo
```

### Test Backup

```bash
# Apply backup operation
kubectl apply -f config/samples/ops-backup.yaml

# Check backup status
kubectl get opsrequest backup-postgresql-demo

# List CNPG backups
kubectl get backup
```

### Test Upgrade

```bash
# Apply upgrade operation
kubectl apply -f config/samples/ops-upgrade.yaml

# Watch upgrade progress
kubectl get opsrequest upgrade-postgresql-demo -w
```

## Step 6: Verify Everything Works

### Check Database Connectivity

```bash
# Port-forward to database
kubectl port-forward svc/postgresql-demo-rw 5432:5432

# In another terminal, connect with psql
psql -h localhost -U app -d app

# Or use kubectl exec
kubectl exec -it postgresql-demo-1 -- psql -U app -d app
```

### Check Status of All Resources

```bash
# All DatabaseClusters
kubectl get databasecluster

# All OpsRequests
kubectl get opsrequest

# All CNPG clusters
kubectl get cluster

# All pods
kubectl get pods -l dbaas.io/cluster=postgresql-demo
```

## Troubleshooting

### Operator Not Starting

```bash
# Check operator logs
kubectl logs -n dbaas-system deployment/dbaas-operator-controller-manager

# Check RBAC permissions
kubectl auth can-i create clusters.postgresql.cnpg.io --as=system:serviceaccount:dbaas-system:dbaas-operator-controller-manager

# Check CRDs installed
kubectl get crds | grep dbaas
```

### Cluster Not Creating

```bash
# Check DatabaseCluster status
kubectl describe databasecluster postgresql-demo

# Check if CNPG cluster was created
kubectl get cluster postgresql-demo

# Check CNPG operator logs
kubectl logs -n cnpg-system deployment/cnpg-controller-manager
```

### Operations Not Working

```bash
# Check OpsRequest status
kubectl describe opsrequest <ops-name>

# Check operator logs during operation
kubectl logs -n dbaas-system deployment/dbaas-operator-controller-manager -f

# Check if target cluster exists
kubectl get databasecluster <cluster-name>
```

## Clean Up

```bash
# Delete test cluster
kubectl delete databasecluster postgresql-demo

# Delete operations
kubectl delete opsrequest --all

# Uninstall operator
kubectl delete -f config/manager/manager.yaml
kubectl delete -f config/rbac/role.yaml

# Delete CRDs (this will delete all custom resources!)
kubectl delete -f config/crd/bases/

# Uninstall CNPG (optional)
kubectl delete -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.23/releases/cnpg-1.23.0.yaml

# Delete namespace
kubectl delete namespace dbaas-system
```

## Next Steps

1. **Add Backup Storage**: Create BackupStorage CR for S3/GCS
2. **Add Monitoring**: Create MonitoringConfig CR for PMM/Prometheus
3. **Test More Operations**: Vertical scaling, reconfiguration, etc.
4. **Production Setup**:
   - Use proper image registry
   - Configure resource limits
   - Set up monitoring and alerting
   - Configure backup retention

## Quick Start Script

Save this as `quickstart.sh`:

```bash
#!/bin/bash
set -e

echo "🚀 Installing CNPG operator..."
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.23/releases/cnpg-1.23.0.yaml
kubectl wait --for=condition=Available --timeout=300s deployment/cnpg-controller-manager -n cnpg-system

echo "📦 Installing DBaaS CRDs..."
kubectl apply -f config/crd/bases/

echo "🏗️ Creating namespace..."
kubectl create namespace dbaas-system --dry-run=client -o yaml | kubectl apply -f -

echo "🔐 Applying RBAC..."
kubectl apply -f config/rbac/role.yaml

echo "🚢 Deploying operator..."
kubectl apply -f config/manager/manager.yaml

echo "⏳ Waiting for operator to be ready..."
kubectl wait --for=condition=Available --timeout=300s deployment/dbaas-operator-controller-manager -n dbaas-system

echo "📝 Creating test resources..."
kubectl apply -f config/samples/databaseengine-cnpg.yaml
kubectl apply -f config/samples/postgresql-cluster.yaml

echo "✅ Done! Check status with:"
echo "  kubectl get databasecluster postgresql-demo -w"
```

Run it:
```bash
chmod +x quickstart.sh
./quickstart.sh
```
