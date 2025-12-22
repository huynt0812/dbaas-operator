# DBaaS Operator Architecture

## Overview

DBaaS Operator provides a unified Database-as-a-Service platform on Kubernetes by abstracting multiple database operators (CNPG, Percona, Strimzi, etc.) behind a consistent API.

## Design Principles

1. **Abstraction**: Client only interacts with `DatabaseCluster` CRD
2. **Extensibility**: Pluggable provider architecture for different database engines
3. **Consistency**: Uniform API across PostgreSQL, MongoDB, MySQL, Kafka, etc.
4. **Delegation**: Leverage existing operators (CNPG, Percona) for actual database management
5. **Operator Pattern**: Follow Kubernetes operator best practices

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         User/Client                          │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ kubectl apply
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                    DatabaseCluster CR                        │
│  - engine: postgresql                                        │
│  - clusterSize: 3                                            │
│  - storage: 10Gi                                             │
│  - config: [{name: max_connections, value: "200"}]           │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ watches
                        ▼
┌─────────────────────────────────────────────────────────────┐
│            DatabaseCluster Controller                        │
│                                                              │
│  1. PreReconcileHook()                                       │
│  2. GetApplier() → Applier                                   │
│  3. Apply Transformations:                                   │
│     - Metadata()                                             │
│     - Engine()                                               │
│     - Proxy()                                                │
│     - Monitoring()                                           │
│     - Backup()                                               │
│     - DataSource()                                           │
│  4. Create/Update Child Cluster                              │
│  5. Sync Status                                              │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ delegates to
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                   Provider Factory                           │
│                                                              │
│  GetProvider(engineType) → Provider                          │
│    - postgresql → CNPGProvider                               │
│    - mongodb → PerconaMongoProvider                          │
│    - mysql → PerconaMySQLProvider                            │
│    - kafka → StrimziProvider                                 │
└───────────────────────┬─────────────────────────────────────┘
                        │
        ┌───────────────┴───────────────┐
        │                               │
        ▼                               ▼
┌──────────────────┐          ┌──────────────────┐
│  CNPGProvider    │          │ Other Providers  │
│                  │          │                  │
│ - GetApplier()   │          │ (Future)         │
│ - Status()       │          │                  │
│ - Operations()   │          │                  │
│ - Cleanup()      │          │                  │
└────────┬─────────┘          └──────────────────┘
         │
         │ creates
         ▼
┌─────────────────────────────────────────────────────────────┐
│                   CNPG Cluster CR                            │
│  (CloudNativePG Operator manages this)                      │
│                                                              │
│  - instances: 3                                              │
│  - imageName: postgres:16.0                                  │
│  - storage: 10Gi                                             │
│  - postgresqlParameters:                                     │
│      max_connections: "200"                                  │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ managed by
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              CloudNativePG Operator                          │
│                                                              │
│  Creates and manages:                                        │
│  - PostgreSQL Pods                                           │
│  - Services                                                  │
│  - PVCs                                                      │
│  - Backups                                                   │
└─────────────────────────────────────────────────────────────┘
```

## Day-2 Operations Flow

```
┌─────────────────────────────────────────────────────────────┐
│                      OpsRequest CR                           │
│  - clusterRef: postgresql-demo                               │
│  - type: HorizontalScaling                                   │
│  - horizontalScaling:                                        │
│      replicas: 5                                             │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ watches
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              OpsRequest Controller                           │
│                                                              │
│  1. Validate cluster exists                                  │
│  2. Get Provider for cluster engine                          │
│  3. Get OperationsHandler                                    │
│  4. Execute operation:                                       │
│     switch ops.Type:                                         │
│       case HorizontalScaling:                                │
│         handler.HorizontalScaling(ctx, cluster, ops)         │
│       case Backup:                                           │
│         handler.Backup(ctx, cluster, ops)                    │
│       ...                                                    │
│  5. Poll GetStatus() periodically                            │
│  6. Update OpsRequest.Status                                 │
│  7. Add Action Logs                                          │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ delegates to
                        ▼
