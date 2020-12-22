#!/usr/bin/env bash
#
# Copyright (c) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#

SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)

if [ -z "${VERRAZZANO_APP_OP_IMAGE:-}" ] ; then
  echo "Environment variable VERRAZZANO_APP_OP_IMAGE must be set to the Verrazzano application operator image"
  exit 1
fi

set -e
"$SCRIPT_DIR"/scripts/1-install-oam-runtime.sh
"$SCRIPT_DIR"/scripts/2-install-wls-operator.sh
"$SCRIPT_DIR"/scripts/3-install-vz-oam.sh
