// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package integ_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano/application-operator/test/integ/k8s"
	"github.com/verrazzano/verrazzano/application-operator/test/integ/util"
	"time"
)

const verrazzanoOperator = "verrazzano-application-operator"
const verrazzanoApplication = "verrazzano-application"

const appService = "hello-workload"
const appPodPrefix = "hello-workload"
const appDeployment = "hello-workload"
const appNamespace = "default"

var fewSeconds = 2 * time.Second
var tenSeconds = 10 * time.Second
var thirtySeconds = 30 * time.Second
var oneMinute = 60 * time.Second
var K8sClient k8s.Client

var _ = BeforeSuite(func() {
	var err error
	K8sClient, err = k8s.NewClient()
	if err != nil {
		Fail(fmt.Sprintf("Error creating Kubernetes client to access Verrazzano API objects: %v", err))
	}
})

var _ = AfterSuite(func() {
})

var _ = Describe("Custom Resource Definition for OAM controller runtime", func() {
	It("applicationconfigurations.core.oam.dev exists", func() {
		Expect(K8sClient.DoesCRDExist("applicationconfigurations.core.oam.dev")).To(BeTrue(),
			"The applicationconfigurations.core.oam.dev CRD should exist")
	})
	It("components.core.oam.dev exists", func() {
		Expect(K8sClient.DoesCRDExist("components.core.oam.dev")).To(BeTrue(),
			"The components.core.oam.dev CRD should exist")
	})
	It("containerizedworkloads.core.oam.dev  exists", func() {
		Expect(K8sClient.DoesCRDExist("containerizedworkloads.core.oam.dev")).To(BeTrue(),
			"The containerizedworkloads.core.oam.dev  CRD should exist")
	})
	It("healthscopes.core.oam.dev exists", func() {
		Expect(K8sClient.DoesCRDExist("healthscopes.core.oam.dev")).To(BeTrue(),
			"The healthscopes.core.oam.dev CRD should exist")
	})
	It("ingresstraits.oam.verrazzano.io exists", func() {
		Expect(K8sClient.DoesCRDExist("ingresstraits.oam.verrazzano.io")).To(BeTrue(),
			"The ingresstraits.oam.verrazzano.io  CRD should exist")
	})
	It("manualscalertraits.core.oam.dev exists", func() {
		Expect(K8sClient.DoesCRDExist("manualscalertraits.core.oam.dev")).To(BeTrue(),
			"The manualscalertraits.core.oam.dev  CRD should exist")
	})
	It("scopedefinitions.core.oam.dev exists", func() {
		Expect(K8sClient.DoesCRDExist("scopedefinitions.core.oam.dev")).To(BeTrue(),
			"The scopedefinitions.core.oam.dev  CRD should exist")
	})
})

var _ = Describe("Custom Resource Definition for Verrazzano CRDs", func() {
	It("ingresstraits.oam.verrazzano.io exists", func() {
		Expect(K8sClient.DoesCRDExist("ingresstraits.oam.verrazzano.io")).To(BeTrue(),
			"The ingresstraits.oam.verrazzano.io CRD should exist")
	})
})

var _ = Describe("verrazzano-application namespace resources ", func() {
	It(fmt.Sprintf("Namespace %s exists", verrazzanoApplication), func() {
		Expect(K8sClient.DoesNamespaceExist(verrazzanoApplication)).To(BeTrue(),
			"The namespace should exist")
	})

	It(fmt.Sprintf("ServiceAccount %s exists", verrazzanoOperator), func() {
		Expect(K8sClient.DoesServiceAccountExist(verrazzanoOperator, verrazzanoApplication)).To(BeTrue(),
			"The verrazzano operator service account should exist")
	})
	It(fmt.Sprintf("Deployment %s exists", verrazzanoOperator), func() {
		Expect(K8sClient.DoesDeploymentExist(verrazzanoOperator, verrazzanoApplication)).To(BeTrue(),
			"The verrazzano operator doesn't exist")
	})
	It(fmt.Sprintf("Pod prefixed by %s exists", verrazzanoOperator), func() {
		Expect(K8sClient.DoesPodExist(verrazzanoOperator, verrazzanoApplication)).To(BeTrue(),
			"The verrazzano operator pod doesn't exist")
	})
})

var _ = Describe("Testing hello app lifecycle", func() {
	It("apply component should result in a component in default namespace", func() {
		_, stderr := util.RunCommand("kubectl apply -f testdata/hello-comp.yaml")
		Expect(stderr).To(Equal(""))
		//	Eventually(appComponentExists, fewSeconds).Should(BeTrue())
	})
	It("apply app config should result in a app config in default namespace", func() {
		_, stderr := util.RunCommand("kubectl apply -f testdata/hello-app.yaml")
		Expect(stderr).To(Equal(""))
		//	Eventually(appBindingExists, fewSeconds).Should(BeTrue())
	})
	It("hello deployment should exist ", func() {
		Eventually(appDeploymentExists, tenSeconds).Should(BeTrue())
	})
	It("hello pod should exist ", func() {
		Eventually(appPodExists, oneMinute).Should(BeTrue())
	})
	It("hello service should exist ", func() {
		Expect(K8sClient.DoesServiceExist(appService, appNamespace)).To(BeTrue(),
			"The hello service should exist")
	})
	It("deleting app config", func() {
		_, stderr := util.RunCommand("kubectl delete -f testdata/hello-app.yaml")
		Expect(stderr).To(Equal(""))
		//	Eventually(appBindingExists, fewSeconds).Should(BeFalse())
	})
	It("deleting app component", func() {
		_, stderr := util.RunCommand("kubectl delete -f testdata/hello-comp.yaml")
		Expect(stderr).To(Equal(""))
		//	Eventually(appModelExists, fewSeconds).Should(BeFalse())
	})
	It("hello deployment should  not exist ", func() {
		Eventually(appDeploymentExists, tenSeconds).Should(BeFalse())
	})
	It("hello pod should not exist ", func() {
		Eventually(appPodExists, oneMinute).Should(BeFalse())
	})
	It("hello service should exist ", func() {
		Expect(K8sClient.DoesServiceExist(appService, appNamespace)).To(BeFalse(),
			"The hello service should exist")
	})
})

//// Helper functions
func appNsExists() bool {
	return K8sClient.DoesNamespaceExist(appNamespace)
}
func appDeploymentExists() bool {
	return K8sClient.DoesDeploymentExist(appDeployment, appNamespace)
}
func appPodExists() bool {
	return K8sClient.DoesPodExist(appPodPrefix, appNamespace)
}
