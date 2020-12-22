#!/usr/bin/env bash
#
# Copyright (c) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
SCRIPT_DIR=$(cd $(dirname $0); pwd -P)

set -u

NAMESPACE="oam-hello-helidon"
SECRET="image-pull-secret"

echo "Removing Helidon hello world OAM application."

echo "Delete application definition."
kubectl get applicationconfiguration hello-helidon-appconf -n "${NAMESPACE}" &> /dev/null
if [ $? -ne 0 ]; then
  echo "No application definition not found, skipping."
else
  kubectl delete -f ${SCRIPT_DIR}/
  code=$?
  if [ ${code} -ne 0 ]; then
    echo "ERROR: Deleting application definition failed: ${code}. Exiting."
    exit ${code}
  fi
fi

echo "Delete application ingress."
kubectl get service hello-helidon-ingress -n "${NAMESPACE}" &> /dev/null
if [ $? -ne 0 ]; then
  echo "Application ingress not found, skipping."
else
  kubectl delete service hello-helidon-ingress -n "${NAMESPACE}"
  code=$?
  if [ ${code} -ne 0 ]; then
    echo "ERROR: Deleting application ingress failed: ${code}. Listing services and exiting."
    kubectl get services -n "${NAMESPACE}"
    exit ${code}
  fi
fi

echo "Wait for termination of application pod."
attempt=1
while true; do
  count=$(kubectl get pods -n "${NAMESPACE}" 2> /dev/null | wc -l)
  if [ $count -eq 0 ]; then
    echo "No application pods found on attempt ${attempt}."
    break
  elif [ ${attempt} -eq 1 ]; then
    echo "Application pods found on initial attempt. Retrying after delay."
  elif [ ${attempt} -ge 30 ]; then
    echo "ERROR: Application pods found after ${attempt} attempts. Listing pods and exiting."
    kubectl get pods -n "${NAMESPACE}"
    exit 1
  fi
  attempt=$(($attempt+1))
  sleep 1
done

echo "Delete secret."
kubectl get secret "${SECRET}" -n "${NAMESPACE}" &> /dev/null
if [ $? -ne 0 ]; then
  echo "No secret found, skipping."
else
  kubectl delete secret "${SECRET}" -n "${NAMESPACE}"
  code=$?
  if [ ${code} -ne 0 ]; then
    echo "ERROR: Deleting secret failed: ${code}. Exiting."
    exit ${code}
  fi
fi

echo "Delete namespace."
kubectl get namespace "${NAMESPACE}"
if [ $? -ne 0 ]; then
  echo "No namespace found, skipping."
else
  kubectl delete namespace "${NAMESPACE}"
  code=$?
  if [ ${code} -ne 0 ]; then
    echo "ERROR: Deleting namespace failed: ${code}. Listing namespaces and exiting."
    kubectl get namespaces
    exit ${code}
  fi
fi

echo "Removal of Helidon hello world OAM application successful."
