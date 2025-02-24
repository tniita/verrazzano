// Copyright (C) 2020, 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package util

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
)

// GetKubeconfig returns the KubeConfig in string format
func GetKubeconfig() string {
	var kubeconfig string
	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}
	if kubeconfig = os.Getenv("VERRAZZANO_KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}
	if home := homeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		ginkgo.Fail("Could not get kubeconfig")
	}
	return kubeconfig
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return ""
}
