// Copyright (c) 2021, 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"fmt"
	"os"
	"strings"
	"testing"

	certapiv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano/pkg/bom"
	"github.com/verrazzano/verrazzano/pkg/k8sutil"
	k8sutilfake "github.com/verrazzano/verrazzano/pkg/k8sutil/fake"
	vzapi "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/common"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func getValue(kvs []bom.KeyValue, key string) (string, bool) {
	for _, kv := range kvs {
		if strings.EqualFold(key, kv.Key) {
			return kv.Value, true
		}
	}
	return "", false
}

// TestAppendRegistryOverrides verifies that registry overrides are added as appropriate
// GIVEN a Verrazzano CR
//  WHEN AppendOverrides is called
//  THEN AppendOverrides should add registry overrides
func TestAppendRegistryOverrides(t *testing.T) {
	ctx := spi.NewFakeContext(fake.NewClientBuilder().WithScheme(getScheme()).Build(), &vzAcmeDev, false)
	registry := "foobar"
	imageRepo := "barfoo"
	kvs, _ := AppendOverrides(ctx, "", "", "", []bom.KeyValue{})
	assert.Equal(t, 8, len(kvs)) // should only have LetsEncrypt + useBundledSystemChart Overrides
	_ = os.Setenv(constants.RegistryOverrideEnvVar, registry)
	kvs, _ = AppendOverrides(ctx, "", "", "", []bom.KeyValue{})
	assert.Equal(t, 9, len(kvs))
	v, ok := getValue(kvs, systemDefaultRegistryKey)
	assert.True(t, ok)
	assert.Equal(t, registry, v)

	_ = os.Setenv(constants.ImageRepoOverrideEnvVar, imageRepo)
	kvs, _ = AppendOverrides(ctx, "", "", "", []bom.KeyValue{})
	assert.Equal(t, 9, len(kvs))
	v, ok = getValue(kvs, systemDefaultRegistryKey)
	assert.True(t, ok)
	assert.Equal(t, fmt.Sprintf("%s/%s", registry, imageRepo), v)
}

// TestAppendCAOverrides verifies that CA overrides are added as appropriate for private CAs
// GIVEN a Verrzzano CR
//  WHEN AppendOverrides is called
//  THEN AppendOverrides should add private CA overrides
func TestAppendCAOverrides(t *testing.T) {
	ctx := spi.NewFakeContext(fake.NewClientBuilder().WithScheme(getScheme()).Build(), &vzDefaultCA, false)
	kvs, err := AppendOverrides(ctx, "", "", "", []bom.KeyValue{})
	assert.Nil(t, err)
	v, ok := getValue(kvs, ingressTLSSourceKey)
	assert.True(t, ok)
	assert.Equal(t, caTLSSource, v)
	v, ok = getValue(kvs, privateCAKey)
	assert.True(t, ok)
	assert.Equal(t, privateCAValue, v)
}

// TestIsReady verifies Rancher is enabled or disabled as expected
// GIVEN a Verrzzano CR
//  WHEN IsEnabled is called
//  THEN IsEnabled should return true/false depending on the enabled state of the CR
func TestIsEnabled(t *testing.T) {
	enabled := true
	disabled := false
	c := fake.NewClientBuilder().WithScheme(getScheme()).Build()
	vzWithRancher := vzapi.Verrazzano{
		Spec: vzapi.VerrazzanoSpec{
			Components: vzapi.ComponentSpec{
				Rancher: &vzapi.RancherComponent{
					Enabled: &enabled,
				},
			},
		},
	}
	vzNoRancher := vzapi.Verrazzano{
		Spec: vzapi.VerrazzanoSpec{
			Components: vzapi.ComponentSpec{
				Rancher: &vzapi.RancherComponent{
					Enabled: &disabled,
				},
			},
		},
	}
	var tests = []struct {
		testName string
		ctx      spi.ComponentContext
		enabled  bool
	}{
		{
			"should be enabled",
			spi.NewFakeContext(c, &vzWithRancher, false),
			true,
		},
		{
			"should not be enabled",
			spi.NewFakeContext(c, &vzNoRancher, false),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			r := NewComponent()
			assert.Equal(t, tt.enabled, r.IsEnabled(tt.ctx.EffectiveCR()))
		})
	}
}

