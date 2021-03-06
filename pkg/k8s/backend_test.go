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

package k8s

import (
	"net"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"k8s.io/api/core/v1"

	"github.com/squat/kilo/pkg/mesh"
)

func TestTranslateNode(t *testing.T) {
	for _, tc := range []struct {
		name        string
		annotations map[string]string
		labels      map[string]string
		out         *mesh.Node
		subnet      string
	}{
		{
			name:        "empty",
			annotations: nil,
			out:         &mesh.Node{},
		},
		{
			name: "invalid ip",
			annotations: map[string]string{
				externalIPAnnotationKey: "10.0.0.1",
				internalIPAnnotationKey: "10.0.0.1",
			},
			out: &mesh.Node{},
		},
		{
			name: "valid ip",
			annotations: map[string]string{
				externalIPAnnotationKey: "10.0.0.1/24",
				internalIPAnnotationKey: "10.0.0.2/32",
			},
			out: &mesh.Node{
				ExternalIP: &net.IPNet{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)},
				InternalIP: &net.IPNet{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(32, 32)},
			},
		},
		{
			name:        "invalid subnet",
			annotations: map[string]string{},
			out:         &mesh.Node{},
			subnet:      "foo",
		},
		{
			name:        "normalize subnet",
			annotations: map[string]string{},
			out: &mesh.Node{
				Subnet: &net.IPNet{IP: net.ParseIP("10.2.0.0"), Mask: net.CIDRMask(24, 32)},
			},
			subnet: "10.2.0.1/24",
		},
		{
			name:        "valid subnet",
			annotations: map[string]string{},
			out: &mesh.Node{
				Subnet: &net.IPNet{IP: net.ParseIP("10.2.1.0"), Mask: net.CIDRMask(24, 32)},
			},
			subnet: "10.2.1.0/24",
		},
		{
			name: "region",
			labels: map[string]string{
				regionLabelKey: "a",
			},
			out: &mesh.Node{
				Location: "a",
			},
		},
		{
			name: "region override",
			annotations: map[string]string{
				locationAnnotationKey: "b",
			},
			labels: map[string]string{
				regionLabelKey: "a",
			},
			out: &mesh.Node{
				Location: "b",
			},
		},
		{
			name: "external IP override",
			annotations: map[string]string{
				externalIPAnnotationKey:      "10.0.0.1/24",
				forceExternalIPAnnotationKey: "10.0.0.2/24",
			},
			out: &mesh.Node{
				ExternalIP: &net.IPNet{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)},
			},
		},
		{
			name: "complete",
			annotations: map[string]string{
				externalIPAnnotationKey:      "10.0.0.1/24",
				forceExternalIPAnnotationKey: "10.0.0.2/24",
				internalIPAnnotationKey:      "10.0.0.2/32",
				keyAnnotationKey:             "foo",
				leaderAnnotationKey:          "",
				locationAnnotationKey:        "b",
			},
			labels: map[string]string{
				regionLabelKey: "a",
			},
			out: &mesh.Node{
				ExternalIP: &net.IPNet{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)},
				InternalIP: &net.IPNet{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(32, 32)},
				Key:        []byte("foo"),
				Leader:     true,
				Location:   "b",
				Subnet:     &net.IPNet{IP: net.ParseIP("10.2.1.0"), Mask: net.CIDRMask(24, 32)},
			},
			subnet: "10.2.1.0/24",
		},
	} {
		n := &v1.Node{}
		n.ObjectMeta.Annotations = tc.annotations
		n.ObjectMeta.Labels = tc.labels
		n.Spec.PodCIDR = tc.subnet
		node := translateNode(n)
		if diff := pretty.Compare(node, tc.out); diff != "" {
			t.Errorf("test case %q: got diff: %v", tc.name, diff)
		}
	}
}
