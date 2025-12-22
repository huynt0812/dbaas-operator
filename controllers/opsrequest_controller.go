package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dbaasv1 "github.com/huynt0812/dbaas-operator/api/v1"
	"github.com/huynt0812/dbaas-operator/pkg/provider"
)

// OpsRequestReconciler reconciles an OpsRequest object
type OpsRequestReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	ProviderFactory provider.ProviderFactory
}

// +kubebuilder:rbac:groups=dbaas.io,resources=opsrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dbaas.io,resources=opsrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dbaas.io,resources=opsrequests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop
func (r *OpsRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the OpsRequest instance
	ops := &dbaasv1.OpsRequest{}
	if err := r.Get(ctx, req.NamespacedName, ops); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch OpsRequest")
		return ctrl.Result{}, err
	}

	// Skip if already completed or failed
	if ops.Status.Phase == dbaasv1.OpsRequestPhaseSucceeded || ops.Status.Phase == dbaasv1.OpsRequestPhaseFailed {
		// Check TTL
		if ops.Spec.TTLSecondsAfterFinished != nil && ops.Status.CompletionTime != nil {
			ttl := time.Duration(*ops.Spec.TTLSecondsAfterFinished) * time.Second
			if time.Since(ops.Status.CompletionTime.Time) > ttl {
				log.Info("Deleting OpsRequest due to TTL expiration")
				return ctrl.Result{}, r.Delete(ctx, ops)
			}
		}
		return ctrl.Result{}, nil
	}

	// Get the target DatabaseCluster
	cluster := &dbaasv1.DatabaseCluster{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      ops.Spec.ClusterRef.Name,
		Namespace: ops.Namespace,
	}, cluster); err != nil {
		log.Error(err, "unable to fetch target DatabaseCluster")
		return r.updateStatusFailed(ctx, ops, fmt.Sprintf("target cluster not found: %v", err))
	}

	// Get the provider for this engine type
	prov, err := r.ProviderFactory.GetProvider(cluster.Spec.Engine.Type, r.Client, r.Scheme)
	if err != nil {
		log.Error(err, "unable to get provider", "engineType", cluster.Spec.Engine.Type)
		return r.updateStatusFailed(ctx, ops, fmt.Sprintf("unable to get provider: %v", err))
	}

	// Get operations handler
	opsHandler := prov.Operations()

	// Set status to running if pending
	if ops.Status.Phase == "" || ops.Status.Phase == dbaasv1.OpsRequestPhasePending {
		ops.Status.Phase = dbaasv1.OpsRequestPhaseRunning
		ops.Status.StartTime = &metav1.Time{Time: time.Now()}
		if err := r.Status().Update(ctx, ops); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Execute the operation based on type
	if err := r.executeOperation(ctx, opsHandler, cluster, ops); err != nil {
		log.Error(err, "failed to execute operation")
		return r.updateStatusFailed(ctx, ops, err.Error())
	}

	// Get operation status
	status, err := opsHandler.GetStatus(ctx, cluster, ops)
	if err != nil {
		log.Error(err, "failed to get operation status")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Update OpsRequest status
	ops.Status = *status
	if err := r.Status().Update(ctx, ops); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue if still running
	if status.Phase == dbaasv1.OpsRequestPhaseRunning {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

// executeOperation executes the specified operation
func (r *OpsRequestReconciler) executeOperation(ctx context.Context, opsHandler provider.OperationsHandler, cluster *dbaasv1.DatabaseCluster, ops *dbaasv1.OpsRequest) error {
	switch ops.Spec.Type {
	case dbaasv1.OpsRequestTypeStart:
		return opsHandler.Start(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeStop:
		return opsHandler.Stop(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeRestart:
		return opsHandler.Restart(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeSwitchover:
		return opsHandler.Switchover(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeHorizontalScaling:
		return opsHandler.HorizontalScaling(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeVerticalScaling:
		return opsHandler.VerticalScaling(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeVolumeExpansion:
		return opsHandler.VolumeExpansion(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeReconfiguring:
		return opsHandler.Reconfigure(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeUpgrade:
		return opsHandler.Upgrade(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeBackup:
		return opsHandler.Backup(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeRestore:
		return opsHandler.Restore(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeExpose:
		return opsHandler.Expose(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeRebuildInstance:
		return opsHandler.RebuildInstance(ctx, cluster, ops)
	case dbaasv1.OpsRequestTypeCustom:
		return opsHandler.Custom(ctx, cluster, ops)
	default:
		return fmt.Errorf("unsupported operation type: %s", ops.Spec.Type)
	}
}

// updateStatusFailed updates the OpsRequest status to failed
func (r *OpsRequestReconciler) updateStatusFailed(ctx context.Context, ops *dbaasv1.OpsRequest, message string) (ctrl.Result, error) {
	ops.Status.Phase = dbaasv1.OpsRequestPhaseFailed
	ops.Status.Message = message
	ops.Status.CompletionTime = &metav1.Time{Time: time.Now()}

	if err := r.Status().Update(ctx, ops); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpsRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbaasv1.OpsRequest{}).
		Complete(r)
}
