#!/usr/bin/env bash
#
# Copyright (c) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)
PROJ_DIR=$(cd $(dirname "$0"); cd ../..; pwd -P)

. $SCRIPT_DIR/common.sh

# This makes an attempt to uninstall OAM, ignoring errors so that this can work
# in the case where there is a partial installation

function uninstall {
  log "Uninstalling Verrazzano application operator"
  kubectl delete -f ${PROJ_DIR}/deploy/verrazzano.yaml

  log "Delete Verrazzano IngressTrait trait definition"
  kubectl delete -f ${PROJ_DIR}/deploy/traitdefinition_ingresstrait.yaml

  log "Delete Verrazzano WebLogic workload definition"
  kubectl delete -f ${PROJ_DIR}/deploy/workload-wls.yaml

  log "Deleting Verrazzano IngressTrait CRD"
  kubectl delete -f ${PROJ_DIR}/config/crd/bases/oam.verrazzano.io_ingresstraits.yaml
}

action "Uninstalling Verrazzano application operator " uninstall