┌─────────────────────────────────────────────────────────────┐
│           CNPGOperationsHandler                              │
│                                                              │
│  HorizontalScaling():                                        │
│    1. Get CNPG Cluster                                       │
│    2. Update cluster.spec.instances = 5                      │
│    3. Update CNPG Cluster                                    │
│                                                              │
│  GetStatus():                                                │
│    1. Get CNPG Cluster                                       │
│    2. Check cluster.status.phase                             │
│    3. Return OpsRequestStatus                                │
└─────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

### 1. Core Operator (DBaaS)

**Responsibilities:**
- Define and validate CRDs (DatabaseCluster, OpsRequest, etc.)
- Implement core controllers
- Provide Provider and Applier interfaces
- Manage RBAC and permissions
- Handle parent cluster lifecycle
- Common utilities and status aggregation

**Does NOT:**
- Know database-specific implementation details
- Directly create database Pods/Services
- Implement database operations

### 2. Provider Interface

**Responsibilities:**
- Bridge between DBaaS and underlying operator
- Transform specs (parent → child)
- Transform status (child → parent)
- Implement operations (start, stop, scale, etc.)
- Handle cleanup on deletion
- Provider-specific hooks

**Key Methods:**
```go
type Provider interface {
    GetApplier(cluster) (Applier, error)
    Status(ctx, cluster) (*Status, error)
    Cleanup(ctx, cluster) error
    PreReconcileHook(ctx, cluster) (requeue, error)
    Operations() OperationsHandler
}
```

### 3. Applier Interface

**Responsibilities:**
- Build child cluster specification
- Apply transformations in sequence
- Handle engine-specific configuration
- Support custom config parameters

**Transformation Pipeline:**
```
DatabaseCluster Spec
        ↓
    Metadata() → labels, annotations
        ↓
    Engine() → image, version, resources, storage
        ↓
    Proxy() → proxy/pooler configuration
        ↓
    Monitoring() → PMM/Prometheus setup
        ↓
    PodSchedulingPolicy() → affinity, tolerations
        ↓
    Backup() → backup schedule, retention
        ↓
    DataSource() → restore, clone configuration
        ↓
    GetResult() → Child Cluster CR
```

### 4. Operations Handler

**Responsibilities:**
- Implement day-2 operations
- Translate operations to provider-specific actions
- Track operation progress
- Return operation status

**Supported Operations:**
- **Lifecycle**: Start, Stop, Restart
- **Scaling**: HorizontalScaling, VerticalScaling, VolumeExpansion
- **Configuration**: Reconfiguring, Upgrade
- **Data Management**: Backup, Restore
- **HA**: Switchover, RebuildInstance
- **Custom**: Provider-specific operations

## Data Flow

### Cluster Creation

```
1. User creates DatabaseCluster CR
   ↓
2. DatabaseCluster Controller watches event
   ↓
3. Controller calls Provider.GetApplier()
   ↓
4. Applier transforms spec:
   - Metadata() → labels/annotations
   - Engine() → database config
   - Backup() → backup settings
   - ... (all transformations)
   ↓
5. Applier.GetResult() returns child cluster CR
   ↓
6. Controller creates/updates child cluster
   ↓
7. Child operator (CNPG) creates actual resources
   ↓
8. Controller calls Provider.Status()
   ↓
9. Provider maps child status → parent status
   ↓
10. Controller updates DatabaseCluster.Status
```

### Status Synchronization

```
Child Cluster Status (CNPG)
        ↓
Provider.Status()
        ↓
Map fields:
  - phase
  - instances/readyInstances
  - primary instance
  - instance roles
  - endpoints
  - backup status
  - monitoring status
        ↓
DatabaseCluster Status
        ↓
User sees unified status
```

### Operation Execution

```
1. User creates OpsRequest CR
   ↓
2. OpsRequest Controller validates cluster exists
   ↓
3. Controller gets Provider for cluster engine
   ↓
4. Controller gets OperationsHandler
   ↓
5. Controller calls handler.{Operation}()
   ↓
6. Handler updates child cluster (e.g., change replicas)
   ↓
7. Child operator performs actual operation
   ↓
8. Controller polls handler.GetStatus()
   ↓
9. Controller updates OpsRequest.Status
   ↓
10. Operation completes, status shows Success/Failed
```

## Extension Points

### Adding a New Database Engine

