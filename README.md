# DBaaS Operator

A Kubernetes operator for managing database clusters with a unified API across different database engines. This operator provides database-as-a-service capabilities with support for PostgreSQL (via CloudNativePG), and extensible architecture for MongoDB, MySQL, Kafka, and more.

## Features

- **Unified API**: Single CRD (`DatabaseCluster`) for managing different database types
- **Multi-Engine Support**: Pluggable provider architecture supporting multiple database engines
- **Day-2 Operations**: Comprehensive operations support via `OpsRequest` CRD
- **Automatic Backup**: Scheduled backups with configurable retention policies
- **High Availability**: Built-in support for clustering and replication
- **Monitoring Integration**: PMM (Percona Monitoring and Management) and Prometheus support
- **CNPG Integration**: Production-ready PostgreSQL support via CloudNativePG

## Architecture

### Custom Resource Definitions (CRDs)

1. **DatabaseCluster**: Main CRD for database cluster management
   - Supports PostgreSQL, MongoDB, MySQL, Kafka (extensible)
   - Configures cluster size, storage, resources, backup, monitoring
   - Engine-specific configuration via key-value config array

2. **DatabaseEngine**: Defines available database operators and versions
   - Maps to specific operators (CNPG, Percona, Strimzi, etc.)
   - Lists supported versions and features
   - Provides default configuration

3. **BackupStorage**: Manages backup storage locations and credentials
   - Supports S3, GCS, Azure, NFS
   - Centralizes backup storage configuration

4. **MonitoringConfig**: Configures monitoring integration
   - PMM, Prometheus, Datadog, New Relic support
   - Reusable across multiple clusters

5. **OpsRequest**: Handles day-2 operations
   - Start, Stop, Restart, Switchover
   - HorizontalScaling, VerticalScaling, VolumeExpansion
   - Reconfiguring, Upgrade, Backup, Restore
   - RebuildInstance, Custom operations

### Provider Architecture

The operator uses a provider pattern to support different database engines:

```
DatabaseCluster (Parent) → Provider → Engine Cluster (Child)
                    ↓
         ┌──────────┴──────────┐
         │                     │
    Applier              Operations
         │                     │
    (Spec Transform)    (Day-2 Ops)
```

#### Provider Interface

```go
type Provider interface {
    GetApplier(cluster *DatabaseCluster) (Applier, error)
    Status(ctx, cluster) (*DatabaseClusterStatus, error)
    Cleanup(ctx, cluster) error
    PreReconcileHook(ctx, cluster) (requeueAfter int, err error)
    Operations() OperationsHandler
}
```

#### Applier Interface

Transforms parent cluster spec to child cluster spec:

```go
type Applier interface {
    Metadata() (labels, annotations, error)
    Engine() (Object, error)
    Proxy() (Object, error)
    Monitoring() (Object, error)
    PodSchedulingPolicy() error
    Backup() error
    DataSource() error
    DataImport() error
    GetResult() Object
}
```

#### Operations Handler

Implements day-2 operations:

```go
type OperationsHandler interface {
    Start(ctx, cluster, ops) error
    Stop(ctx, cluster, ops) error
    Restart(ctx, cluster, ops) error
    Switchover(ctx, cluster, ops) error
    HorizontalScaling(ctx, cluster, ops) error
    VerticalScaling(ctx, cluster, ops) error
    VolumeExpansion(ctx, cluster, ops) error
    Reconfigure(ctx, cluster, ops) error
    Upgrade(ctx, cluster, ops) error
    Backup(ctx, cluster, ops) error
    Restore(ctx, cluster, ops) error
    Expose(ctx, cluster, ops) error
    RebuildInstance(ctx, cluster, ops) error
    Custom(ctx, cluster, ops) error
    GetStatus(ctx, cluster, ops) (*OpsRequestStatus, error)
}
```

## CNPG Provider Implementation

The CloudNativePG (CNPG) provider demonstrates the full implementation:

- **Spec Transformation**: Converts `DatabaseCluster` spec to CNPG `Cluster` spec
- **Status Mapping**: Maps CNPG cluster status to `DatabaseCluster` status
- **Operations**: Full support for all day-2 operations
- **Backup/Restore**: Integration with CNPG backup and PITR features
- **Monitoring**: PMM and Prometheus integration

## Quick Start

### Prerequisites

- Kubernetes cluster (1.25+)
- kubectl configured
- CloudNativePG operator installed (for PostgreSQL)

### Installation

1. Install CRDs:
```bash
make install
```

2. Deploy the operator:
```bash
make deploy
```

3. Create a PostgreSQL cluster:
```bash
kubectl apply -f config/samples/postgresql-cluster.yaml
```

### Example: Create a PostgreSQL Cluster

