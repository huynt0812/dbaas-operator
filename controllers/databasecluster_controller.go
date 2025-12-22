package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dbaasv1 "github.com/huynt0812/dbaas-operator/api/v1"
	"github.com/huynt0812/dbaas-operator/pkg/provider"
)

// DatabaseClusterReconciler reconciles a DatabaseCluster object
type DatabaseClusterReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	ProviderFactory provider.ProviderFactory
}

// +kubebuilder:rbac:groups=dbaas.io,resources=databaseclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dbaas.io,resources=databaseclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dbaas.io,resources=databaseclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=postgresql.cnpg.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgresql.cnpg.io,resources=backups,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *DatabaseClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the DatabaseCluster instance
	cluster := &dbaasv1.DatabaseCluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch DatabaseCluster")
		return ctrl.Result{}, err
	}

	// Get the provider for this engine type
	prov, err := r.ProviderFactory.GetProvider(cluster.Spec.Engine.Type, r.Client, r.Scheme)
	if err != nil {
		log.Error(err, "unable to get provider", "engineType", cluster.Spec.Engine.Type)
		return ctrl.Result{}, err
	}

	// Call PreReconcileHook
	requeueAfter, err := prov.PreReconcileHook(ctx, cluster)
	if err != nil {
		log.Error(err, "PreReconcileHook failed")
		return ctrl.Result{}, err
	}
	if requeueAfter > 0 {
		log.Info("PreReconcileHook requested requeue", "after", requeueAfter)
		return ctrl.Result{RequeueAfter: time.Duration(requeueAfter) * time.Second}, nil
	}

	// Handle deletion
	if !cluster.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, cluster, prov)
	}

	// Add finalizer if not present
	if !containsString(cluster.Finalizers, "dbaas.io/finalizer") {
		cluster.Finalizers = append(cluster.Finalizers, "dbaas.io/finalizer")
		if err := r.Update(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Get the applier to build child cluster spec
	applier, err := prov.GetApplier(cluster)
	if err != nil {
		log.Error(err, "unable to get applier")
		return ctrl.Result{}, err
	}

	// Apply all transformations
	if err := r.applyTransformations(applier); err != nil {
		log.Error(err, "failed to apply transformations")
		return ctrl.Result{}, err
	}

	// Get the resulting child cluster object
	childCluster := applier.GetResult()

	// Create or update the child cluster
	if err := r.createOrUpdateChildCluster(ctx, childCluster); err != nil {
		log.Error(err, "failed to create or update child cluster")
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, cluster, prov); err != nil {
		log.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// applyTransformations applies all applier transformations
func (r *DatabaseClusterReconciler) applyTransformations(applier provider.Applier) error {
	// Apply metadata
	if _, _, err := applier.Metadata(); err != nil {
		return fmt.Errorf("failed to apply metadata: %w", err)
	}

	// Apply engine configuration
	if _, err := applier.Engine(); err != nil {
		return fmt.Errorf("failed to apply engine: %w", err)
	}

	// Apply proxy
	if _, err := applier.Proxy(); err != nil {
		return fmt.Errorf("failed to apply proxy: %w", err)
	}

	// Apply monitoring
	if _, err := applier.Monitoring(); err != nil {
		return fmt.Errorf("failed to apply monitoring: %w", err)
	}

	// Apply pod scheduling policy
	if err := applier.PodSchedulingPolicy(); err != nil {
		return fmt.Errorf("failed to apply pod scheduling policy: %w", err)
	}

	// Apply backup
	if err := applier.Backup(); err != nil {
		return fmt.Errorf("failed to apply backup: %w", err)
	}

	// Apply data source
	if err := applier.DataSource(); err != nil {
		return fmt.Errorf("failed to apply data source: %w", err)
	}

	// Apply data import
	if err := applier.DataImport(); err != nil {
		return fmt.Errorf("failed to apply data import: %w", err)
	}

	return nil
}

// createOrUpdateChildCluster creates or updates the child cluster
func (r *DatabaseClusterReconciler) createOrUpdateChildCluster(ctx context.Context, childCluster runtime.Object) error {
	obj := childCluster.(client.Object)

	// Try to create
	err := r.Create(ctx, obj)
	if err == nil {
		return nil
	}

	// If already exists, update
	if errors.IsAlreadyExists(err) {
		return r.Update(ctx, obj)
	}

	return err
}

// updateStatus updates the DatabaseCluster status
func (r *DatabaseClusterReconciler) updateStatus(ctx context.Context, cluster *dbaasv1.DatabaseCluster, prov provider.Provider) error {
	status, err := prov.Status(ctx, cluster)
	if err != nil {
		return err
	}

	cluster.Status = *status
	return r.Status().Update(ctx, cluster)
}

// handleDeletion handles cluster deletion
func (r *DatabaseClusterReconciler) handleDeletion(ctx context.Context, cluster *dbaasv1.DatabaseCluster, prov provider.Provider) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if containsString(cluster.Finalizers, "dbaas.io/finalizer") {
		// Perform cleanup
		if err := prov.Cleanup(ctx, cluster); err != nil {
			log.Error(err, "failed to cleanup resources")
			return ctrl.Result{}, err
		}

		// Remove finalizer
		cluster.Finalizers = removeString(cluster.Finalizers, "dbaas.io/finalizer")
		if err := r.Update(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbaasv1.DatabaseCluster{}).
		Complete(r)
}

// Helper functions
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	result := []string{}
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