1. **Create Provider Package**: `pkg/provider/{engine}/`
2. **Implement Provider Interface**:
   ```go
   type MyEngineProvider struct {
       client client.Client
       scheme *runtime.Scheme
   }
   ```
3. **Implement Applier**:
   ```go
   type MyEngineApplier struct {
       cluster *DatabaseCluster
       childCluster *MyEngineCluster
   }
   ```
4. **Implement Operations Handler**:
   ```go
   type MyEngineOperations struct {
       client client.Client
   }
   ```
5. **Register in Factory**:
   ```go
   func (f *Factory) GetProvider(engineType) {
       case "myengine":
           return NewMyEngineProvider(...)
   }
   ```

### Custom Configuration

Providers can accept custom configuration via `spec.config`:

```yaml
spec:
  config:
    - name: engine_specific_param
      value: "custom_value"
```

Provider's `Engine()` method reads and applies these:

```go
func (a *Applier) Engine() {
    for _, cfg := range a.cluster.Spec.Config {
        // Apply to child cluster
        childCluster.Spec.Config[cfg.Name] = cfg.Value
    }
}
```

## Reconciliation Loop

### DatabaseCluster Controller

```
┌─────────────────────────────────────────┐
│ Event: DatabaseCluster created/updated  │
└──────────────┬──────────────────────────┘
               ▼
        ┌──────────────┐
        │ Get Cluster  │
        └──────┬───────┘
               ▼
        ┌──────────────┐
        │ Get Provider │
        └──────┬───────┘
               ▼
     ┌──────────────────┐
     │ PreReconcileHook │
     └──────┬───────────┘
            │
            ├─ Requeue? ─→ Requeue after N seconds
            │
            ▼
     ┌──────────────┐
     │ Get Applier  │
     └──────┬───────┘
            ▼
  ┌─────────────────────┐
  │ Apply Transformations│
  └──────┬──────────────┘
         ▼
  ┌─────────────────────┐
  │ Create/Update Child  │
  └──────┬──────────────┘
         ▼
  ┌─────────────────┐
  │ Update Status   │
  └──────┬──────────┘
         ▼
  ┌─────────────────┐
  │ Requeue (30s)   │
  └─────────────────┘
```

### OpsRequest Controller

```
┌──────────────────────────────────────┐
│ Event: OpsRequest created/updated    │
└──────────────┬───────────────────────┘
               ▼
        ┌──────────────┐
        │ Get OpsReq   │
        └──────┬───────┘
               ▼
     ┌─────────────────┐
     │ Already Done?   │
     └──────┬──────────┘
            │
            ├─ Yes ─→ Check TTL ─→ Delete if expired
            │
            ▼ No
     ┌──────────────────┐
     │ Get Target Cluster│
     └──────┬───────────┘
            ▼
     ┌──────────────┐
     │ Get Provider │
     └──────┬───────┘
            ▼
  ┌──────────────────────┐
  │ Get OperationsHandler│
  └──────┬───────────────┘
         ▼
  ┌─────────────────┐
  │ Execute Operation│
  └──────┬──────────┘
         ▼
  ┌─────────────────┐
  │ Get Status      │
  └──────┬──────────┘
         ▼
  ┌─────────────────┐
  │ Update Status   │
  └──────┬──────────┘
         │
         ├─ Running ─→ Requeue (5s)
         │
         ▼ Done
      Complete
```

## Future Enhancements

1. **Multi-Cluster Support**: Manage databases across multiple Kubernetes clusters
2. **Advanced Scheduling**: Intelligent placement based on workload
3. **Cost Optimization**: Right-sizing recommendations
4. **Auto-Scaling**: Horizontal/vertical auto-scaling based on metrics
5. **Advanced Backup**: Incremental backups, backup verification
6. **Migration Tools**: Database migration and upgrade automation
7. **Observability**: Enhanced metrics, traces, and dashboards
8. **GitOps Integration**: ArgoCD/Flux integration
9. **Multi-Tenancy**: Namespace isolation and quotas
10. **CLI Tool**: Command-line tool for database operations

## References

- [Kubernetes Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- [CloudNativePG Documentation](https://cloudnative-pg.io/)
- [Kubebuilder Book](https://book.kubebuilder.io/)
