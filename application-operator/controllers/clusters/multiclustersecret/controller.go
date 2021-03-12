// Copyright (c) 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package multiclustersecret

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/verrazzano/verrazzano/application-operator/controllers/clusters"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clustersv1alpha1 "github.com/verrazzano/verrazzano/application-operator/apis/clusters/v1alpha1"
)

// Reconciler reconciles a MultiClusterSecret object
type Reconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	AgentChannel chan clusters.StatusUpdateMessage
}

// Reconcile reconciles a MultiClusterSecret resource. It fetches the embedded Secret, mutates it
// based on the MultiClusterSecret, and updates the status of the MultiClusterSecret to reflect the
// success or failure of the changes to the embedded Secret
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("multiclustersecret", req.NamespacedName)
	var mcSecret clustersv1alpha1.MultiClusterSecret
	result := reconcile.Result{}
	ctx := context.Background()
	err := r.fetchMultiClusterSecret(ctx, req.NamespacedName, &mcSecret)
	if err != nil {
		return result, clusters.IgnoreNotFoundWithLog("MultiClusterSecret", err, logger)
	}

	oldState := clusters.SetEffectiveStateIfChanged(mcSecret.Spec.Placement, &mcSecret.Status)
	if !clusters.IsPlacedInThisCluster(ctx, r, mcSecret.Spec.Placement) {
		if oldState != mcSecret.Status.State {
			// This must be done whether the resource is placed in this cluster or not, because we
			// could be in an admin cluster and receive cluster level statuses from managed clusters,
			// which can change our effective state
			err = r.Status().Update(ctx, &mcSecret)
		}
		return ctrl.Result{}, err
	}

	logger.Info("MultiClusterSecret create or update with underlying secret",
		"secret", mcSecret.Spec.Template.Metadata.Name,
		"placement", mcSecret.Spec.Placement.Clusters[0].Name)
	opResult, err := r.createOrUpdateSecret(ctx, mcSecret)

	return r.updateStatus(ctx, &mcSecret, opResult, err)
}

func (r *Reconciler) updateStatus(ctx context.Context, mcSecret *clustersv1alpha1.MultiClusterSecret, opResult controllerutil.OperationResult, err error) (ctrl.Result, error) {
	clusterName := clusters.GetClusterName(ctx, r.Client)
	newCondition := clusters.GetConditionFromResult(err, opResult, "Secret")
	return clusters.UpdateStatus(mcSecret, &mcSecret.Status, mcSecret.Spec.Placement, newCondition, clusterName,
		r.AgentChannel, func() error { return r.Status().Update(ctx, mcSecret) })
}

// SetupWithManager registers our controller with the manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.MultiClusterSecret{}).
		Complete(r)
}

func (r *Reconciler) fetchMultiClusterSecret(ctx context.Context, name types.NamespacedName, mcSecretRef *clustersv1alpha1.MultiClusterSecret) error {
	return r.Get(ctx, name, mcSecretRef)
}

func (r *Reconciler) createOrUpdateSecret(ctx context.Context, mcSecret clustersv1alpha1.MultiClusterSecret) (controllerutil.OperationResult, error) {
	var secret corev1.Secret
	secret.Namespace = mcSecret.Namespace
	secret.Name = mcSecret.Name

	return controllerutil.CreateOrUpdate(ctx, r.Client, &secret, func() error {
		r.mutateSecret(mcSecret, &secret)
		// This SetControllerReference call will trigger garbage collection i.e. the secret
		// will automatically get deleted when the mcSecret is deleted
		return controllerutil.SetControllerReference(&mcSecret, &secret, r.Scheme)
	})

}

// mutateSecret mutates the corev1.Secret to reflect the contents of the parent MultiClusterSecret
func (r *Reconciler) mutateSecret(mcSecret clustersv1alpha1.MultiClusterSecret, secret *corev1.Secret) {
	secret.Type = mcSecret.Spec.Template.Type
	secret.Data = mcSecret.Spec.Template.Data
	secret.StringData = mcSecret.Spec.Template.StringData
	secret.Labels = mcSecret.Spec.Template.Metadata.Labels
	secret.Annotations = mcSecret.Spec.Template.Metadata.Annotations
}
