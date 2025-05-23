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

package storage

import (
	// "fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//	"time"
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cmd")
}

var _ = Describe("Storage operations", func() {
	// It("works with mutex", func() {
	// 	err := TestGetValue("127.0.0.1:2379")
	// 	Expect(err).NotTo(HaveOccurred())
	// })

	// It("gets a key", func() {

	// 	err := TestGetValue("127.0.0.1:2379")
	// 	Expect(err).NotTo(HaveOccurred())

	// 	// const ifname string = "eth0"
	// 	// const nspath string = "/some/where"

	// 	// ip, ipnet, err := net.ParseCIDR("192.168.2.200/24")
	// 	// Expect(err).NotTo(HaveOccurred())

	// 	// firstip, lastip, err := GetIPRange(ip, *ipnet)
	// 	// Expect(err).NotTo(HaveOccurred())

	// 	// Expect(fmt.Sprint(firstip)).To(Equal("192.168.2.0"))
	// 	// Expect(fmt.Sprint(lastip)).To(Equal("192.168.2.255"))

	// })
	// // For some reason this is going REALLY slow for me, can't prevent it.
	// It("errors on etcd not responding", func() {
	// 	SetTimeouts(500 * time.Millisecond)
	// 	err := TestGetValue("invalid.address:1234")
	// 	Expect(err).To(HaveOccurred())
	// })
})
