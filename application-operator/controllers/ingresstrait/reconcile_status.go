// Copyright (c) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package ingresstrait

import (
	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileStatus is used to collect the results of creating or updating child resources during reconciliation.
type reconcileStatus struct {
	children []cpv1alpha1.TypedReference
	results  []controllerutil.OperationResult
	errors   []error
}

// containsErrors scans the status to determine if any errors were recorded.
func (s *reconcileStatus) containsErrors() bool {
	for _, err := range s.errors {
		if err != nil {
			return true
		}
	}
	return false
}

// createConditionedStatus creates conditioned status for use in object status.
// If no errors are found in the reconcile status a success condition is returned.
// Otherwise reconcile errors statuses are returned for each error.
func (s *reconcileStatus) createConditionedStatus() cpv1alpha1.ConditionedStatus {
	var conditions []cpv1alpha1.Condition
	for _, err := range s.errors {
		if err != nil {
			conditions = append(conditions, cpv1alpha1.ReconcileError(err))
		}
	}
	if len(conditions) == 0 {
		conditions = append(conditions, cpv1alpha1.ReconcileSuccess())
	}
	return cpv1alpha1.ConditionedStatus{Conditions: conditions}
}

// createTypedReferences creates the typed reference slice for use in object status.
func (s *reconcileStatus) createTypedReferences() []cpv1alpha1.TypedReference {
	// Copies the slice.
	return append([]cpv1alpha1.TypedReference{}, s.children...)
}
