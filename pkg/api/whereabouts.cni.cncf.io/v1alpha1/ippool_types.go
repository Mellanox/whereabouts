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

package v1alpha1

import (
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IPPoolSpec defines the desired state of IPPool
type IPPoolSpec struct {
	// Range is a RFC 4632/4291-style string that represents an IP address and prefix length in CIDR notation
	Range string `json:"range"`
	// Allocations is the set of allocated IPs for the given range. Its` indices are a direct mapping to the
	// IP with the same index/offset for the pool's range.
	Allocations map[string]IPAllocation `json:"allocations"`
}

// ParseCIDR formats the Range of the IPPool
func (i IPPool) ParseCIDR() (net.IP, *net.IPNet, error) {
	return net.ParseCIDR(i.Spec.Range)
}

// IPAllocation represents metadata about the pod/container owner of a specific IP
type IPAllocation struct {
	ContainerID string `json:"id"`
	PodRef      string `json:"podref"`
	IfName      string `json:"ifname,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true

// IPPool is the Schema for the ippools API
type IPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IPPoolSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// IPPoolList contains a list of IPPool
type IPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPPool `json:"items"`
}
