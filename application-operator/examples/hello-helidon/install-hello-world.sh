#!/usr/bin/env bash
#
# Copyright (c) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
SCRIPT_DIR=$(cd $(dirname $0); pwd -P)

set -u

DOCKER_SVR="${1:-$GHCR_REPO}"
DOCKER_USR="${2:-$GITHUB_PKGS_CREDS_USR}"
DOCKER_PWD="${3:-$GITHUB_PKGS_CREDS_PSW}"

NAMESPACE="oam-hello-helidon"
SECRET="image-pull-secret"

if [ -z "${DOCKER_SVR}" ]; then
  echo "ERROR: Container registry required as first argument or GHCR_REPO environment variable."
  exit 1
fi
if [ -z "${DOCKER_USR}" ]; then
  echo "ERROR: Container registry username required as second argument or GITHUB_PKGS_CREDS_USR environment variable."
  exit 1
fi
if [ -z "${DOCKER_PWD}" ]; then
  echo "ERROR: Container registry username required as second argument or GITHUB_PKGS_CREDS_PSW environment variable."
  exit 1
fi

echo "Installing Helidon hello world OAM application."

echo "Wait for OAM runtime pod to be ready."
attempt=1
while true; do
  kubectl -n oam-system wait --for=condition=ready pods --selector='app.kubernetes.io/name=oam-kubernetes-runtime' --timeout 15s
  if [ $? -eq 0 ]; then
    echo "OAM runtime pods found ready on attempt ${attempt}."
    break
  elif [ ${attempt} -eq 1 ]; then
    echo "No OAM runtime pods found ready on initial attempt. Retrying after delay."
  elif [ ${attempt} -ge 20 ]; then
    echo "ERROR: No OAM runtime pods found ready after ${attempt} attempts. Listing pods."
    kubectl get pods -n oam-system
    echo "ERROR: Exiting."
    exit 1
  fi
  attempt=$(($attempt+1))
  sleep .5
done

status=$(kubectl get namespace ${NAMESPACE} -o jsonpath="{.status.phase}" 2> /dev/null)
if [ "${status}" == "Active" ]; then
  echo "Found namespace ${NAMESPACE}."
else
  echo "Create namespace ${NAMESPACE}."
  kubectl create namespace "${NAMESPACE}"
  if [ $? -ne 0 ]; then
      echo "ERROR: Failed to create namespace ${NAMESPACE}, exiting."
      exit 1
  fi
fi

echo "Create image pull secret."
if [ "${skip_secrets:-false}" != "true" ]; then
  kubectl get secret "${SECRET}" -n "${NAMESPACE}" &> /dev/null
  if [ $? -eq 0 ]; then
    echo "Delete existing secret."
    kubectl delete secret "${SECRET}" -n "${NAMESPACE}"
  fi
  kubectl create secret docker-registry "${SECRET}" -n "${NAMESPACE}"\
    --docker-server="${DOCKER_SVR}" \
    --docker-username="${DOCKER_USR}" \
    --docker-password="${DOCKER_PWD}"
  if [ $? -ne 0 ]; then
      echo "ERROR: Failed to create image pull secret. Listing secrets."
      kubectl get secret "${SECRET}" -n "${NAMESPACE}"
      exit 1
  fi
fi

echo "Apply application configuration."
kubectl apply -f ${SCRIPT_DIR}/
code=$?
if [ ${code} -ne 0 ]; then
  echo "ERROR: Applying application configuration failed: ${code}. Exiting."
  exit ${code}
fi

echo "Wait for at least one running workload pod."
attempt=1
while true; do
  kubectl -n "${NAMESPACE}" wait --for=condition=ready pods --selector='app.oam.dev/name=hello-helidon-appconf' --timeout 15s
  if [ $? -eq 0 ]; then
    echo "Application pods found ready on attempt ${attempt}."
    break
  elif [ ${attempt} -eq 1 ]; then
    echo "No application pods found ready on initial attempt. Retrying after delay."
  elif [ ${attempt} -ge 30 ]; then
    echo "ERROR: No application pod found ready after ${attempt} attempts. Listing pods."
    kubectl get pods -n "${NAMESPACE}"
    echo "ERROR: Exiting."
    exit 1
  fi
  attempt=$(($attempt+1))
  sleep .5
done

echo "Expose application via load balancer service."
kubectl get service hello-helidon-ingress -n "${NAMESPACE}" &> /dev/null
if [ $? -eq 0 ]; then
  echo "Application ingress already exists, skipping."
else
  kubectl expose service hello-helidon-workload -n "${NAMESPACE}" --port=8080 --target-port=8080 --type=LoadBalancer --name=hello-helidon-ingress
  code=$?
  if [ ${code} -ne 0 ]; then
    echo "ERROR: Exposing application failed: ${code}. Exiting."
    exit ${code}
  fi
fi

echo "Determine application endpoint."
attempt=1
while true; do
  host=$(kubectl get service -n "${NAMESPACE}" hello-helidon-ingress -o jsonpath={.status.loadBalancer.ingress[0].ip})
  if [[ ! -z "${host}" ]]; then
    echo "Application endpoint found on attempt ${attempt}, host \"${host}\"."
    break
  elif [ ${attempt} -eq 1 ]; then
    echo "No application endpoints found on initial attempt. Retrying after delay."
  elif [ ${attempt} -ge 60 ]; then
    echo "ERROR: No application endpoints found ater ${attempt} attempts. Listing services and exiting."
    kubectl get services -n "${NAMESPACE}"
    exit 1
  fi
  attempt=$(($attempt+1))
  sleep 1
done

port=8080
url="http://${host}:${port}/greet"
expect="Hello World"
echo "Connect to application endpoint ${url}"
attempt=1
while true; do
  reply=$(curl -s --connect-timeout 30 --retry 10 --retry-delay 30 -X GET ${url})
  code=$?
  if [ ${code} -ne 0 ] && [ ${attempt} -ge 15 ]; then
    echo "ERROR: Application connection failed: ${code}. Exiting."
    exit ${code}
  elif [ ${code} -ne 0 ] && [ ${attempt} -lt 15 ]; then
    echo "Connect to application endpoint ${url} failed with code ${code}. Retrying after delay."
  elif [[ "$reply" != *"${expect}"* ]]; then
    echo "ERROR: Application reply unexpected: ${reply}, expected: ${expect}. Exiting."
    exit 1
  else
    echo "Application reply correct: ${reply}"
    break
  fi
  attempt=$(($attempt+1))
  sleep 1
done

echo "Installation of Helidon hello world OAM application successful."