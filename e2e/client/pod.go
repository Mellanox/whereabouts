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
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// WaitForPodReady polls up to timeout seconds for pod to enter steady state (running or succeeded state).
// Returns an error if the pod never enters a steady state.
func WaitForPodReady(ctx context.Context, cs *kubernetes.Clientset, namespace, podName string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, isPodRunning(ctx, cs, podName, namespace))
}

// WaitForPodToDisappear polls up to timeout seconds for pod to be gone from the Kubernetes cluster.
// Returns an error if the pod is never deleted, or if GETing it returns an error other than `NotFound`.
func WaitForPodToDisappear(ctx context.Context, cs *kubernetes.Clientset, namespace, podName string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, isPodGone(ctx, cs, podName, namespace))
}

// WaitForPodBySelector waits up to timeout seconds for all pods in 'namespace' with given 'selector' to enter provided state
// If no pods are found, return nil.
func WaitForPodBySelector(ctx context.Context, cs *kubernetes.Clientset, namespace, selector string, timeout time.Duration) error {
	podList, err := ListPods(ctx, cs, namespace, selector)
	if err != nil {
		return err
	}

	if len(podList.Items) == 0 {
		return nil
	}

	for _, pod := range podList.Items {
		if err := WaitForPodReady(ctx, cs, namespace, pod.Name, timeout); err != nil {
			return err
		}
	}
	return nil
}

// ListPods returns the list of currently scheduled or running pods in `namespace` with the given selector
func ListPods(ctx context.Context, cs *kubernetes.Clientset, namespace, selector string) (*corev1.PodList, error) {
	listOptions := metav1.ListOptions{LabelSelector: selector}
	podList, err := cs.CoreV1().Pods(namespace).List(ctx, listOptions)

	if err != nil {
		return nil, err
	}
	return podList, nil
}

func isPodRunning(ctx context.Context, cs *kubernetes.Clientset, podName, namespace string) wait.ConditionWithContextFunc {
	return func(context.Context) (bool, error) {
		pod, err := cs.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case corev1.PodRunning:
			return true, nil
		case corev1.PodFailed:
			return false, errors.New("pod failed")
		case corev1.PodSucceeded:
			return false, errors.New("pod succeeded")
		}

		return false, nil
	}
}

func isPodGone(ctx context.Context, cs *kubernetes.Clientset, podName, namespace string) wait.ConditionWithContextFunc {
	return func(context.Context) (bool, error) {
		pod, err := cs.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil && k8serrors.IsNotFound(err) {
			return true, nil
		} else if err != nil {
			return false, fmt.Errorf("something weird happened with the pod, which is in state: [%s]. Errors: %w", pod.Status.Phase, err)
		}

		return false, nil
	}
}
