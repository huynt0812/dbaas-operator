# Testing Guide - DBaaS Operator

## Quick Test (Recommended)

### Method 1: Using Script

```bash
# Make script executable
chmod +x quickstart.sh

# Run quickstart
./quickstart.sh
```

### Method 2: Using Makefile

```bash
# Install everything at once
make quickstart

# Or step by step
make deploy-cnpg      # Install CNPG operator
make install          # Install CRDs
make deploy           # Deploy DBaaS operator
make deploy-samples   # Create test cluster
```

## Step-by-Step Manual Testing

### 1. Generate CRDs First

```bash
# Install controller-gen if not already installed
go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0

# Generate CRDs
make manifests

# Or use controller-gen directly
$(go env GOPATH)/bin/controller-gen \
  crd:crdVersions=v1 \
  rbac:roleName=manager-role \
  webhook \
  paths="./..." \
  output:crd:artifacts:config=config/crd/bases
```

### 2. Install CNPG Operator

```bash
# Install CNPG
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.23/releases/cnpg-1.23.0.yaml

# Wait for CNPG to be ready
kubectl wait --for=condition=Available --timeout=300s \
  deployment/cnpg-controller-manager -n cnpg-system

# Verify
kubectl get pods -n cnpg-system
```

### 3. Install DBaaS Operator CRDs

```bash
# Apply CRDs
kubectl apply -f config/crd/bases/

# Verify CRDs installed
kubectl get crds | grep dbaas.io
```

Expected output:
```
backupstorages.dbaas.io
databaseclusters.dbaas.io
databaseengines.dbaas.io
monitoringconfigs.dbaas.io
opsrequests.dbaas.io
```

### 4. Deploy Operator

**Option A: Run locally (development)**

```bash
# Run operator locally with your kubeconfig
go run ./main.go
```

**Option B: Deploy to cluster (production-like)**

```bash
# Build image
docker build -t dbaas-operator:latest .

# Load into cluster (for local K8s)
# For Minikube:
minikube image load dbaas-operator:latest

# For Kind:
kind load docker-image dbaas-operator:latest

# Deploy
make deploy

# Check deployment
kubectl get pods -n dbaas-system
kubectl logs -n dbaas-system deployment/dbaas-operator-controller-manager -f
```

### 5. Create Test Resources

```bash
# Create DatabaseEngine
kubectl apply -f config/samples/databaseengine-cnpg.yaml

# Verify
kubectl get databaseengine cnpg-postgresql
```

```bash
# Create PostgreSQL cluster
kubectl apply -f config/samples/postgresql-cluster.yaml

# Watch creation
kubectl get databasecluster postgresql-demo -w
```

### 6. Verify Cluster Creation

```bash
# Check DatabaseCluster
kubectl get databasecluster postgresql-demo -o yaml

# Check CNPG Cluster (child resource)
kubectl get cluster postgresql-demo

# Check pods
kubectl get pods -l dbaas.io/cluster=postgresql-demo

# Check services
kubectl get svc -l dbaas.io/cluster=postgresql-demo
```

Expected status:
```yaml
status:
  phase: Ready
  database:
    ready: true
    instances: 3
    readyInstances: 3
    primaryInstance: postgresql-demo-1
```

## Testing Day-2 Operations

### Test 1: Horizontal Scaling

```bash
# Scale from 3 to 5 replicas
kubectl apply -f config/samples/ops-horizontal-scaling.yaml

# Watch operation
kubectl get opsrequest scale-postgresql-demo -w

# Verify scaling
kubectl get pods -l dbaas.io/cluster=postgresql-demo
# Should see 5 pods now

# Check operation status
kubectl get opsrequest scale-postgresql-demo -o yaml
```

Expected:
```yaml
status:
  phase: Succeeded
  actionLog:
    - action: HorizontalScaling
      status: Succeeded
```

### Test 2: Vertical Scaling

