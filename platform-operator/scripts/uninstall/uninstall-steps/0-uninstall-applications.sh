#!/bin/bash
#
# Copyright (c) 2020, 2021, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)
INSTALL_DIR=$SCRIPT_DIR/../../install
UNINSTALL_DIR=$SCRIPT_DIR/..

. $INSTALL_DIR/common.sh
. $INSTALL_DIR/config.sh
. $UNINSTALL_DIR/uninstall-utils.sh

set -o pipefail

function initializing_uninstall {
  # Deleting rancher through API
  log "Listing pods in cattle-system namespace"
  kubectl get pods -n cattle-system
  log "Deleting Rancher through API"
  rancher_exists=$(kubectl get namespace cattle-system) || return 0
  rancher_host_name="$(kubectl get ingress -n cattle-system --no-headers -o custom-columns=":spec.rules[0].host")" || err_return $? "Could not retrieve Rancher hostname" || return 0
  rancher_cluster_url="https://${rancher_host_name}/v3/clusters/local"
  rancher_admin_password=$(kubectl get secret --namespace cattle-system rancher-admin-secret -o jsonpath={.data.password}) || err_return $? "Could not retrieve rancher-admin-secret" || return 0
  rancher_admin_password=$(echo ${rancher_admin_password} | base64 --decode) || err_return $? "Could not decode rancher-admin-secret" || return 0

  if [ "$rancher_admin_password" ] && [ "$rancher_host_name" ] ; then
    log "Retrieving Rancher access token."
    get_rancher_access_token "${rancher_host_name}" "${rancher_admin_password}"
  fi

  if [ "${RANCHER_ACCESS_TOKEN}" ]; then
    log "Updating ${rancher_cluster_url}"
    local temp_output="/tmp/delete_cluster.out"
    status=$(curl -o ${temp_output} -s -w "%{http_code}\n" $(get_rancher_resolve ${rancher_hostname}) -X DELETE -H "Accept: application/json" -H "Authorization: Bearer ${RANCHER_ACCESS_TOKEN}" --insecure "${rancher_cluster_url}")
    if [ "$status" != 200 ] ; then
      local cluster_delete_output=$(cat $temp_output)
      log "${cluster_delete_output}"
      rm "$temp_output"
      return 0
    fi

    # Wait 60s for local cluster to delete
    local max_retries=6
    local retries=0
    while true ; do
      still_exists="$(curl -s $(get_rancher_resolve ${rancher_hostname}) -X GET -H "Accept: application/json" -H "Authorization: Bearer ${RANCHER_ACCESS_TOKEN}" --insecure "${rancher_cluster_url}")"
      state="$(echo "$still_exists" | jq -r ".state" )"
      if [ "$state" != "active" ] && [ "$state" != "removing" ] ; then
        break
      else
        log "Rancher cluster is still in state: ${state}"
        sleep 10
      fi
      ((retries+=1))
      if [ "$retries" -ge "$max_retries" ] ; then
        break
      fi
    done
  fi
  return 0
}

# Delete all of the OAM ApplicationConfiguration resources in all namespaces.
function delete_oam_applications_configurations {
  delete_k8s_resource_from_all_namespaces applicationconfigurations.core.oam.dev
}

# Delete all of the OAM Component resources in all namespaces.
function delete_oam_components {
  delete_k8s_resource_from_all_namespaces components.core.oam.dev
}

action "Initializing Uninstall" initializing_uninstall || exit 1
action "Deleting OAM application configurations" delete_oam_applications_configurations || exit 1
action "Deleting OAM components" delete_oam_components || exit 1
