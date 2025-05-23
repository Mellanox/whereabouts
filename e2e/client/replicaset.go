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
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// WaitForReplicaSetSteadyState only plays nice with the replicaSet it's being used with.
// Any pods that might be up still from a previous test may cause unexpected results.
func WaitForReplicaSetSteadyState(ctx context.Context, cs *kubernetes.Clientset, namespace, label string, replicaSet *appsv1.ReplicaSet, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, isReplicaSetSteady(ctx, cs, replicaSet.Name, namespace, label))
}

// WaitForReplicaSetToDisappear polls up to timeout seconds for replicaset to be gone from the Kubernetes cluster.
// Returns an error if the replicaset is never deleted, or if GETing it returns an error other than `NotFound`.
func WaitForReplicaSetToDisappear(ctx context.Context, cs *kubernetes.Clientset, namespace, rsName string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, isReplicaSetGone(ctx, cs, rsName, namespace))
}

func isReplicaSetSteady(ctx context.Context, cs *kubernetes.Clientset, replicaSetName, namespace, label string) wait.ConditionWithContextFunc {
	return func(context.Context) (bool, error) {
		podList, err := ListPods(ctx, cs, namespace, label)
		if err != nil {
			return false, err
		}

		replicaSet, err := cs.AppsV1().ReplicaSets(namespace).Get(ctx, replicaSetName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if isReplicaSetSynchronized(replicaSet, podList) {
			return true, nil
		} else {
			return false, nil
		}
	}
}

// check two things:
//  1. number of pods that are ready should equal that of spec
//  2. number of pods matching replicaSet's selector should equal that of spec
//     (in 0 replicas case, replicas should finish terminating before this comes true)
func isReplicaSetSynchronized(replicaSet *appsv1.ReplicaSet, podList *corev1.PodList) bool {
	return replicaSet.Status.ReadyReplicas == (*replicaSet.Spec.Replicas) && int32(len(podList.Items)) == (*replicaSet.Spec.Replicas)
}

func isReplicaSetGone(ctx context.Context, cs *kubernetes.Clientset, rsName, namespace string) wait.ConditionWithContextFunc {
	return func(context.Context) (bool, error) {
		replicaSet, err := cs.AppsV1().ReplicaSets(namespace).Get(ctx, rsName, metav1.GetOptions{})
		if err != nil && k8serrors.IsNotFound(err) {
			return true, nil
		} else if err != nil {
			return false, fmt.Errorf("something weird happened with the replicaset, which is in state: [%s]. Errors: %w", replicaSet.Status.Conditions, err)
		}

		return false, nil
	}
}