Create `test-vertical-scaling.yaml`:
```yaml
apiVersion: dbaas.io/v1
kind: OpsRequest
metadata:
  name: vscale-postgresql-demo
spec:
  clusterRef:
    name: postgresql-demo
  type: VerticalScaling
  verticalScaling:
    resources:
      requests:
        cpu: "2"
        memory: 4Gi
      limits:
        cpu: "4"
        memory: 8Gi
```

```bash
# Apply
kubectl apply -f test-vertical-scaling.yaml

# Watch
kubectl get opsrequest vscale-postgresql-demo -w

# Verify resources changed
kubectl get cluster postgresql-demo -o jsonpath='{.spec.resources}'
```

### Test 3: Reconfiguration

Create `test-reconfigure.yaml`:
```yaml
apiVersion: dbaas.io/v1
kind: OpsRequest
metadata:
  name: reconfig-postgresql-demo
spec:
  clusterRef:
    name: postgresql-demo
  type: Reconfiguring
  reconfiguring:
    config:
      - name: max_connections
        value: "300"
      - name: shared_buffers
        value: "2GB"
```

```bash
# Apply
kubectl apply -f test-reconfigure.yaml

# Verify config changed
kubectl get cluster postgresql-demo \
  -o jsonpath='{.spec.postgresConfiguration.parameters}'
```

### Test 4: Backup

```bash
# Create backup
kubectl apply -f config/samples/ops-backup.yaml

# Watch backup
kubectl get opsrequest backup-postgresql-demo -w

# List CNPG backups
kubectl get backup

# Check backup details
kubectl describe backup <backup-name>
```

### Test 5: Upgrade

```bash
# Upgrade PostgreSQL version
kubectl apply -f config/samples/ops-upgrade.yaml

# Watch upgrade
kubectl get opsrequest upgrade-postgresql-demo -w

# Verify version changed
kubectl get cluster postgresql-demo -o jsonpath='{.spec.imageName}'
```

### Test 6: Restart

Create `test-restart.yaml`:
```yaml
apiVersion: dbaas.io/v1
kind: OpsRequest
metadata:
  name: restart-postgresql-demo
spec:
  clusterRef:
    name: postgresql-demo
  type: Restart
```

```bash
# Apply
kubectl apply -f test-restart.yaml

# Watch pods restart
kubectl get pods -l dbaas.io/cluster=postgresql-demo -w
```

## Database Connectivity Tests

### Test Connection

```bash
# Port-forward to primary
kubectl port-forward svc/postgresql-demo-rw 5432:5432 &

# Connect with psql (in another terminal)
psql -h localhost -U app -d app
# Password: check secret postgresql-demo-app

# Get password
kubectl get secret postgresql-demo-app \
  -o jsonpath='{.data.password}' | base64 -d
```

### Test from Inside Cluster

```bash
# Create test pod
kubectl run -it --rm psql-client \
  --image=postgres:16 \
  --restart=Never \
  -- psql -h postgresql-demo-rw -U app -d app

# Inside psql:
SELECT version();
\l
\dt
```

### Test Read/Write

```bash
# Connect to primary (read-write)
kubectl exec -it postgresql-demo-1 -- psql -U app -d app

# Create test table
CREATE TABLE test (id SERIAL PRIMARY KEY, data TEXT);
INSERT INTO test (data) VALUES ('Hello from DBaaS!');
SELECT * FROM test;

# Connect to replica (read-only)
kubectl exec -it postgresql-demo-2 -- psql -U app -d app

# Should be able to read
SELECT * FROM test;

# Should NOT be able to write
INSERT INTO test (data) VALUES ('This will fail');
-- ERROR: cannot execute INSERT in a read-only transaction
```

## Performance Testing

### Simple Load Test

```bash
# Install pgbench
kubectl exec -it postgresql-demo-1 -- \
  pgbench -i -s 10 -U app -d app

# Run benchmark
kubectl exec -it postgresql-demo-1 -- \
  pgbench -c 10 -j 2 -t 1000 -U app -d app
```

## Monitoring Tests

