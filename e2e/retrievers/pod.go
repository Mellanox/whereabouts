// Copyright 2025 whereabouts authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package retrievers

import (
	"encoding/json"
	"fmt"

	core "k8s.io/api/core/v1"

	nettypes "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
)

func filterNetworkStatus(
	networkStatuses []nettypes.NetworkStatus, predicate func(nettypes.NetworkStatus) bool) *nettypes.NetworkStatus {
	for i, networkStatus := range networkStatuses {
		if predicate(networkStatus) {
			return &networkStatuses[i]
		}
	}
	return nil
}

func SecondaryIfaceIPValue(pod *core.Pod, ifName string) ([]string, error) {
	podNetStatus, found := pod.Annotations[nettypes.NetworkStatusAnnot]
	if !found {
		return nil, fmt.Errorf("the pod must feature the `networks-status` annotation")
	}

	var netStatus []nettypes.NetworkStatus
	if err := json.Unmarshal([]byte(podNetStatus), &netStatus); err != nil {
		return nil, err
	}

	secondaryInterfaceNetworkStatus := filterNetworkStatus(netStatus, func(status nettypes.NetworkStatus) bool {
		return status.Interface == ifName
	})

	if secondaryInterfaceNetworkStatus == nil {
		return nil, fmt.Errorf("the pod does not have the requested secondary interface")
	}

	if len(secondaryInterfaceNetworkStatus.IPs) == 0 {
		return nil, fmt.Errorf("the pod does not have IPs for its secondary interfaces")
	}

	return secondaryInterfaceNetworkStatus.IPs, nil
}
