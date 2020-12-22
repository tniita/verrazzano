#!/usr/bin/env bash
#
# Copyright (c) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)
PROJ_DIR=$(cd $(dirname "$0"); cd ../..; pwd -P)
BUILD_DEPLOY=${PROJ_DIR}/build/deploy

if [ -z "${VERRAZZANO_APP_OP_IMAGE:-}" ] ; then
  echo "Environment variable VERRAZZANO_APP_OP_IMAGE must be set to the Verrazzano application operator image in ghir.io"
  exit 1
fi

. $SCRIPT_DIR/common.sh

function install_traits {
  log "Create verrazzano IngressTrait CRD"
  kubectl apply -f ${PROJ_DIR}/config/crd/bases/oam.verrazzano.io_ingresstraits.yaml
  if [ $? -ne 0 ]; then
    error "Failed to create verrazzano IngressTrait CRD."
    return 1
  fi

  log "Create Verrazzano IngressTrait trait definition"
  kubectl apply -f ${PROJ_DIR}/deploy/traitdefinition_ingresstrait.yaml
  if [ $? -ne 0 ]; then
    error "Failed to create verrazzano IngressTrait CRD."
    return 1
  fi
}

function install_workloads {
  log "Create Verrazzano WebLogic workload definition"
  kubectl apply -f ${PROJ_DIR}/deploy/workload-wls.yaml
  if [ $? -ne 0 ]; then
    error "Failed to create WebLogic workload CRD."
    return 1
  fi
}

function install_vz_operator {
  log "Install Verrazzano application operator"
  kubectl apply -f ${BUILD_DEPLOY}/verrazzano.yaml
  if [ $? -ne 0 ]; then
    error "Failed to install Verazzano application operator."
    return 1
  fi
}

# Update the image name in the verrazzano deployment file
mkdir -p ${BUILD_DEPLOY}
cat ${PROJ_DIR}/deploy/verrazzano.yaml | sed -e "s|IMAGE_NAME|${VERRAZZANO_APP_OP_IMAGE}|g" > ${BUILD_DEPLOY}/verrazzano.yaml

# do the install
action "Installing Verrazzano OAM traits " install_traits || fail "Failed to install Verrazzano OAM traits."
action "Installing Verrazzano OAM workloads " install_workloads || fail "Failed to install Verrazzano OAM workloads."
action "Installing Verrazzano application operator " install_vz_operator || fail "Failed to install Verrazzano application operator."