```yaml
apiVersion: dbaas.io/v1
kind: DatabaseCluster
metadata:
  name: my-postgres
spec:
  engine:
    type: postgresql
    version: "16.0"
  clusterSize: 3
  storage:
    size: 10Gi
  resources:
    requests:
      cpu: "1"
      memory: 2Gi
  backup:
    enabled: true
    schedule: "0 2 * * *"
  config:
    - name: max_connections
      value: "200"
```

### Example: Scale the Cluster

```yaml
apiVersion: dbaas.io/v1
kind: OpsRequest
metadata:
  name: scale-postgres
spec:
  clusterRef:
    name: my-postgres
  type: HorizontalScaling
  horizontalScaling:
    replicas: 5
```

### Example: Backup the Cluster

```yaml
apiVersion: dbaas.io/v1
kind: OpsRequest
metadata:
  name: backup-postgres
spec:
  clusterRef:
    name: my-postgres
  type: Backup
  backup:
    backupName: my-postgres-backup-20250101
```

## Controller Workflow

### DatabaseCluster Controller

1. **Watch**: Monitors all DatabaseCluster and child cluster resources
2. **PreReconcileHook**: Provider-specific pre-processing (e.g., check restore status)
3. **Spec Transformation**:
   - Call all Applier methods to build child cluster spec
   - Metadata() → Engine() → Proxy() → Monitoring() → etc.
4. **Create/Update**: Apply child cluster resource
5. **Status Sync**: Map child cluster status to parent cluster status
6. **Cleanup**: On deletion, cleanup child resources via provider

### OpsRequest Controller

1. **Validate**: Check target cluster exists
2. **Execute**: Call appropriate operation handler method
3. **Monitor**: Poll operation status periodically
4. **Update Status**: Sync operation progress to OpsRequest status
5. **Action Log**: Record operation events
6. **TTL Cleanup**: Auto-delete completed operations after TTL

## Development

### Project Structure

```
dbaas-operator/
├── api/v1/                      # CRD definitions
│   ├── databasecluster_types.go
│   ├── databaseengine_types.go
│   ├── backupstorage_types.go
│   ├── monitoringconfig_types.go
│   └── opsrequest_types.go
├── controllers/                 # Controllers
│   ├── databasecluster_controller.go
│   └── opsrequest_controller.go
├── pkg/provider/               # Provider framework
│   ├── interface.go            # Provider & Applier interfaces
│   ├── factory.go              # Provider factory
│   └── cnpg/                   # CNPG provider implementation
│       ├── provider.go
│       ├── applier.go
│       └── operations.go
├── config/                     # Kubernetes configs
│   ├── crd/                    # CRD manifests
│   └── samples/                # Sample resources
├── Makefile
├── Dockerfile
└── main.go
```

### Adding a New Database Provider

1. Create provider directory: `pkg/provider/{engine}/`
2. Implement `Provider` interface
3. Implement `Applier` interface
4. Implement `OperationsHandler` interface
5. Register in `factory.go`

Example:
```go
// pkg/provider/mongodb/provider.go
type MongoDBProvider struct {
    client client.Client
    scheme *runtime.Scheme
}

func (p *MongoDBProvider) GetApplier(cluster *DatabaseCluster) (Applier, error) {
    return NewMongoDBApplier(cluster), nil
}
// ... implement other methods
```

### Building and Testing

```bash
# Run tests
make test

# Build binary
make build

# Run locally (requires kubeconfig)
make run

# Build Docker image
make docker-build IMG=myregistry/dbaas-operator:v1.0.0

# Push Docker image
make docker-push IMG=myregistry/dbaas-operator:v1.0.0
```

## Responsibilities

### DBaaS Operator Responsibilities

- Define and validate CRDs
- Implement core controllers (Cluster, OpsRequest, Engine, etc.)
- Provide Provider and Applier interfaces
- Manage RBAC and permissions
- Handle parent cluster lifecycle
- Common utilities and helpers

### Provider Responsibilities

- Implement Provider interface
- Implement Applier interface for spec transformation
- Implement OperationsHandler for day-2 operations
- Map child cluster status to parent cluster status
- Handle provider-specific cleanup logic
- Support custom configuration parameters

## Roadmap

- [x] Core architecture and interfaces
- [x] CNPG provider for PostgreSQL
- [ ] Percona provider for MongoDB
- [ ] Percona provider for MySQL
- [ ] Strimzi provider for Kafka
- [ ] Validation webhooks
- [ ] Conversion webhooks for API versioning
- [ ] Advanced monitoring with Grafana dashboards
- [ ] Multi-cluster support
- [ ] GitOps integration
- [ ] CLI tool for database management

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

Apache License 2.0

## Support

For issues and questions:
- GitHub Issues: https://github.com/huynt0812/dbaas-operator/issues
- Documentation: https://github.com/huynt0812/dbaas-operator/docs
