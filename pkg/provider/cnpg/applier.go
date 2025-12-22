package cnpg

import (
	"fmt"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	dbaasv1 "github.com/huynt0812/dbaas-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// CNPGApplier implements the Applier interface for CloudNativePG
type CNPGApplier struct {
	cluster      *dbaasv1.DatabaseCluster
	cnpgCluster  *cnpgv1.Cluster
	client       client.Client
	scheme       *runtime.Scheme
}

// NewApplier creates a new CNPG applier instance
func NewApplier(cluster *dbaasv1.DatabaseCluster, c client.Client, scheme *runtime.Scheme) *CNPGApplier {
	cnpgCluster := &cnpgv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		Spec: cnpgv1.ClusterSpec{
			Instances: int(cluster.Spec.ClusterSize),
		},
	}

	return &CNPGApplier{
		cluster:     cluster,
		cnpgCluster: cnpgCluster,
		client:      c,
		scheme:      scheme,
	}
}

// Metadata applies metadata (labels, annotations, etc.)
func (a *CNPGApplier) Metadata() (map[string]string, map[string]string, error) {
	labels := make(map[string]string)
	annotations := make(map[string]string)

	// Copy labels from parent cluster
	for k, v := range a.cluster.Labels {
		labels[k] = v
	}

	// Add DBaaS-specific labels
	labels["dbaas.io/cluster"] = a.cluster.Name
	labels["dbaas.io/engine"] = a.cluster.Spec.Engine.Type

	// Copy annotations from parent cluster
	for k, v := range a.cluster.Annotations {
		annotations[k] = v
	}

	// Apply to CNPG cluster
	a.cnpgCluster.Labels = labels
	a.cnpgCluster.Annotations = annotations

	return labels, annotations, nil
}

// Engine applies engine-specific configuration
func (a *CNPGApplier) Engine() (runtime.Object, error) {
	// Set PostgreSQL version
	a.cnpgCluster.Spec.ImageName = fmt.Sprintf("ghcr.io/cloudnative-pg/postgresql:%s", a.cluster.Spec.Engine.Version)

	// Set storage configuration
	storageSize := a.cluster.Spec.Storage.Size.String()
	a.cnpgCluster.Spec.StorageConfiguration = cnpgv1.StorageConfiguration{
		Size: storageSize,
	}

	if a.cluster.Spec.Storage.StorageClassName != nil {
		a.cnpgCluster.Spec.StorageConfiguration.StorageClass = a.cluster.Spec.Storage.StorageClassName
	}

	// Set resource requirements
	if a.cluster.Spec.Resources.Requests != nil || a.cluster.Spec.Resources.Limits != nil {
		a.cnpgCluster.Spec.Resources = corev1.ResourceRequirements{
			Requests: a.cluster.Spec.Resources.Requests,
			Limits:   a.cluster.Spec.Resources.Limits,
		}
	} else {
		// Set default resources
		a.cnpgCluster.Spec.Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
		}
	}

	// Apply custom configuration from spec.config
	if len(a.cluster.Spec.Config) > 0 {
		postgresqlConfig := make(map[string]string)
		for _, cfg := range a.cluster.Spec.Config {
			postgresqlConfig[cfg.Name] = cfg.Value
		}
		a.cnpgCluster.Spec.PostgresConfiguration = cnpgv1.PostgresConfiguration{
			Parameters: postgresqlConfig,
		}
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(a.cluster, a.cnpgCluster, a.scheme); err != nil {
		return nil, err
	}

	return a.cnpgCluster, nil
}

// Proxy applies proxy configuration
func (a *CNPGApplier) Proxy() (runtime.Object, error) {
	if a.cluster.Spec.Proxy != nil && a.cluster.Spec.Proxy.Enabled {
		// CNPG uses PgBouncer pooler
		if a.cluster.Spec.Proxy.Type == "pgbouncer" || a.cluster.Spec.Proxy.Type == "" {
			a.cnpgCluster.Spec.Managed = &cnpgv1.ManagedConfiguration{
				Roles: []cnpgv1.RoleConfiguration{
					{
						Name: "app",
					},
				},
			}

			// Note: Full PgBouncer integration would require additional configuration
			// This is a simplified example
			// pooler instances would be configured here
		}
	}
	return nil, nil
}

