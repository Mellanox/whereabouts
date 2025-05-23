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

#An easy way to run this is to create a kind cluster using ./hack/e2e-setup-kind-cluster

HERE="$(dirname "$(readlink --canonicalize ${BASH_SOURCE[0]})")"
ROOT="$(readlink --canonicalize "$HERE/..")"
WHEREABOUTSNAD="$ROOT/yamls/whereaboutsScaleNAD.yaml"
SCALEDEPLOYMENT="$ROOT/yamls/scaleTestDeployment.yaml"

#create the whereabouts nad
oc apply -f "$WHEREABOUTSNAD"
#start a timer to record how long the pods take to spin up
start=$SECONDS
#create the deployment (change the replicas feild in the scale-deployment yaml if you want to test a different number of pods)
oc apply -f "$SCALEDEPLOYMENT"
kubectl rollout status deploy/scale-deployment
#wait for all pods to be deployed

#Log the amount of time it took the pods to create
createTime=$(( SECONDS - start ))
echo Pod creation duration:"$createTime"

#delete the deployment and track pod deletion timing
oc delete deploy/scale-deployment

