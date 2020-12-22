#!/usr/bin/env bash
#
# Copyright (c) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)

. $SCRIPT_DIR/common.sh

# This makes an attempt to uninstall OAM, ignoring errors so that this can work
# in the case where there is a partial installation
function uninstall_oam {

  log "Uninstall OAM"
  helm delete oam --namespace oam-system crossplane-master/oam-kubernetes-runtime

  log "Delete OAM roles"
  kubectl delete clusterrole oam-kubernetes-runtime-oam:system:aggregate-to-controller
  kubectl delete clusterrolebinding oam-kubernetes-runtime-oam:system:aggregate-to-controller
  kubectl delete clusterrolebinding cluster-admin-binding-oam
  kubectl delete ScopeDefinition healthscopes.core.oam.dev
  kubectl delete TraitDefinition manualscalertraits.core.oam.dev
  kubectl delete WorkloadDefinition containerizedworkloads.core.oam.dev

  log "Delete oam-system namespace"
  kubectl delete namespace oam-system

}

action "Uninstalling OAM runtime" uninstall_oam