### Check Metrics

```bash
# If monitoring is enabled
kubectl get servicemonitor

# Check Prometheus targets
# (requires Prometheus operator)
```

### Check Logs

```bash
# Operator logs
kubectl logs -n dbaas-system \
  deployment/dbaas-operator-controller-manager -f

# Database logs
kubectl logs postgresql-demo-1

# CNPG operator logs
kubectl logs -n cnpg-system \
  deployment/cnpg-controller-manager
```

## Troubleshooting Tests

### Test 1: Operator Crash Recovery

```bash
# Delete operator pod
kubectl delete pod -n dbaas-system \
  -l control-plane=controller-manager

# Verify it restarts and reconciles
kubectl get pods -n dbaas-system -w
kubectl get databasecluster
```

### Test 2: Database Pod Failure

```bash
# Delete a replica pod
kubectl delete pod postgresql-demo-2

# Verify it's recreated
kubectl get pods -l dbaas.io/cluster=postgresql-demo -w

# Check cluster status
kubectl get databasecluster postgresql-demo
```

### Test 3: Primary Failover

```bash
# Delete primary pod
PRIMARY=$(kubectl get cluster postgresql-demo \
  -o jsonpath='{.status.currentPrimary}')
kubectl delete pod $PRIMARY

# Watch failover
kubectl get cluster postgresql-demo -w

# Verify new primary elected
kubectl get cluster postgresql-demo \
  -o jsonpath='{.status.currentPrimary}'
```

## Status Verification

```bash
# Use make target
make status

# Or manually:
echo "=== DatabaseClusters ==="
kubectl get databasecluster

echo "=== CNPG Clusters ==="
kubectl get cluster

echo "=== Pods ==="
kubectl get pods -l dbaas.io/cluster=postgresql-demo

echo "=== OpsRequests ==="
kubectl get opsrequest

echo "=== Backups ==="
kubectl get backup
```

## Cleanup

```bash
# Using make
make clean

# Or manually:
kubectl delete databasecluster postgresql-demo
kubectl delete opsrequest --all
kubectl delete databaseengine cnpg-postgresql

# Uninstall operator
make undeploy
make uninstall

# Uninstall CNPG (optional)
kubectl delete -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.23/releases/cnpg-1.23.0.yaml
```

## Common Issues

### Issue 1: CRDs Not Found

```bash
# Generate CRDs first
make manifests

# Then install
kubectl apply -f config/crd/bases/
```

### Issue 2: Operator Not Starting

```bash
# Check logs
kubectl logs -n dbaas-system \
  deployment/dbaas-operator-controller-manager

# Check RBAC
kubectl auth can-i create clusters.postgresql.cnpg.io \
  --as=system:serviceaccount:dbaas-system:dbaas-operator-controller-manager
```

### Issue 3: Cluster Stuck in Initializing

```bash
# Check DatabaseCluster
kubectl describe databasecluster postgresql-demo

# Check CNPG Cluster
kubectl describe cluster postgresql-demo

# Check CNPG operator
kubectl logs -n cnpg-system \
  deployment/cnpg-controller-manager
```

### Issue 4: Operations Failing

```bash
# Check OpsRequest
kubectl describe opsrequest <ops-name>

# Check operator logs during operation
kubectl logs -n dbaas-system \
  deployment/dbaas-operator-controller-manager -f

# Check if cluster exists
kubectl get databasecluster
```

## Test Checklist

- [ ] CRDs installed successfully
- [ ] CNPG operator running
- [ ] DBaaS operator running
- [ ] DatabaseEngine created
- [ ] DatabaseCluster created and Ready
- [ ] CNPG Cluster created
- [ ] Pods running (3 replicas)
- [ ] Can connect to database
- [ ] Horizontal scaling works
- [ ] Vertical scaling works
- [ ] Reconfiguration works
- [ ] Backup works
- [ ] Restart works
- [ ] Operator crash recovery works
- [ ] Pod failure recovery works
- [ ] Primary failover works
