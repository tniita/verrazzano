// Copyright (c) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package ingresstrait

import (
	"fmt"
	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	asserts "github.com/stretchr/testify/assert"
	"testing"
)

// TestReconcileStatusContainsErrors tests positive and negative use cases using containsErrors
func TestReconcileStatusContainsErrors(t *testing.T) {
	assert := asserts.New(t)
	var status reconcileStatus

	// GIVEN status with all default fields
	// WHEN the status is checked for errors
	// THEN verify that errors are not detected
	status = reconcileStatus{}
	assert.False(status.containsErrors())

	// GIVEN status that contains an error
	// WHEN the status is checked for errors
	// THEN verify that errors are detected
	status = reconcileStatus{errors: []error{fmt.Errorf("test-error")}}
	assert.True(status.containsErrors())

	// GIVEN status that contains no errors
	// WHEN the status is checked for errors
	// THEN verify that errors are not detected
	status = reconcileStatus{errors: []error{nil}}
	assert.False(status.containsErrors())

	// GIVEN status that contains both errors and nils
	// WHEN the status is checked for errors
	// THEN verify that errors are detected
	status = reconcileStatus{errors: []error{nil, fmt.Errorf("test-error"), nil}}
	assert.True(status.containsErrors())
}

func TestReconcileStatusCreateConditionedStatus(t *testing.T) {
	assert := asserts.New(t)
	var recStatus reconcileStatus
	var condStatus cpv1alpha1.ConditionedStatus

	// GIVEN a default reconcile status
	// WHEN the conditioned status is created from the reconcile status
	// THEN verify that the conditioned status indicates success
	recStatus = reconcileStatus{}
	condStatus = recStatus.createConditionedStatus()
	assert.Len(condStatus.Conditions, 1)
	assert.Equal(cpv1alpha1.ReasonReconcileSuccess, condStatus.Conditions[0].Reason)

	// GIVEN a reconcile status with one error
	// WHEN the conditioned status is created from the reconcile status
	// THEN verify that the error is included in the conditioned status
	recStatus = reconcileStatus{errors: []error{fmt.Errorf("test-error")}}
	condStatus = recStatus.createConditionedStatus()
	assert.Len(condStatus.Conditions, 1)
	assert.Equal(cpv1alpha1.ReasonReconcileError, condStatus.Conditions[0].Reason)
	assert.Equal("test-error", condStatus.Conditions[0].Message)

	// GIVEN a reconcile status with one success
	// WHEN the conditioned status is created from the reconcile status
	// THEN verify that the conditioned status indicates success
	recStatus = reconcileStatus{errors: []error{nil}}
	condStatus = recStatus.createConditionedStatus()
	assert.Len(condStatus.Conditions, 1)
	assert.Equal(cpv1alpha1.ReasonReconcileSuccess, condStatus.Conditions[0].Reason)

	// GIVEN a reconcile status with both success and error
	// WHEN the conditioned status is created from the reconcile status
	// THEN verify that the conditioned status indicates failure
	recStatus = reconcileStatus{errors: []error{fmt.Errorf("test-error"), nil}}
	condStatus = recStatus.createConditionedStatus()
	assert.Len(condStatus.Conditions, 1)
	assert.Equal(cpv1alpha1.ReasonReconcileError, condStatus.Conditions[0].Reason)
	assert.Equal("test-error", condStatus.Conditions[0].Message)
}

func TestReconcileStatusCreateTypedReferences(t *testing.T) {
	assert := asserts.New(t)
	var status reconcileStatus
	var refs []cpv1alpha1.TypedReference

	// GIVEN a default reconcile status
	// WHEN a types reference slice is retrieved
	// THEN verify the slice is empty.
	status = reconcileStatus{}
	refs = status.createTypedReferences()
	assert.Len(refs, 0)

	// GIVEN a reconcile status with a child reference
	// WHEN a types reference slice is retrieved
	// THEN verify the retrieved slice matches the reconcile status child slice.
	status = reconcileStatus{children: []cpv1alpha1.TypedReference{{
		APIVersion: "test-apiver",
		Kind:       "test-kind",
		Name:       "test-name"}}}
	refs = status.createTypedReferences()
	assert.Equal(status.children, refs)
}
