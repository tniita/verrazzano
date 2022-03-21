// Copyright (c) 2021, 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package mysql

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/verrazzano/verrazzano/pkg/bom"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/istio"

	vzapi "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"

	vzconst "github.com/verrazzano/verrazzano/platform-operator/constants"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/secret"
	"github.com/verrazzano/verrazzano/platform-operator/internal/config"
)

// ComponentName is the name of the component
const ComponentName = "mysql"

// ComponentNamespace is the namespace of the component
const ComponentNamespace = vzconst.KeycloakNamespace

// ComponentJSONName is the josn name of the verrazzano component in CRD
const ComponentJSONName = "keycloak.mysql"

// mysqlComponent represents an MySQL component
type mysqlComponent struct {
	helm.HelmComponent
}

// Verify that mysqlComponent implements Component
var _ spi.Component = mysqlComponent{}

// NewComponent returns a new MySQL component
func NewComponent() spi.Component {
	return mysqlComponent{
		helm.HelmComponent{
			ReleaseName:             ComponentName,
			JSONName:                ComponentJSONName,
			ChartDir:                filepath.Join(config.GetThirdPartyDir(), ComponentName),
			ChartNamespace:          ComponentNamespace,
			IgnoreNamespaceOverride: true,
			SupportsOperatorInstall: true,
			ImagePullSecretKeyname:  secret.DefaultImagePullSecretKeyName,
			ValuesFile:              filepath.Join(config.GetHelmOverridesDir(), "mysql-values.yaml"),
			AppendOverridesFunc:     appendMySQLOverrides,
			Dependencies:            []string{istio.ComponentName},
		},
	}
}

// IsReady calls MySQL isMySQLReady function
func (c mysqlComponent) IsReady(context spi.ComponentContext) bool {
	if c.HelmComponent.IsReady(context) {
		return isMySQLReady(context)
	}
	return false
}

// IsEnabled mysql-specific enabled check for installation
// If keycloak is enabled, mysql is enabled; disabled otherwise
func (c mysqlComponent) IsEnabled(effectiveCR *vzapi.Verrazzano) bool {
	comp := effectiveCR.Spec.Components.Keycloak
	if comp == nil || comp.Enabled == nil {
		return true
	}
	return *comp.Enabled
}

// PreInstall calls MySQL preInstall function
func (c mysqlComponent) PreInstall(ctx spi.ComponentContext) error {
	return preInstall(ctx, c.ChartNamespace)
}

// PostInstall calls MySQL postInstall function
func (c mysqlComponent) PostInstall(ctx spi.ComponentContext) error {
	return postInstall(ctx)
}

// ValidateUpdate checks if the specified new Verrazzano CR is valid for this component to be updated
func (c mysqlComponent) ValidateUpdate(old *vzapi.Verrazzano, new *vzapi.Verrazzano) error {
	// Block all changes for now, particularly around storage changes

	// compare the VolumeSourceOverrides and reject if the type or size or storage class is different
	oldSetting, err := doGenerateVolumeSourceOverrides(old, []bom.KeyValue{})
	if err != nil {
		return err
	}
	newSetting, err := doGenerateVolumeSourceOverrides(new, []bom.KeyValue{})
	if err != nil {
		return err
	}
	// Reject any persistence-specific changes via the mysqlInstallArgs settings
	if bom.FindKV(oldSetting, "persistence.enabled") != bom.FindKV(newSetting, "persistence.enabled") {
		return fmt.Errorf("Can not change persistence enabled setting in component: %s", ComponentJSONName)
	}
	if bom.FindKV(oldSetting, "persistence.size") != bom.FindKV(newSetting, "persistence.size") {
		return fmt.Errorf("Can not change persistence volume size in component: %s", ComponentJSONName)
	}
	if bom.FindKV(oldSetting, "persistence.storageClass") != bom.FindKV(newSetting, "persistence.storageClass") {
		return fmt.Errorf("Can not change persistence storage class in component: %s", ComponentJSONName)
	}
	// Reject any installArgs changes for now
	if !reflect.DeepEqual(c.getInstallArgs(old), c.getInstallArgs(new)) {
		return fmt.Errorf("Updates to mysqlInstallArgs not allowed for %s", ComponentJSONName)
	}
	return nil
}

func (c mysqlComponent) getInstallArgs(vz *vzapi.Verrazzano) []vzapi.InstallArgs {
	if vz != nil && vz.Spec.Components.Keycloak != nil {
		return vz.Spec.Components.Keycloak.MySQL.MySQLInstallArgs
	}
	return nil
}
