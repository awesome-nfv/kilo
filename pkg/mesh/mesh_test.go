// Copyright 2019 the Kilo authors
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

package mesh

import (
	"net"
	"testing"
)

func TestNewAllocator(t *testing.T) {
	_, c1, err := net.ParseCIDR("10.1.0.0/16")
	if err != nil {
		t.Fatalf("failed to parse CIDR: %v", err)
	}
	a1 := newAllocator(*c1)
	_, c2, err := net.ParseCIDR("10.1.0.0/32")
	if err != nil {
		t.Fatalf("failed to parse CIDR: %v", err)
	}
	a2 := newAllocator(*c2)
	_, c3, err := net.ParseCIDR("10.1.0.0/31")
	if err != nil {
		t.Fatalf("failed to parse CIDR: %v", err)
	}
	a3 := newAllocator(*c3)
	for _, tc := range []struct {
		name string
		a    *allocator
		next string
	}{
		{
			name: "10.1.0.0/16 first",
			a:    a1,
			next: "10.1.0.1/32",
		},
		{
			name: "10.1.0.0/16 second",
			a:    a1,
			next: "10.1.0.2/32",
		},
		{
			name: "10.1.0.0/32",
			a:    a2,
			next: "<nil>",
		},
		{
			name: "10.1.0.0/31 first",
			a:    a3,
			next: "10.1.0.1/32",
		},
		{
			name: "10.1.0.0/31 second",
			a:    a3,
			next: "<nil>",
		},
	} {
		next := tc.a.next()
		if next.String() != tc.next {
			t.Errorf("test case %q: expected %s, got %s", tc.name, tc.next, next.String())
		}
	}
}

func TestReady(t *testing.T) {
	internalIP := oneAddressCIDR(net.ParseIP("1.1.1.1"))
	externalIP := oneAddressCIDR(net.ParseIP("2.2.2.2"))
	for _, tc := range []struct {
		name  string
		node  *Node
		ready bool
	}{
		{
			name:  "nil",
			node:  nil,
			ready: false,
		},
		{
			name:  "empty fields",
			node:  &Node{},
			ready: false,
		},
		{
			name: "empty external IP",
			node: &Node{
				InternalIP: internalIP,
				Key:        []byte{},
				Subnet:     &net.IPNet{IP: net.ParseIP("10.2.0.0"), Mask: net.CIDRMask(16, 32)},
			},
			ready: false,
		},
		{
			name: "empty internal IP",
			node: &Node{
				ExternalIP: externalIP,
				Key:        []byte{},
				Subnet:     &net.IPNet{IP: net.ParseIP("10.2.0.0"), Mask: net.CIDRMask(16, 32)},
			},
			ready: false,
		},
		{
			name: "empty key",
			node: &Node{
				ExternalIP: externalIP,
				InternalIP: internalIP,
				Subnet:     &net.IPNet{IP: net.ParseIP("10.2.0.0"), Mask: net.CIDRMask(16, 32)},
			},
			ready: false,
		},
		{
			name: "empty subnet",
			node: &Node{
				ExternalIP: externalIP,
				InternalIP: internalIP,
				Key:        []byte{},
			},
			ready: false,
		},
		{
			name: "valid",
			node: &Node{
				ExternalIP: externalIP,
				InternalIP: internalIP,
				Key:        []byte{},
				Subnet:     &net.IPNet{IP: net.ParseIP("10.2.0.0"), Mask: net.CIDRMask(16, 32)},
			},
			ready: true,
		},
	} {
		ready := tc.node.Ready()
		if ready != tc.ready {
			t.Errorf("test case %q: expected %t, got %t", tc.name, tc.ready, ready)
		}
	}
}
