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

package entities

import (
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testImage = "quay.io/dougbtv/alpine:latest"

func PodObject(podName string, namespace string, label, annotations map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: podMeta(podName, namespace, label, annotations),
		Spec:       podSpec("samplepod"),
	}
}

func podSpec(containerName string) corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    containerName,
				Command: containerCmd(),
				Image:   testImage,
			},
		},
	}
}

func StatefulSetSpec(statefulSetName string, namespace string, serviceName string, replicaNumber int, annotations map[string]string) *v1.StatefulSet {
	const labelKey = "app"

	replicas := int32(replicaNumber)
	webAppLabels := map[string]string{labelKey: serviceName}
	return &v1.StatefulSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Name: serviceName},
		Spec: v1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: webAppLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: podMeta(statefulSetName, namespace, webAppLabels, annotations),
				Spec:       podSpec(statefulSetName),
			},
			ServiceName:         serviceName,
			PodManagementPolicy: v1.ParallelPodManagement,
		},
	}
}

func ReplicaSetObject(replicaCount int32, rsName string, namespace string, label map[string]string, annotations map[string]string) *v1.ReplicaSet {
	numReplicas := &replicaCount

	const podName = "samplepod"
	return &v1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ReplicaSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      rsName,
			Namespace: namespace,
			Labels:    label,
		},
		Spec: v1.ReplicaSetSpec{
			Replicas: numReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: label,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      label,
					Annotations: annotations,
					Namespace:   namespace,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    podName,
							Command: containerCmd(),
							Image:   testImage,
						},
					},
				},
			},
		},
	}
}

func podMeta(podName string, namespace string, label map[string]string, annotations map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        podName,
		Namespace:   namespace,
		Labels:      label,
		Annotations: annotations,
	}
}

func containerCmd() []string {
	return []string{"/bin/ash", "-c", "trap : TERM INT; sleep infinity & wait"}
}
