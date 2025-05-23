#!/bin/sh
# Copyright 2025 whereabouts authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# SPDX-License-Identifier: Apache-2.0


set -u -e

source /lib.sh

echo "Sleep and Watching for service account token and CA file changes..."
# enter sleep/watch loop
while true; do
  # Check the md5sum of the service account token and ca.
  svcaccountsum="$(get_token_md5sum)"
  casum="$(get_ca_file_md5sum)"
  if [ "$svcaccountsum" != "$LAST_SERVICEACCOUNT_MD5SUM" ] || ! [ "$SKIP_TLS_VERIFY" == "true" ] && [ "$casum" != "$LAST_KUBE_CA_FILE_MD5SUM" ]; then
    log "Detected service account or CA file change, regenerating kubeconfig..."
    generateKubeConfig
    LAST_SERVICEACCOUNT_MD5SUM="$svcaccountsum"
    LAST_KUBE_CA_FILE_MD5SUM="$casum"
  fi

  sleep 1s
done
