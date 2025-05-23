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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// OverlappingRangeIPReservationSpec defines the desired state of OverlappingRangeIPReservation
type OverlappingRangeIPReservationSpec struct {
	ContainerID string `json:"containerid,omitempty"`
	PodRef      string `json:"podref"`
	IfName      string `json:"ifname,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true

// OverlappingRangeIPReservation is the Schema for the OverlappingRangeIPReservations API
type OverlappingRangeIPReservation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec OverlappingRangeIPReservationSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// OverlappingRangeIPReservationList contains a list of OverlappingRangeIPReservation
type OverlappingRangeIPReservationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []OverlappingRangeIPReservation `json:"items"`
}
