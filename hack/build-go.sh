#!/usr/bin/env bash
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


set -eu
cmd=whereabouts
eval $(go env | grep -e "GOHOSTOS" -e "GOHOSTARCH")
GO=${GO:-go}
GOOS=${GOOS:-${GOHOSTOS}}
GOARCH=${GOARCH:-${GOHOSTARCH}}
GOFLAGS=${GOFLAGS:-}
GLDFLAGS=${GLDFLAGS:-}

# version information
hasGit=true
git version > /dev/null 2>&1 || hasGit=false
GIT_SHA=""
GIT_TREE_STATE=""
GIT_TAG=""
GIT_LAST_TAG=""
RELEASE_STATUS="unreleased"
if $hasGit; then
    GIT_SHA=$(git rev-parse --short HEAD)
    # Tree state is "dirty" if there are uncommitted changes, untracked files are ignored
    GIT_TREE_STATE=$(test -n "`git status --porcelain --untracked-files=no`" && echo "dirty" || echo "clean")
    # Empty string if we are not building a tag
    GIT_TAG=$(git describe --tags --abbrev=0 --exact-match 2>/dev/null || true)
    # Find most recent tag
    GIT_TAG_LAST=$(git describe --tags --abbrev=0 2>/dev/null || true)
fi
# VERSION override mechanism if needed
VERSION=${VERSION:-}
if [[ -n "${VERSION}" || -n "${GIT_TAG}" ]]; then
    RELEASE_STATUS="released"
fi
if [[ -z "${VERSION}" ]]; then
    VERSION="${GIT_TAG_LAST}"
fi
echo "VERSION: ${VERSION}"
echo "GIT_SHA: ${GIT_SHA}"
echo "GIT_TREE_STATE: ${GIT_TREE_STATE}"
echo "RELEASE_STATUS: ${RELEASE_STATUS}"
VERSION_LDFLAGS="-X github.com/k8snetworkplumbingwg/whereabouts/pkg/version.Version=${VERSION} \
-X github.com/k8snetworkplumbingwg/whereabouts/pkg/version.GitSHA=${GIT_SHA} \
-X github.com/k8snetworkplumbingwg/whereabouts/pkg/version.GitTreeState=${GIT_TREE_STATE} \
-X github.com/k8snetworkplumbingwg/whereabouts/pkg/version.ReleaseStatus=${RELEASE_STATUS}"
GLDFLAGS="${GLDFLAGS} ${VERSION_LDFLAGS}"

CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} ${GO} build ${GOFLAGS} -ldflags "${GLDFLAGS}" -o bin/${cmd} ./cmd/
CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} ${GO} build ${GOFLAGS} -ldflags "${GLDFLAGS}" -o bin/ip-control-loop ./cmd/controlloop/
CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} ${GO} build ${GOFLAGS} -ldflags "${GLDFLAGS}" -o bin/node-slice-controller ./cmd/nodeslicecontroller/