func TestPreInstall(t *testing.T) {
	caSecret := createCASecret()
	c := fake.NewClientBuilder().WithScheme(getScheme()).WithObjects(&caSecret).Build()
	ctx := spi.NewFakeContext(c, &vzDefaultCA, false)
	assert.Nil(t, NewComponent().PreInstall(ctx))
}

// TestIsReady verifies that a ready-state Rancher shows as ready
// GIVEN a ready Rancher install
//  WHEN IsReady is called
//  THEN IsReady should return true
func TestIsReady(t *testing.T) {
	readyClient := fake.NewClientBuilder().WithScheme(getScheme()).WithObjects(
		newReadyDeployment(ComponentNamespace, ComponentName),
		newPod(ComponentNamespace, ComponentName),
		newReplicaSet(ComponentNamespace, ComponentName),
		newReadyDeployment(ComponentNamespace, rancherWebhookDeployment),
		newPod(ComponentNamespace, rancherWebhookDeployment),
		newReplicaSet(ComponentNamespace, rancherWebhookDeployment),
		newReadyDeployment(FleetLocalSystemNamespace, fleetAgentDeployment),
		newPod(FleetLocalSystemNamespace, fleetAgentDeployment),
		newReplicaSet(FleetLocalSystemNamespace, fleetAgentDeployment),
		newReadyDeployment(FleetSystemNamespace, fleetControllerDeployment),
		newPod(FleetSystemNamespace, fleetControllerDeployment),
		newReplicaSet(FleetSystemNamespace, fleetControllerDeployment),
		newReadyDeployment(FleetSystemNamespace, gitjobDeployment),
		newPod(FleetSystemNamespace, gitjobDeployment),
		newReplicaSet(FleetSystemNamespace, gitjobDeployment),
	).Build()
	unreadyDeployClient := fake.NewClientBuilder().WithScheme(getScheme()).WithObjects(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ComponentNamespace,
				Name:      ComponentName,
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas: 0,
				Replicas:          1,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ComponentNamespace,
				Name:      rancherWebhookDeployment,
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas: 0,
				Replicas:          1,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: FleetLocalSystemNamespace,
				Name:      fleetAgentDeployment,
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas: 0,
				Replicas:          1,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: FleetSystemNamespace,
				Name:      fleetControllerDeployment,
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas: 0,
				Replicas:          1,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: FleetSystemNamespace,
				Name:      gitjobDeployment,
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas: 0,
				Replicas:          1,
			},
		},
	).Build()

	SetFakeCheckRancherUpgradeFailureFunc()
	defer SetDefaultCheckRancherUpgradeFailureFunc()
	var tests = []struct {
		testName string
		ctx      spi.ComponentContext
		isReady  bool
	}{
		{
			"should be ready",
			spi.NewFakeContext(readyClient, &vzDefaultCA, true),
			true,
		},
		{
			"should not be ready due to deployment",
			spi.NewFakeContext(unreadyDeployClient, &vzDefaultCA, true),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.isReady, NewComponent().IsReady(tt.ctx))
		})
	}
}

// TestPostInstall tests a happy path post install run
// GIVEN a Rancher install state where all components are ready
//  WHEN PostInstall is called
//  THEN PostInstall should return nil
func TestPostInstall(t *testing.T) {
	component := NewComponent()
	ctxWithoutIngress, ctxWithIngress := prepareContexts()
	assert.IsType(t, fmt.Errorf(""), component.PostInstall(ctxWithoutIngress))
	assert.Nil(t, component.PostInstall(ctxWithIngress))
}

