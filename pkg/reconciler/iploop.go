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

package reconciler

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"

	whereaboutsv1alpha1 "github.com/k8snetworkplumbingwg/whereabouts/pkg/api/whereabouts.cni.cncf.io/v1alpha1"
	"github.com/k8snetworkplumbingwg/whereabouts/pkg/logging"
	"github.com/k8snetworkplumbingwg/whereabouts/pkg/storage"
	"github.com/k8snetworkplumbingwg/whereabouts/pkg/storage/kubernetes"
	"github.com/k8snetworkplumbingwg/whereabouts/pkg/types"
)

type ReconcileLooper struct {
	k8sClient              kubernetes.Client
	liveWhereaboutsPods    map[string]podWrapper
	orphanedIPs            []OrphanedIPReservations
	orphanedClusterWideIPs []whereaboutsv1alpha1.OverlappingRangeIPReservation
}

type OrphanedIPReservations struct {
	Pool        storage.IPPool
	Allocations []types.IPReservation
}

func NewReconcileLooper() (*ReconcileLooper, error) {
	logging.Debugf("NewReconcileLooper - inferred connection data")
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return nil, logging.Errorf("failed to instantiate the Kubernetes client: %+v", err)
	}
	return NewReconcileLooperWithClient(k8sClient)
}

func NewReconcileLooperWithClient(k8sClient *kubernetes.Client) (*ReconcileLooper, error) {
	ipPools, err := k8sClient.ListIPPools()
	if err != nil {
		return nil, logging.Errorf("failed to retrieve all IP pools: %v", err)
	}

	pods, err := k8sClient.ListPods()
	if err != nil {
		return nil, err
	}

	whereaboutsPodRefs := getPodRefsServedByWhereabouts(ipPools)
	looper := &ReconcileLooper{
		k8sClient:           *k8sClient,
		liveWhereaboutsPods: indexPods(pods, whereaboutsPodRefs),
	}

	if err := looper.findOrphanedIPsPerPool(ipPools); err != nil {
		return nil, err
	}

	if err := looper.findClusterWideIPReservations(); err != nil {
		return nil, err
	}
	return looper, nil
}

func (rl *ReconcileLooper) findOrphanedIPsPerPool(ipPools []storage.IPPool) error {
	for _, pool := range ipPools {
		orphanIP := OrphanedIPReservations{
			Pool: pool,
		}
		for _, ipReservation := range pool.Allocations() {
			logging.Debugf("the IP reservation: %s", ipReservation)
			if ipReservation.PodRef == "" {
				_ = logging.Errorf("pod ref missing for Allocations: %s", ipReservation)
				continue
			}
			if !rl.isOrphanedIP(ipReservation.PodRef, ipReservation.IP.String()) {
				logging.Debugf("pod ref %s is not listed in the live pods list", ipReservation.PodRef)
				orphanIP.Allocations = append(orphanIP.Allocations, ipReservation)
			}
		}
		if len(orphanIP.Allocations) > 0 {
			rl.orphanedIPs = append(rl.orphanedIPs, orphanIP)
		}
	}

	return nil
}

func (rl ReconcileLooper) isOrphanedIP(podRef string, ip string) bool {
	for livePodRef, livePod := range rl.liveWhereaboutsPods {
		if podRef == livePodRef {
			isFound := isIpOnPod(&livePod, podRef, ip)
			if !isFound && (livePod.phase == v1.PodPending) {
				/* Sometimes pods are still coming up, and may not yet have Multus
				 * annotation added to it yet. We don't want to check the IPs yet
				 * so re-fetch the Pod 5x
				 */
				podToMatch := &livePod
				retries := 0

				logging.Debugf("Re-fetching Pending Pod: %s IP-to-match: %s", livePodRef, ip)

				for retries < storage.PodRefreshRetries {
					retries += 1
					podToMatch = rl.refreshPod(livePodRef)
					if podToMatch == nil {
						logging.Debugf("Cleaning up...")
						return false
					} else if podToMatch.phase != v1.PodPending {
						logging.Debugf("Pending Pod is now in phase: %s", podToMatch.phase)
						break
					} else {
						isFound = isIpOnPod(podToMatch, podRef, ip)
						// Short-circuit - Pending Pod may have IP now
						if isFound {
							logging.Debugf("Pod now has IP annotation while in Pending")
							return true
						}
						time.Sleep(time.Duration(250) * time.Millisecond)
					}
				}
				isFound = isIpOnPod(podToMatch, podRef, ip)
			}

			return isFound
		}
	}
	return false
}

