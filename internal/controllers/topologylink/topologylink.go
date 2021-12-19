/*
Copyright 2021 NDD.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package topologylink

import (

	//nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"

	"fmt"
	"strings"

	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/utils"
	topov1alpha1 "github.com/yndd/nddr-topology/apis/topo/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	shPrefix    = "logical-sh-link"
	mhPrefix    = "logical-mh-link"
	labelPrefix = "nddo-infra"
)

func buildLogicalTopologyLink(cr topov1alpha1.Tl) *topov1alpha1.TopologyLink {
	// the name of the logical link is set based on multi-homing or single homing
	// sh-lag:              <logical-sh-link>-<node-name-epA>-<lag-name-epA>-<node-name-epB><lag-name-epB>
	// mh-lag-A - sh-lag-B: <logical-mh-link>-<multihoming-A-name>-<node-name-epB><lag-name-epB>
	// sh-lag-A - mh-lag-B: <logical-mh-link>-<node-name-epB><lag-name-epB>-<multihoming-A-name>
	// mh-lag-A - mh-lag-B: <logical-mh-link>-<multihoming-A-name>-<multihoming-B-name>

	var (
		name           string
		nodeNameA      string
		nodeNameB      string
		interfaceNameA string
		interfaceNameB string
		epATags        []*topov1alpha1.TopoTopologyLinkEndpointsTag
		epBTags        []*topov1alpha1.TopoTopologyLinkEndpointsTag
	)
	fmt.Printf("buildLogicalTopologyLink: %s mhA: %t, mhB: %t\n", cr.GetName(), cr.GetEndPointAMultiHoming(), cr.GetEndPointBMultiHoming())
	if cr.GetEndPointAMultiHoming() || cr.GetEndPointBMultiHoming() {
		if cr.GetEndPointAMultiHoming() {
			// multihomed endpoint A
			name = strings.Join([]string{mhPrefix, cr.GetEndPointAMultiHomingName()}, "-")
			nodeNameA = ""
			interfaceNameA = cr.GetEndPointAMultiHomingName()
			epATags = cr.GetEndpointATagRaw()
			epATags = append(epATags, []*topov1alpha1.TopoTopologyLinkEndpointsTag{
				{Key: utils.StringPtr(topov1alpha1.KeyLinkEPMultiHoming), Value: utils.StringPtr("true")},
				{Key: utils.StringPtr(cr.GetEndpointANodeName()), Value: utils.StringPtr(cr.GetLagAName())},
			}...)
		} else {
			name = strings.Join([]string{mhPrefix, cr.GetEndpointANodeName()}, "-")
			nodeNameA = cr.GetEndpointANodeName()
			interfaceNameA = cr.GetLagAName()
			epATags = cr.GetEndpointATagRaw()
		}
		if cr.GetEndPointBMultiHoming() {
			// multihomed endpoint B
			name = strings.Join([]string{name, cr.GetEndPointBMultiHomingName()}, "-")
			nodeNameB = ""
			interfaceNameB = cr.GetEndPointAMultiHomingName()
			epBTags = cr.GetEndpointBTagRaw()
			epBTags = append(epBTags, []*topov1alpha1.TopoTopologyLinkEndpointsTag{
				{Key: utils.StringPtr(topov1alpha1.KeyLinkEPMultiHoming), Value: utils.StringPtr("true")},
				{Key: utils.StringPtr(cr.GetEndpointBNodeName()), Value: utils.StringPtr(cr.GetLagBName())},
			}...)
		} else {
			name = strings.Join([]string{name, cr.GetEndpointBNodeName()}, "-")
			nodeNameB = cr.GetEndpointBNodeName()
			interfaceNameB = cr.GetLagBName()
			epBTags = cr.GetEndpointBTagRaw()
		}
	} else {
		name = strings.Join([]string{shPrefix, cr.GetEndpointANodeName(), cr.GetLagAName(), cr.GetEndpointBNodeName(), cr.GetLagBName()}, "-")
		nodeNameA = cr.GetEndpointANodeName()
		interfaceNameA = cr.GetLagAName()
		epATags = cr.GetEndpointATagRaw()
		nodeNameB = cr.GetEndpointBNodeName()
		interfaceNameB = cr.GetLagBName()
		epBTags = cr.GetEndpointBTagRaw()
	}

	fmt.Printf("buildLogicalTopologyLink: name: %s nodeA: %s, nodeB: %s, itfcA: %s, itfceB: %s\n", name, nodeNameA, nodeNameB, interfaceNameA, interfaceNameB)
	fmt.Printf("buildLogicalTopologyLink: epAtags: %v, epBtags: %v\n", cr.GetEndpointATag(), cr.GetEndpointBTag())
	return &topov1alpha1.TopologyLink{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       cr.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(cr, topov1alpha1.TopologyLinkGroupVersionKind))},
		},
		Spec: topov1alpha1.TopologyLinkSpec{
			TopologyName: utils.StringPtr(cr.GetTopologyName()),
			TopoTopologyLink: &topov1alpha1.TopoTopologyLink{
				AdminState: utils.StringPtr("enable"),
				Endpoints: []*topov1alpha1.TopoTopologyLinkEndpoints{
					{
						NodeName:      utils.StringPtr(nodeNameA),
						InterfaceName: utils.StringPtr(interfaceNameA),
						Tag:           epATags,
					},
					{
						NodeName:      utils.StringPtr(nodeNameB),
						InterfaceName: utils.StringPtr(interfaceNameB),
						Tag:           epBTags,
					},
				},
				Tag: []*topov1alpha1.TopoTopologyLinkTag{
					{Key: utils.StringPtr("lag"), Value: utils.StringPtr("true")},
				},
			},
		},
		Status: topov1alpha1.TopologyLinkStatus{
			TopoTopologyLink: &topov1alpha1.NddrTopologyTopologyLink{
				Tag: cr.GetStatusTagsRaw(),
			},
		},
	}
}

func updateLogicalTopologyLink(cr topov1alpha1.Tl, mhtl *topov1alpha1.TopologyLink) *topov1alpha1.TopologyLink {
	if cr.GetEndPointAMultiHoming() && mhtl.GetEndPointAMultiHoming() && (cr.GetEndPointAMultiHomingName() == mhtl.GetEndPointAMultiHomingName()) {
		nodeNameA := cr.GetEndpointANodeName()
		interfaceNameA := cr.GetLagAName()

		fmt.Printf("updateLogicalTopologyLink: nodename: %s, itfcename: %s\n", nodeNameA, interfaceNameA)

		found := false
		for _, tag := range mhtl.GetEndpointATagRaw() {
			if *tag.Key == nodeNameA && *tag.Value == interfaceNameA {
				found = true
				break
			}
		}
		if !found {
			mhtl.AddEndPointATag(nodeNameA, interfaceNameA)
		}
	}
	if cr.GetEndPointBMultiHoming() && mhtl.GetEndPointBMultiHoming() && (cr.GetEndPointBMultiHomingName() == mhtl.GetEndPointBMultiHomingName()) {
		nodeNameB := cr.GetEndpointBNodeName()
		interfaceNameB := cr.GetLagBName()

		found := false
		for _, tag := range mhtl.GetEndpointBTagRaw() {
			if *tag.Key == nodeNameB && *tag.Value == interfaceNameB {
				found = true
				break
			}
		}
		if !found {
			mhtl.AddEndPointBTag(nodeNameB, interfaceNameB)
		}
	}
	return mhtl
}

func updateDeleteLogicalTopologyLink(cr topov1alpha1.Tl, mhtl *topov1alpha1.TopologyLink) *topov1alpha1.TopologyLink {
	if cr.GetEndPointAMultiHoming() && mhtl.GetEndPointAMultiHoming() && (cr.GetEndPointAMultiHomingName() == mhtl.GetEndPointAMultiHomingName()) {
		nodeNameA := cr.GetEndpointANodeName()
		interfaceNameA := cr.GetLagAName()

		fmt.Printf("updateLogicalTopologyLink: nodename: %s, itfcename: %s\n", nodeNameA, interfaceNameA)

		mhtl.DeleteEndPointATag(nodeNameA, interfaceNameA)
	}
	if cr.GetEndPointBMultiHoming() && mhtl.GetEndPointBMultiHoming() && (cr.GetEndPointBMultiHomingName() == mhtl.GetEndPointBMultiHomingName()) {
		nodeNameB := cr.GetEndpointBNodeName()
		interfaceNameB := cr.GetLagBName()

		mhtl.DeleteEndPointBTag(nodeNameB, interfaceNameB)
	}
	return mhtl
}
