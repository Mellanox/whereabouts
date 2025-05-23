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

package client

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func GetNodeSubnet(cs *ClientInfo, nodeName, sliceName, namespace string) (string, error) {
	slice, err := cs.WbClient.WhereaboutsV1alpha1().NodeSlicePools(namespace).Get(context.TODO(), sliceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	for _, allocation := range slice.Status.Allocations {
		if allocation.NodeName == nodeName {
			return allocation.SliceRange, nil
		}
	}
	return "", fmt.Errorf("slice range not found for node")
}

func WaitForNodeSliceReady(ctx context.Context, cs *ClientInfo, namespace, nodeSliceName string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, isNodeSliceReady(ctx, cs, namespace, nodeSliceName))
}

func isNodeSliceReady(ctx context.Context, cs *ClientInfo, namespace, nodeSliceName string) wait.ConditionWithContextFunc {
	return func(context.Context) (bool, error) {
		_, err := cs.WbClient.WhereaboutsV1alpha1().NodeSlicePools(namespace).Get(ctx, nodeSliceName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		return true, nil
	}
}
