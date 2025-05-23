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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// WaitForStatefulSetGone ...
func WaitForStatefulSetGone(ctx context.Context, cs *kubernetes.Clientset, namespace, serviceName string, labelSelector string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, isStatefulSetGone(ctx, cs, serviceName, namespace, labelSelector))
}

func isStatefulSetGone(ctx context.Context, cs *kubernetes.Clientset, serviceName string, namespace string, labelSelector string) wait.ConditionWithContextFunc {
	return func(context.Context) (done bool, err error) {
		statefulSet, err := cs.AppsV1().StatefulSets(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return false, fmt.Errorf("something weird happened with the stateful set whose status is: [%s]. Errors: %w", statefulSet.Status.String(), err)
		}

		associatedPods, err := cs.CoreV1().Pods(namespace).List(ctx, selectViaLabels(labelSelector))
		if err != nil {
			return false, err
		}

		return isStatefulSetEmpty(statefulSet) && areAssociatedPodsGone(associatedPods), nil
	}
}

func selectViaLabels(labelSelector string) metav1.ListOptions {
	return metav1.ListOptions{LabelSelector: labelSelector}
}

func isStatefulSetEmpty(statefulSet *appsv1.StatefulSet) bool {
	return statefulSet.Status.CurrentReplicas == int32(0)
}

func areAssociatedPodsGone(pods *corev1.PodList) bool {
	return len(pods.Items) == 0
}

func WaitForStatefulSetCondition(ctx context.Context, cs *kubernetes.Clientset, namespace, serviceName string, expectedReplicas int, timeout time.Duration, predicate statefulSetPredicate) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, doesStatefulsetComplyWithCondition(ctx, cs, serviceName, namespace, expectedReplicas, predicate))
}

func doesStatefulsetComplyWithCondition(ctx context.Context, cs *kubernetes.Clientset, serviceName string, namespace string, expectedReplicas int, predicate statefulSetPredicate) wait.ConditionWithContextFunc {
	return func(context.Context) (bool, error) {
		statefulSet, err := cs.AppsV1().StatefulSets(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		return predicate(statefulSet, expectedReplicas), nil
	}
}

func IsStatefulSetReadyPredicate(statefulSet *appsv1.StatefulSet, expectedReplicas int) bool {
	return statefulSet.Status.ReadyReplicas == int32(expectedReplicas)
}

func IsStatefulSetDegradedPredicate(statefulSet *appsv1.StatefulSet, expectedReplicas int) bool {
	return statefulSet.Status.ReadyReplicas < int32(expectedReplicas)
}