// TestPostUpgrade tests a happy path post upgrade run
// GIVEN a Rancher install state where all components are ready
//  WHEN PostUpgrade is called
//  THEN PostUpgrade should return nil
func TestPostUpgrade(t *testing.T) {
	component := NewComponent()
	ctxWithoutIngress, ctxWithIngress := prepareContexts()
	assert.Nil(t, component.PostUpgrade(ctxWithoutIngress))
	assert.Nil(t, component.PostUpgrade(ctxWithIngress))
}

func Test_rancherComponent_ValidateUpdate(t *testing.T) {
	disabled := false
	tests := []struct {
		name    string
		old     *vzapi.Verrazzano
		new     *vzapi.Verrazzano
		wantErr bool
	}{
		{
			name: "enable",
			old: &vzapi.Verrazzano{
				Spec: vzapi.VerrazzanoSpec{
					Components: vzapi.ComponentSpec{
						Rancher: &vzapi.RancherComponent{
							Enabled: &disabled,
						},
					},
				},
			},
			new:     &vzapi.Verrazzano{},
			wantErr: false,
		},
		{
			name: "disable",
			old:  &vzapi.Verrazzano{},
			new: &vzapi.Verrazzano{
				Spec: vzapi.VerrazzanoSpec{
					Components: vzapi.ComponentSpec{
						Rancher: &vzapi.RancherComponent{
							Enabled: &disabled,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:    "no change",
			old:     &vzapi.Verrazzano{},
			new:     &vzapi.Verrazzano{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewComponent()
			if err := c.ValidateUpdate(tt.old, tt.new); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func prepareContexts() (spi.ComponentContext, spi.ComponentContext) {
	// mock the k8s resources used in post install
	caSecret := createCASecret()
	rootCASecret := createRootCASecret()
	adminSecret := createAdminSecret()
	rancherPodList := createRancherPodListWithAllRunning()
	ingress := v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: common.CattleSystem,
			Name:      constants.RancherIngress,
		},
	}
	time := metav1.Now()
	cert := certapiv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{Name: certificates[0].Name, Namespace: certificates[0].Namespace},
		Status: certapiv1.CertificateStatus{
			Conditions: []certapiv1.CertificateCondition{
				{Type: certapiv1.CertificateConditionReady, Status: cmmeta.ConditionTrue, LastTransitionTime: &time},
			},
		},
	}
	serverURLSetting := createServerURLSetting()
	ociDriver := createOciDriver()
	okeDriver := createOkeDriver()

	clientWithoutIngress := fake.NewClientBuilder().WithScheme(getScheme()).WithObjects(&caSecret, &rootCASecret, &adminSecret, &rancherPodList.Items[0], &serverURLSetting, &ociDriver, &okeDriver).Build()
	ctxWithoutIngress := spi.NewFakeContext(clientWithoutIngress, &vzDefaultCA, false)

	clientWithIngress := fake.NewClientBuilder().WithScheme(getScheme()).WithObjects(&caSecret, &rootCASecret, &adminSecret, &rancherPodList.Items[0], &ingress, &cert, &serverURLSetting, &ociDriver, &okeDriver).Build()
	ctxWithIngress := spi.NewFakeContext(clientWithIngress, &vzDefaultCA, false)

	// mock the pod executor when resetting the Rancher admin password
	k8sutil.NewPodExecutor = k8sutilfake.NewPodExecutor
	k8sutilfake.PodSTDOUT = "password"
	k8sutil.ClientConfig = func() (*rest.Config, kubernetes.Interface, error) {
		config, k := k8sutilfake.NewClientsetConfig()
		return config, k, nil
	}

	return ctxWithoutIngress, ctxWithIngress

}

func newReadyDeployment(namespace string, name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    map[string]string{"app": name},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 1,
			Replicas:          1,
			UpdatedReplicas:   1,
		},
	}
}

func newPod(namespace string, name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name + "-95d8c5d96-m6mbr",
			Labels: map[string]string{
				"pod-template-hash": "95d8c5d96",
				"app":               name,
			},
		},
	}
}

func newReplicaSet(namespace string, name string) *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   namespace,
			Name:        name + "-95d8c5d96",
			Annotations: map[string]string{"deployment.kubernetes.io/revision": "1"},
		},
	}
}
