package interfaces

import (
	"context"

	dbaasv1 "github.com/huynt0812/dbaas-operator/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provider defines the interface that all database providers must implement
type Provider interface {
	// GetApplier returns an Applier instance for building the child cluster spec
	GetApplier(cluster *dbaasv1.DatabaseCluster) (Applier, error)

	// Status maps the child cluster status to parent cluster status
	Status(ctx context.Context, cluster *dbaasv1.DatabaseCluster) (*dbaasv1.DatabaseClusterStatus, error)

	// Cleanup performs cleanup operations when deleting a cluster
	Cleanup(ctx context.Context, cluster *dbaasv1.DatabaseCluster) error

	// PreReconcileHook is called before reconciling the cluster
	// Returns requeue duration (0 means no requeue) and error
	PreReconcileHook(ctx context.Context, cluster *dbaasv1.DatabaseCluster) (requeueAfter int, err error)

	// Operations returns the operations handler for this provider
	Operations() OperationsHandler
}

// Applier defines the interface for building child cluster specifications
type Applier interface {
	// Metadata applies metadata (labels, annotations, etc.)
	Metadata() (map[string]string, map[string]string, error)

	// Engine applies engine-specific configuration
	Engine() (runtime.Object, error)

	// Proxy applies proxy configuration
	Proxy() (runtime.Object, error)

	// Monitoring applies monitoring configuration
	Monitoring() (runtime.Object, error)

	// PodSchedulingPolicy applies pod scheduling constraints
	PodSchedulingPolicy() error

	// Backup applies backup configuration
	Backup() error

	// DataSource applies data source configuration (restore, clone)
	DataSource() error

	// DataImport applies data import configuration
	DataImport() error

	// GetResult returns the final child cluster object
	GetResult() runtime.Object
}

// OperationsHandler defines the interface for day-2 operations
type OperationsHandler interface {
	// Start starts the database cluster
	Start(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// Stop stops the database cluster
	Stop(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// Restart restarts the database cluster
	Restart(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// Switchover performs a switchover operation
	Switchover(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// HorizontalScaling performs horizontal scaling
	HorizontalScaling(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// VerticalScaling performs vertical scaling
	VerticalScaling(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// VolumeExpansion performs volume expansion
	VolumeExpansion(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// Reconfigure performs reconfiguration
	Reconfigure(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// Upgrade performs version upgrade
	Upgrade(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// Backup performs backup operation
	Backup(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// Restore performs restore operation
	Restore(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// Expose exposes the database service
	Expose(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// RebuildInstance rebuilds a specific instance
	RebuildInstance(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// Custom performs custom operations
	Custom(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error

	// GetStatus returns the current status of an operation
	GetStatus(ctx context.Context, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) (*dbaasv1.OpsRequestStatus, error)
}

// ProviderFactory creates provider instances
type ProviderFactory interface {
	// GetProvider returns a provider instance for the given engine type
	GetProvider(engineType string, client client.Client, scheme *runtime.Scheme) (Provider, error)
}