func (rl ReconcileLooper) refreshPod(podRef string) *podWrapper {
	namespace, podName := splitPodRef(podRef)
	if namespace == "" || podName == "" {
		logging.Errorf("Invalid podRef format: %s", podRef)
		return nil
	}

	pod, err := rl.k8sClient.GetPod(namespace, podName)
	if err != nil {
		logging.Errorf("Failed to refresh Pod %s: %s\n", podRef, err)
		return nil
	}

	wrappedPod := wrapPod(*pod)
	logging.Debugf("Got refreshed pod: %v", wrappedPod)
	return wrappedPod
}

func splitPodRef(podRef string) (string, string) {
	namespacedName := strings.Split(podRef, "/")
	if len(namespacedName) != 2 {
		logging.Errorf("Failed to split podRef %s", podRef)
		return "", ""
	}

	return namespacedName[0], namespacedName[1]
}

func composePodRef(pod v1.Pod) string {
	return fmt.Sprintf("%s/%s", pod.GetNamespace(), pod.GetName())
}

func (rl ReconcileLooper) ReconcileIPPools() ([]net.IP, error) {
	findAllocationIndex := func(reservation types.IPReservation, reservations []types.IPReservation) int {
		for idx, r := range reservations {
			if r.PodRef == reservation.PodRef && r.IP.Equal(reservation.IP) {
				return idx
			}
		}
		return -1
	}

	var totalCleanedUpIps []net.IP
	for _, orphanedIP := range rl.orphanedIPs {
		currentIPReservations := orphanedIP.Pool.Allocations()

		// Process orphaned allocation peer pool
		var cleanedUpIpsPerPool []net.IP
		for _, allocation := range orphanedIP.Allocations {
			idx := findAllocationIndex(allocation, currentIPReservations)
			if idx < 0 {
				// Should never happen
				logging.Debugf("Failed to find allocation for pod ref: %s and IP: %s", allocation.PodRef, allocation.IP.String())
				continue
			}

			// Delete entry
			currentIPReservations[idx] = currentIPReservations[len(currentIPReservations)-1]
			currentIPReservations = currentIPReservations[:len(currentIPReservations)-1]

			cleanedUpIpsPerPool = append(cleanedUpIpsPerPool, allocation.IP)
		}

		if len(cleanedUpIpsPerPool) != 0 {
			logging.Debugf("Going to update the reserve list to: %+v", currentIPReservations)

			ctx, cancel := context.WithTimeout(context.Background(), storage.RequestTimeout)
			if err := orphanedIP.Pool.Update(ctx, currentIPReservations); err != nil {
				cancel()
				return nil, logging.Errorf("failed to update the reservation list: %v", err)
			}

			cancel()
			totalCleanedUpIps = append(totalCleanedUpIps, cleanedUpIpsPerPool...)
		}
	}

	return totalCleanedUpIps, nil
}

func (rl *ReconcileLooper) findClusterWideIPReservations() error {
	clusterWideIPReservations, err := rl.k8sClient.ListOverlappingIPs()
	if err != nil {
		return logging.Errorf("failed to list all OverLappingIPs: %v", err)
	}

	for _, clusterWideIPReservation := range clusterWideIPReservations {
		ip := clusterWideIPReservation.GetName()
		// De-normalize the IP
		// In the UpdateOverlappingRangeAllocation function, the IP address is created with a "normalized" name to comply with the k8s api.
		// We must denormalize here in order to properly look up the IP address in the regular format, which pods use.
		denormalizedip := strings.ReplaceAll(ip, "-", ":")

		podRef := clusterWideIPReservation.Spec.PodRef

		if !rl.isOrphanedIP(podRef, denormalizedip) {
			logging.Debugf("pod ref %s is not listed in the live pods list", podRef)
			rl.orphanedClusterWideIPs = append(rl.orphanedClusterWideIPs, clusterWideIPReservation)
		}
	}

	return nil
}

func (rl ReconcileLooper) ReconcileOverlappingIPAddresses() error {
	var failedReconciledClusterWideIPs []string

	for _, overlappingIPStruct := range rl.orphanedClusterWideIPs {
		if err := rl.k8sClient.DeleteOverlappingIP(&overlappingIPStruct); err != nil {
			logging.Errorf("failed to remove cluster wide IP: %s", overlappingIPStruct.GetName())
			failedReconciledClusterWideIPs = append(failedReconciledClusterWideIPs, overlappingIPStruct.GetName())
			continue
		}
		logging.Verbosef("removed stale overlappingIP allocation [%s]", overlappingIPStruct.GetName())
	}

	if len(failedReconciledClusterWideIPs) != 0 {
		return logging.Errorf("could not reconcile cluster wide IPs: %v", failedReconciledClusterWideIPs)
	}
	return nil
}
