#!/usr/bin/env bash
#
# Copyright (c) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)

. $SCRIPT_DIR/common.sh

function install_oam {
  log "Create oam-system namespace"
  kubectl create namespace oam-system
  if [ $? -ne 0 ]; then
    error "Failed to create oam-system namespace."
    return 1
  fi

  log "Add OAM helm repository"
  helm repo add crossplane-master https://charts.crossplane.io/master/
  if [ $? -ne 0 ]; then
    error "Failed add OAM helm repository."
    return 1
  fi

  log "Install OAM"
  helm install oam --namespace oam-system crossplane-master/oam-kubernetes-runtime --devel
  if [ $? -ne 0 ]; then
    error "Failed to OAM helm install."
    return 1
  fi

  log "Setup OAM roles"
  kubectl create clusterrolebinding cluster-admin-binding-oam --clusterrole cluster-admin --user system:serviceaccount:oam-system:oam-kubernetes-runtime-oam
  if [ $? -ne 0 ]; then
    error "Failed to setup OAM roles."
    return 1
  fi
}

action "Installing OAM runtime" install_oam || fail "Failed to install OAM runtime."
