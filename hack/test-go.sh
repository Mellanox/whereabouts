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

# single test: go test -v ./pkg/storage/
# without cache: go test -count=1 -v ./pkg/storage/
set -eox pipefail

GO=${GO:-go}
SKIP_STATIC_CHECK=$false
#parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
    -s|--skip-static-check)
      SKIP_STATIC_CHECK=true
      shift
      ;;
    *)
      echo "Invalid arguement: $1"
      echo "Usage: $0 [-s|--skip-static-check]"
      exit 1
  esac
done

echo "Running go vet ..."
${GO} vet --tags=test ./cmd/... ./pkg/...

BASEDIR=$(pwd)

if [ $SKIP_STATIC_CHECK ]
then
    echo "Skipped golang staticcheck"
else
  echo "Installing golang staticcheck ..."
  GOBIN=${BASEDIR}/bin go install honnef.co/go/tools/cmd/staticcheck@latest
  echo "Running golang staticcheck ..."
  ${BASEDIR}/bin/staticcheck --tags=test ./...
fi

echo "Running go tests..."
KUBEBUILDER_ASSETS="$(pwd)/bin" ${GO} test \
    --tags=test \
    -v \
    -covermode=count \
    -coverprofile=coverage.out \
    $(${GO} list ./... | grep -v e2e | tr "\n" " ")
