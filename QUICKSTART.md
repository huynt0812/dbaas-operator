# Quick Start Guide

## TL;DR - Fastest Way to Test

```bash
# 1. Clone and navigate to repo
cd dbaas-operator

# 2. Run quickstart script
./quickstart.sh

# 3. Watch your cluster
kubectl get databasecluster postgresql-demo -w
```

That's it! 🚀

## What the quickstart does:

1. ✅ Installs CloudNativePG operator
2. ✅ Generates and installs CRDs
3. ✅ Deploys DBaaS operator
4. ✅ Creates a PostgreSQL cluster with 3 replicas

## Using Makefile

```bash
# One command to rule them all
make quickstart

# Or step by step
make deploy-cnpg      # Install CNPG
make install          # Install CRDs
make deploy           # Deploy operator
make deploy-samples   # Create test cluster

# Check status
make status

# View logs
make logs

# Clean up
make clean
```

## Manual Steps (if you want control)

### Prerequisites

```bash
# Check you have kubectl and cluster access
kubectl cluster-info
```

### 1. Generate CRDs

```bash
# Install controller-gen
go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0

# Generate
make manifests
```

### 2. Install CNPG Operator

```bash
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.23/releases/cnpg-1.23.0.yaml
```

### 3. Install DBaaS CRDs

```bash
kubectl apply -f config/crd/bases/
```

### 4. Deploy Operator

**Option A: Run locally**
```bash
go run ./main.go
```

**Option B: Deploy to cluster**
```bash
# Build image
docker build -t dbaas-operator:latest .

# Load into cluster (for Minikube/Kind)
minikube image load dbaas-operator:latest
# OR
kind load docker-image dbaas-operator:latest

# Deploy
make deploy
```

### 5. Create Test Resources

```bash
kubectl apply -f config/samples/databaseengine-cnpg.yaml
kubectl apply -f config/samples/postgresql-cluster.yaml
```

## Verify Everything Works

```bash
# Check cluster status
kubectl get databasecluster postgresql-demo

# Should show:
# NAME               ENGINE       VERSION   PHASE   READY   AGE
# postgresql-demo    postgresql   16.0      Ready   true    2m

# Check pods
kubectl get pods -l dbaas.io/cluster=postgresql-demo

# Should show 3 pods running
```

## Test Operations

```bash
# Scale to 5 replicas
kubectl apply -f config/samples/ops-horizontal-scaling.yaml

# Create backup
kubectl apply -f config/samples/ops-backup.yaml

# Upgrade version
kubectl apply -f config/samples/ops-upgrade.yaml

# Check operations status
kubectl get opsrequest
```

## Connect to Database

```bash
# Port forward
kubectl port-forward svc/postgresql-demo-rw 5432:5432 &

# Get password
PASSWORD=$(kubectl get secret postgresql-demo-app \
  -o jsonpath='{.data.password}' | base64 -d)

# Connect
psql -h localhost -U app -d app
# When prompted, enter: $PASSWORD
```

## Troubleshooting

### Operator not starting?

```bash
# Check logs
kubectl logs -n dbaas-system deployment/dbaas-operator-controller-manager
```

### Cluster stuck?

```bash
# Check DatabaseCluster
kubectl describe databasecluster postgresql-demo

# Check CNPG Cluster
kubectl describe cluster postgresql-demo

# Check CNPG operator
kubectl logs -n cnpg-system deployment/cnpg-controller-manager
```

### CRDs not found?

```bash
# Make sure you generated them
make manifests

# Then install
kubectl apply -f config/crd/bases/
```

## Next Steps

- 📖 Read [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment guide
- 🧪 Read [TESTING.md](TESTING.md) for comprehensive testing scenarios
- 🏗️ Read [ARCHITECTURE.md](ARCHITECTURE.md) for architecture details
- 📚 Read [README.md](README.md) for full documentation

## Clean Up

```bash
# Quick cleanup
make clean

# Or manual
kubectl delete databasecluster postgresql-demo
kubectl delete opsrequest --all
make undeploy
make uninstall
```

## Common Commands Cheat Sheet

```bash
# Status
make status

# Logs
make logs

# Deploy everything
make deploy-all

# Deploy samples
make deploy-samples

# Clean everything
make clean

# Help
make help
```

## What's Created?

When you run quickstart, you get:

- **Namespace**: `dbaas-system` - for the operator
- **Namespace**: `cnpg-system` - for CNPG operator
- **CRDs**: 5 custom resource definitions
- **Operator**: DBaaS operator deployment
- **DatabaseEngine**: CNPG engine definition
- **DatabaseCluster**: 3-node PostgreSQL cluster
- **CNPG Cluster**: Underlying PostgreSQL cluster
- **Pods**: 3 PostgreSQL pods
- **Services**: Read-write and read-only endpoints

## Resource Usage

Minimal setup requires approximately:
- **CPU**: ~1 core total
- **Memory**: ~2GB total
- **Storage**: ~10GB for database

Good for:
- ✅ Minikube
- ✅ Kind
- ✅ K3s
- ✅ Docker Desktop
- ✅ Cloud K8s clusters

## Support

Having issues?
1. Check [DEPLOYMENT.md](DEPLOYMENT.md) troubleshooting section
2. Check [TESTING.md](TESTING.md) common issues
3. Review logs: `make logs`
4. Check status: `make status`

## Happy Testing! 🎉
