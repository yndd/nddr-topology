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

package v1alpha1

import (
	"reflect"

	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/yndd/ndd-runtime/pkg/utils"
)

var _ Tl = &TopologyLink{}

// +k8s:deepcopy-gen=false
type Tl interface {
	resource.Object
	resource.Conditioned

	GetTopologyName() string
	GetAdminState() string
	GetDescription() string
	GetTags() map[string]string
	GetEndpoints() []*TopoTopologyLinkEndpoints
	GetEndpointANodeName() string
	GetEndpointBNodeName() string
	GetEndpointAInterfaceName() string
	GetEndpointBInterfaceName() string
	GetEndpointATag() map[string]string
	GetEndpointBTag() map[string]string
	GetLag() bool
	GetLagAName() string
	GetLagBName() string
	GetStatus() string
	GetNodes() []*NddrTopologyTopologyLinkStateNode
	InitializeResource() error
	SetStatus(string)
	SetReason(string)
	SetEndpoint(nodeName string, ep *NddrTopologyTopologyLinkStateNodeEndpoint)
}

// GetCondition of this Network Node.
func (x *TopologyLink) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ct)
}

// SetConditions of the Network Node.
func (x *TopologyLink) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

func (x *TopologyLink) GetTopologyName() string {
	if reflect.ValueOf(x.Spec.TopologyName).IsZero() {
		return ""
	}
	return *x.Spec.TopologyName
}

func (x *TopologyLink) GetAdminState() string {
	if reflect.ValueOf(x.Spec.TopoTopologyLink.AdminState).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyLink.AdminState
}

func (x *TopologyLink) GetDescription() string {
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Description).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyLink.Description
}

func (x *TopologyLink) GetTags() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Tag).IsZero() {
		return s
	}
	for _, tag := range x.Spec.TopoTopologyLink.Tag {
		s[*tag.Key] = *tag.Value
	}
	return s
}

func (x *TopologyLink) GetEndpoints() []*TopoTopologyLinkEndpoints {
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return nil
	}
	return x.Spec.TopoTopologyLink.Endpoints
}

func (x *TopologyLink) GetEndpointANodeName() string {
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyLink.Endpoints[0].NodeName
}

func (x *TopologyLink) GetEndpointAInterfaceName() string {
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyLink.Endpoints[0].InterfaceName
}

func (x *TopologyLink) GetEndpointATag() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return s
	}
	for _, tag := range x.Spec.TopoTopologyLink.Endpoints[0].Tag {
		s[*tag.Key] = *tag.Value
	}
	return s
}

func (x *TopologyLink) GetEndpointBNodeName() string {
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyLink.Endpoints[1].NodeName
}

func (x *TopologyLink) GetEndpointBInterfaceName() string {
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyLink.Endpoints[1].InterfaceName
}

func (x *TopologyLink) GetEndpointBTag() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return s
	}
	for _, tag := range x.Spec.TopoTopologyLink.Endpoints[1].Tag {
		s[*tag.Key] = *tag.Value
	}
	return s
}

func (x *TopologyLink) GetLag() bool {
	if _, ok := x.GetTags()[Lag]; ok {
		return true
	}
	return false
}

func (x *TopologyLink) GetLagAName() string {
	if n, ok := x.GetEndpointATag()[LagName]; ok {
		return n
	}
	return ""
}

func (x *TopologyLink) GetLagBName() string {
	if n, ok := x.GetEndpointBTag()[LagName]; ok {
		return n
	}
	return ""
}

func (x *TopologyLink) GetStatus() string {
	return *x.Status.TopoTopologyLink.State.Status
}

func (x *TopologyLink) GetNodes() []*NddrTopologyTopologyLinkStateNode {
	return x.Status.TopoTopologyLink.State.Node
}

func (x *TopologyLink) InitializeResource() error {
	eps := make([]*NddrTopologyTopologyLinkEndpoints, 0, len(x.Spec.TopoTopologyLink.Endpoints))
	for _, ep := range x.Spec.TopoTopologyLink.Endpoints {
		epTags := make([]*NddrTopologyTopologyLinkEndpointsTag, 0, len(ep.Tag))
		for _, tag := range ep.Tag {
			epTags = append(epTags, &NddrTopologyTopologyLinkEndpointsTag{
				Key:   tag.Key,
				Value: tag.Value,
			})
		}

		eps = append(eps, &NddrTopologyTopologyLinkEndpoints{
			InterfaceName: ep.InterfaceName,
			NodeName:      ep.NodeName,
			Tag:           epTags,
		})
	}

	tags := make([]*NddrTopologyTopologyLinkTag, 0, len(x.Spec.TopoTopologyLink.Tag))
	for _, tag := range x.Spec.TopoTopologyLink.Tag {
		tags = append(tags, &NddrTopologyTopologyLinkTag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	if x.Status.TopoTopologyLink != nil && x.Status.TopoTopologyLink.State != nil {
		x.Status.TopoTopologyLink.AdminState = x.Spec.TopoTopologyLink.AdminState
		x.Status.TopoTopologyLink.Description = x.Spec.TopoTopologyLink.Description
		x.Status.TopoTopologyLink.Endpoints = eps
		x.Status.TopoTopologyLink.Tag = tags
		return nil
	}

	x.Status.TopoTopologyLink = &NddrTopologyTopologyLink{
		Name:        x.Spec.TopoTopologyLink.Name,
		AdminState:  x.Spec.TopoTopologyLink.AdminState,
		Description: x.Spec.TopoTopologyLink.Description,
		Endpoints:   eps,
		Tag:         tags,
		State: &NddrTopologyTopologyLinkState{
			Status: utils.StringPtr(""),
			Reason: utils.StringPtr(""),
			Node:   make([]*NddrTopologyTopologyLinkStateNode, 0),
			Tag:    make([]*NddrTopologyTopologyLinkStateTag, 0),
		},
	}
	return nil
}

func (x *TopologyLink) SetStatus(s string) {
	x.Status.TopoTopologyLink.State.Status = &s
}

func (x *TopologyLink) SetReason(s string) {
	x.Status.TopoTopologyLink.State.Reason = &s
}

func (x *TopologyLink) SetEndpoint(nodeName string, ep *NddrTopologyTopologyLinkStateNodeEndpoint) {
	for _, node := range x.Status.TopoTopologyLink.State.Node {
		if *node.Name == nodeName {
			for _, nodeep := range node.Endpoint {
				if *nodeep.Name == *ep.Name {
					nodeep.Lag = ep.Lag
					nodeep.LagSubLink = ep.LagSubLink
					nodeep.Name = ep.Name
					return
				}
			}
			node.Endpoint = append(node.Endpoint, ep)
			return
		}
	}
	// if we come here we need to create the node
	x.Status.TopoTopologyLink.State.Node = append(x.Status.TopoTopologyLink.State.Node, &NddrTopologyTopologyLinkStateNode{
		Name: &nodeName,
		Endpoint: []*NddrTopologyTopologyLinkStateNodeEndpoint{
			ep,
		},
	})
}