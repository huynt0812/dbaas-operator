package cnpg

import (
	"context"
	"fmt"
	"time"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	dbaasv1 "github.com/huynt0812/dbaas-operator/api/v1"
	"github.com/huynt0812/dbaas-operator/pkg/provider"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CNPGProvider implements the Provider interface for CloudNativePG
type CNPGProvider struct {
	client client.Client
	scheme *runtime.Scheme
}

// NewProvider creates a new CNPG provider instance
func NewProvider(c client.Client, scheme *runtime.Scheme) provider.Provider {
	return &CNPGProvider{
		client: c,
		scheme: scheme,
	}
}

// GetApplier returns an Applier instance for building the CNPG Cluster spec
func (p *CNPGProvider) GetApplier(cluster *dbaasv1.DatabaseCluster) (provider.Applier, error) {
	return NewApplier(cluster, p.client, p.scheme), nil
}

// Status maps the CNPG Cluster status to DatabaseCluster status
func (p *CNPGProvider) Status(ctx context.Context, cluster *dbaasv1.DatabaseCluster) (*dbaasv1.DatabaseClusterStatus, error) {
	// Get the CNPG Cluster
	cnpgCluster := &cnpgv1.Cluster{}
	err := p.client.Get(ctx, types.NamespacedName{
		Name:      cluster.Name,
		Namespace: cluster.Namespace,
	}, cnpgCluster)
	if err != nil {
		return nil, err
	}

	status := &dbaasv1.DatabaseClusterStatus{
		ObservedGeneration: cluster.Generation,
	}

	// Map phase
	if cnpgCluster.Status.Phase == "" {
		status.Phase = dbaasv1.ClusterPhaseInitializing
	} else {
		status.Phase = mapCNPGPhaseToDBaaSPhase(cnpgCluster.Status.Phase)
	}

	// Map database status
	status.Database = &dbaasv1.DatabaseStatus{
		Ready:          cnpgCluster.Status.Phase == "Cluster in healthy state",
		Instances:      cnpgCluster.Status.Instances,
		ReadyInstances: cnpgCluster.Status.ReadyInstances,
		PrimaryInstance: cnpgCluster.Status.CurrentPrimary,
		Roles:          make(map[string]string),
		Endpoints: &dbaasv1.DatabaseEndpoints{
			Primary: fmt.Sprintf("%s-rw.%s.svc.cluster.local", cluster.Name, cluster.Namespace),
			Replica: fmt.Sprintf("%s-ro.%s.svc.cluster.local", cluster.Name, cluster.Namespace),
		},
	}

	// Map instance roles
	for _, instance := range cnpgCluster.Status.InstancesStatus {
		if instance.IsPrimary {
			status.Database.Roles[instance.Pod] = "primary"
		} else {
			status.Database.Roles[instance.Pod] = "replica"
		}
	}

	// Map backup status
	if cluster.Spec.Backup != nil && cluster.Spec.Backup.Enabled {
		status.Backup = &dbaasv1.BackupStatus{}
		if cnpgCluster.Status.FirstRecoverabilityPoint != "" {
			// Parse time from CNPG format
			// This is simplified - actual implementation would parse the time properly
			status.Backup.LastBackupTime = &cluster.CreationTimestamp
		}
	}

	// Map monitoring status
	if cluster.Spec.Monitoring != nil && cluster.Spec.Monitoring.Enabled {
		status.Monitoring = &dbaasv1.MonitoringStatus{
			Enabled: true,
			Ready:   true, // Simplified - would check actual monitoring status
		}
	}

	// Map proxy status
	if cluster.Spec.Proxy != nil && cluster.Spec.Proxy.Enabled {
		status.Proxy = &dbaasv1.ProxyStatus{
			Ready:         true, // Simplified
			Replicas:      cluster.Spec.Proxy.Replicas,
			ReadyReplicas: cluster.Spec.Proxy.Replicas,
		}
	}

	// Set message
	if len(cnpgCluster.Status.Conditions) > 0 {
		status.Message = cnpgCluster.Status.Conditions[0].Message
	}

	return status, nil
}

// Cleanup performs cleanup operations when deleting a cluster
func (p *CNPGProvider) Cleanup(ctx context.Context, cluster *dbaasv1.DatabaseCluster) error {
	// Delete the CNPG Cluster
	cnpgCluster := &cnpgv1.Cluster{}
	err := p.client.Get(ctx, types.NamespacedName{
		Name:      cluster.Name,
		Namespace: cluster.Namespace,
	}, cnpgCluster)
	if err != nil {
		return client.IgnoreNotFound(err)
	}

	return p.client.Delete(ctx, cnpgCluster)
}

// PreReconcileHook is called before reconciling the cluster
func (p *CNPGProvider) PreReconcileHook(ctx context.Context, cluster *dbaasv1.DatabaseCluster) (requeueAfter int, err error) {
	// Check if we're in the middle of a restore operation
	if cluster.Spec.DataSource != nil && cluster.Spec.DataSource.BackupSource != nil {
		cnpgCluster := &cnpgv1.Cluster{}
		err := p.client.Get(ctx, types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		}, cnpgCluster)
		if err != nil {
			return 0, err
		}

		// If cluster is still initializing, requeue after 5 seconds
		if cnpgCluster.Status.Phase != "Cluster in healthy state" {
			return 5, nil
		}
	}

	return 0, nil
}

// Operations returns the operations handler for this provider
func (p *CNPGProvider) Operations() provider.OperationsHandler {
	return NewOperationsHandler(p.client, p.scheme)
}

// Helper function to map CNPG phase to DBaaS phase
func mapCNPGPhaseToDBaaSPhase(cnpgPhase string) dbaasv1.ClusterPhase {
	switch cnpgPhase {
	case "Cluster in healthy state":
		return dbaasv1.ClusterPhaseReady
	case "Creating primary instance":
		return dbaasv1.ClusterPhaseInitializing
	case "Upgrading cluster":
		return dbaasv1.ClusterPhaseUpdating
	default:
		return dbaasv1.ClusterPhaseInitializing
	}
}
