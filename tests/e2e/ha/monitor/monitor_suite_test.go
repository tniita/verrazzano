// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package monitor

import (
	"flag"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano/pkg/test/framework"
)

var runContinuous bool
var t = framework.NewTestFramework("monitor")

func init() {
	flag.BoolVar(&runContinuous, "runContinuous", true, "run monitors continuously if set")
}

func TestHAMonitor(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "HA Monitoring Suite")
}
