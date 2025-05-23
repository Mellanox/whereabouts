#!/bin/bash
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

set -eo pipefail

while true; do
  case "$1" in
    -n|--number-of-compute)
      NUMBER_OF_COMPUTE_NODES=$2
      break
      ;;
    *)
      echo "define argument -n (number of compute nodes)"
      exit 1
  esac
done

HERE="$(dirname "$(readlink --canonicalize ${BASH_SOURCE[0]})")"
ROOT="$(readlink --canonicalize "$HERE/..")"
MULTUS_DAEMONSET_URL="https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset.yml"
CNIS_DAEMONSET_PATH="$ROOT/hack/cni-install.yml"
TIMEOUT_K8="5000s"
RETRY_MAX=10
INTERVAL=10
TIMEOUT=300
TIMEOUT_K8="${TIMEOUT}s"
KIND_CLUSTER_NAME="whereabouts"
OCI_BIN="${OCI_BIN:-"docker"}"
IMG_PROJECT="whereabouts"
IMG_REGISTRY="ghcr.io/k8snetworkplumbingwg"
IMG_TAG="latest"
IMG_NAME="$IMG_REGISTRY/$IMG_PROJECT:$IMG_TAG"

create_cluster() {
workers="$(for i in $(seq $NUMBER_OF_COMPUTE_NODES); do echo "  - role: worker"; done)"
  # deploy cluster with kind
  cat <<EOF | kind create cluster --name $KIND_CLUSTER_NAME --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
$workers
EOF
}


check_requirements() {
  for cmd in "$OCI_BIN" kind kubectl; do
    if ! command -v "$cmd" &> /dev/null; then
      echo "$cmd is not available"
      exit 1
    fi
  done
}

retry() {
  local status=0
  local retries=${RETRY_MAX:=5}
  local delay=${INTERVAL:=5}
  local to=${TIMEOUT:=20}
  cmd="$*"

  while [ $retries -gt 0 ]
  do
    status=0
    timeout $to bash -c "echo $cmd && $cmd" || status=$?
    if [ $status -eq 0 ]; then
      break;
    fi
    echo "Exit code: '$status'. Sleeping '$delay' seconds before retrying"
    sleep $delay
    let retries--
  done
  return $status
}

echo "## checking requirements"
check_requirements
echo "## delete existing KinD cluster if it exists"
kind delete clusters $KIND_CLUSTER_NAME
echo "## start KinD cluster"
create_cluster
kind export kubeconfig --name $KIND_CLUSTER_NAME
echo "## wait for coreDNS"
kubectl -n kube-system wait --for=condition=available deploy/coredns --timeout=$TIMEOUT_K8
echo "## install multus"
retry kubectl create -f "${MULTUS_DAEMONSET_URL}"
retry kubectl -n kube-system wait --for=condition=ready -l name="multus" pod --timeout=$TIMEOUT_K8
echo "## install CNIs"
retry kubectl create -f "${CNIS_DAEMONSET_PATH}"
retry kubectl -n kube-system wait --for=condition=ready -l name="cni-plugins" pod --timeout=$TIMEOUT_K8
echo "## build whereabouts"
pushd "$ROOT"
$OCI_BIN build . -t "$IMG_NAME"
popd

echo "## load image into KinD"
trap "rm /tmp/whereabouts-img.tar || true" EXIT
"$OCI_BIN" save -o /tmp/whereabouts-img.tar "$IMG_NAME"
kind load image-archive --name "$KIND_CLUSTER_NAME" /tmp/whereabouts-img.tar

echo "## install whereabouts"
for file in "daemonset-install.yaml" "whereabouts.cni.cncf.io_ippools.yaml" "whereabouts.cni.cncf.io_overlappingrangeipreservations.yaml" "whereabouts.cni.cncf.io_nodeslicepools.yaml"; do
  # insert 'imagePullPolicy: Never' under the container 'image' so it is certain that the image used
  # by the daemonset is the one loaded into KinD and not one pulled from a repo
  sed '/        image:/a\        imagePullPolicy: Never' "$ROOT/doc/crds/$file" | retry kubectl apply -f -
done
# deployment has an extra tab for the sed so doing out of the loop
sed '/          image:/a\          imagePullPolicy: Never' "$ROOT/doc/crds/node-slice-controller.yaml" | retry kubectl apply -f -
retry kubectl wait -n kube-system --for=condition=ready -l app=whereabouts pod --timeout=$TIMEOUT_K8
retry kubectl wait -n kube-system --for=condition=ready -l app=whereabouts-controller pod --timeout=$TIMEOUT_K8
echo "## done"