// Monitoring applies monitoring configuration
func (a *CNPGApplier) Monitoring() (runtime.Object, error) {
	if a.cluster.Spec.Monitoring != nil && a.cluster.Spec.Monitoring.Enabled {
		// Enable monitoring in CNPG
		a.cnpgCluster.Spec.Monitoring = &cnpgv1.MonitoringConfiguration{
			EnablePodMonitor: true,
		}

		// If PMM is configured, set up PMM sidecar
		if a.cluster.Spec.Monitoring.PMM != nil {
			// PMM would be configured as a sidecar container
			// This is simplified - actual implementation would add PMM container
		}
	}
	return nil, nil
}

// PodSchedulingPolicy applies pod scheduling constraints
func (a *CNPGApplier) PodSchedulingPolicy() error {
	if a.cluster.Spec.PodSchedulingPolicy != nil {
		policy := a.cluster.Spec.PodSchedulingPolicy

		// Apply affinity
		if policy.Affinity != nil {
			a.cnpgCluster.Spec.Affinity = cnpgv1.AffinityConfiguration{
				PodAntiAffinityType: "preferred",
			}
			// Note: Full affinity configuration would be more complex
		}

		// Apply node selector
		if len(policy.NodeSelector) > 0 {
			a.cnpgCluster.Spec.NodeMaintenanceWindow = &cnpgv1.NodeMaintenanceWindow{
				InProgress: false,
			}
			// Note: CNPG handles node selector differently
		}

		// Apply tolerations
		if len(policy.Tolerations) > 0 {
			// CNPG doesn't have direct toleration support in the current API version
			// Would need to use pod template modifications
		}
	}
	return nil
}

// Backup applies backup configuration
func (a *CNPGApplier) Backup() error {
	if a.cluster.Spec.Backup != nil && a.cluster.Spec.Backup.Enabled {
		backup := &cnpgv1.BackupConfiguration{
			RetentionPolicy: a.cluster.Spec.Backup.RetentionPolicy,
		}

		if a.cluster.Spec.Backup.BackupStorageRef != nil {
			// Would fetch BackupStorage CR and configure barman object store
			// This is simplified
			backup.BarmanObjectStore = &cnpgv1.BarmanObjectStoreConfiguration{
				DestinationPath: "s3://backups/",
				ServerName:      a.cluster.Name,
			}
		}

		a.cnpgCluster.Spec.Backup = backup
	}
	return nil
}

// DataSource applies data source configuration (restore, clone)
func (a *CNPGApplier) DataSource() error {
	if a.cluster.Spec.DataSource != nil {
		if a.cluster.Spec.DataSource.BackupSource != nil {
			// Configure bootstrap from backup
			// Note: CNPG v1.23 has different BackupSource structure
			// This is a simplified placeholder - full implementation would use proper CNPG API
			a.cnpgCluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
				Recovery: &cnpgv1.BootstrapRecovery{},
			}
		} else if a.cluster.Spec.DataSource.CloneSource != nil {
			// Configure bootstrap from clone
			// Note: Simplified - actual implementation would be more complex
			a.cnpgCluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
				PgBaseBackup: &cnpgv1.BootstrapPgBaseBackup{
					Source: a.cluster.Spec.DataSource.CloneSource.ClusterName,
				},
			}
		}
	} else {
		// Default bootstrap - initialize a new cluster
		a.cnpgCluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
			InitDB: &cnpgv1.BootstrapInitDB{
				Database: "app",
				Owner:    "app",
			},
		}
	}
	return nil
}

// DataImport applies data import configuration
func (a *CNPGApplier) DataImport() error {
	// Data import would be handled through custom logic
	// Not directly supported in this simplified implementation
	return nil
}

// GetResult returns the final CNPG Cluster object
func (a *CNPGApplier) GetResult() runtime.Object {
	return a.cnpgCluster
}
