// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package v1beta1

import (
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"testing"
)

func TestConvertTo(t *testing.T) {
	var tests = []converisonTestCase{
		{
			"converts to v1alpha1 in the basic case",
			testCaseBasic,
			false,
		},
		{
			"converts status to v1alpha1",
			testCaseStatus,
			false,
		},
		{
			"converts rancher keycloak auth",
			testCaseRancherKeycloak,
			false,
		},
		{
			"converts all comps to v1alpha1",
			testCaseToAllComps,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// load the expected v1beta1 CR for conversion
			v1Beta1CR, err := loadV1Beta1(tt.testCase)
			assert.NoError(t, err)

			// compute the actual v1beta1 CR from the v1alpha1 CR
			v1alpha1Actual := &v1alpha1.Verrazzano{}
			err = v1Beta1CR.ConvertTo(v1alpha1Actual)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// load the expected v1alpha1 CR
				v1alpha1Expected, err := loadV1Alpha1CR(tt.testCase)
				assert.NoError(t, err)
				// expected and actual v1beta1 CRs must be equal
				assert.EqualValues(t, v1alpha1Expected.ObjectMeta, v1alpha1Actual.ObjectMeta)
				assert.EqualValues(t, v1alpha1Expected.Spec, v1alpha1Actual.Spec)
				assert.EqualValues(t, v1alpha1Expected.Status, v1alpha1Actual.Status)
			}
		})
	}
}
