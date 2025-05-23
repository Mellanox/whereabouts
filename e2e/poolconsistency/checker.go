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

package poolconsistency

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/k8snetworkplumbingwg/whereabouts/e2e/retrievers"
	"github.com/k8snetworkplumbingwg/whereabouts/pkg/storage"
)

type Checker struct {
	ipPool  storage.IPPool
	podList []corev1.Pod
}

func NewPoolConsistencyCheck(ipPool storage.IPPool, podList []corev1.Pod) *Checker {
	return &Checker{
		ipPool:  ipPool,
		podList: podList,
	}
}

func (pc *Checker) MissingIPs() []string {
	var mismatchedIPs []string
	for _, pod := range pc.podList {
		podIPs, err := retrievers.SecondaryIfaceIPValue(&pod, "net1")
		podIP := podIPs[len(podIPs)-1]
		if err != nil {
			return []string{}
		}

		var found bool
		for _, allocation := range pc.ipPool.Allocations() {
			reservedIP := allocation.IP.String()

			if reservedIP == podIP {
				found = true
				break
			}
		}

		if !found {
			mismatchedIPs = append(mismatchedIPs, podIP)
		}
	}
	return mismatchedIPs
}

func (pc *Checker) StaleIPs() []string {
	var staleIPs []string
	for _, allocation := range pc.ipPool.Allocations() {
		reservedIP := allocation.IP.String()
		found := false
		for _, pod := range pc.podList {
			podIPs, err := retrievers.SecondaryIfaceIPValue(&pod, "net1")
			podIP := podIPs[len(podIPs)-1]
			if err != nil {
				continue
			}

			if reservedIP == podIP {
				found = true
				break
			}
		}

		if !found {
			staleIPs = append(staleIPs, allocation.IP.String())
		}
	}
	return staleIPs
}
