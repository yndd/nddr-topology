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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ TlList = &TopologyLinkList{}

// +k8s:deepcopy-gen=false
type TlList interface {
	client.ObjectList

	GetLinks() []Tl
}

func (x *TopologyLinkList) GetLinks() []Tl {
	xs := make([]Tl, len(x.Items))
	for i, r := range x.Items {
		r := r // Pin range variable so we can take its address.
		xs[i] = &r
	}
	return xs
}

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
	GetEndpointATagRaw() []*TopoTopologyLinkEndpointsTag
	GetEndpointBTagRaw() []*TopoTopologyLinkEndpointsTag
	GetEndPointAKind() string
	GetEndPointBKind() string
	GetEndPointAGroup() string
	GetEndPointBGroup() string
	GetEndPointAMultiHoming() bool
	GetEndPointBMultiHoming() bool
	GetEndPointAMultiHomingName() string
	GetEndPointBMultiHomingName() string
	GetLagMember() bool
	GetLag() bool
	GetLacp() bool
	GetLacpFallbackA() bool
	GetLacpFallbackB() bool
	GetLagAName() string
	GetLagBName() string
	GetStatus() string
	GetNodes() []*NddrTopologyTopologyLinkStateNode
	GetStatusTagsRaw() []*NddrTopologyTopologyLinkTag
	InitializeResource() error
	SetStatus(string)
	SetReason(string)
	SetNodeEndpoint(nodeName string, ep *NddrTopologyTopologyLinkStateNodeEndpoint)
	GetNodeEndpoints() []*NddrTopologyTopologyLinkStateNode
	SetKind(s string)
	GetKind() string
	AddEndPointATag(string, string)
	AddEndPointBTag(string, string)
	DeleteEndPointATag(key string, value string)
	DeleteEndPointBTag(key string, value string)
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

func (x *TopologyLink) GetStatusTagsRaw() []*NddrTopologyTopologyLinkTag {
	if x.Status.TopoTopologyLink != nil && x.Status.TopoTopologyLink.Tag != nil {
		return x.Status.TopoTopologyLink.Tag
	}
	return make([]*NddrTopologyTopologyLinkTag, 0)
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

func (x *TopologyLink) GetEndpointBNodeName() string {
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyLink.Endpoints[1].NodeName
}

func (x *TopologyLink) GetEndpointAInterfaceName() string {
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyLink.Endpoints[0].InterfaceName
}

func (x *TopologyLink) GetEndpointBInterfaceName() string {
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyLink.Endpoints[1].InterfaceName
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

func (x *TopologyLink) GetEndpointATagRaw() []*TopoTopologyLinkEndpointsTag {
	s := make([]*TopoTopologyLinkEndpointsTag, 0)
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return s
	}
	return x.Spec.TopoTopologyLink.Endpoints[0].Tag
}

func (x *TopologyLink) GetEndpointBTagRaw() []*TopoTopologyLinkEndpointsTag {
	s := make([]*TopoTopologyLinkEndpointsTag, 0)
	if reflect.ValueOf(x.Spec.TopoTopologyLink.Endpoints).IsZero() {
		return s
	}
	return x.Spec.TopoTopologyLink.Endpoints[1].Tag
}

func (x *TopologyLink) GetEndPointAKind() string {
	if n, ok := x.GetEndpointATag()[KeyLinkEPKind]; ok {
		return n
	}
	// default
	return "infra"
}

func (x *TopologyLink) GetEndPointBKind() string {
	if n, ok := x.GetEndpointBTag()[KeyLinkEPKind]; ok {
		return n
	}
	// default
	return "infra"
}

func (x *TopologyLink) GetEndPointAGroup() string {
	if n, ok := x.GetEndpointATag()[KeyLinkEPGroup]; ok {
		return n
	}
	// default
	return ""
}

func (x *TopologyLink) GetEndPointBGroup() string {
	if n, ok := x.GetEndpointBTag()[KeyLinkEPGroup]; ok {
		return n
	}
	// default
	return ""
}

func (x *TopologyLink) GetEndPointAMultiHoming() bool {
	if n, ok := x.GetEndpointATag()[KeyLinkEPMultiHoming]; ok {
		return n == "true"
	}
	// default
	return false
}

func (x *TopologyLink) GetEndPointBMultiHoming() bool {
	if n, ok := x.GetEndpointBTag()[KeyLinkEPMultiHoming]; ok {
		return n == "true"
	}
	// default
	return false
}

func (x *TopologyLink) GetEndPointAMultiHomingName() string {
	if n, ok := x.GetEndpointATag()[KeyLinkEPMultiHomingName]; ok {
		return n
	}
	// default
	return ""
}

func (x *TopologyLink) GetEndPointBMultiHomingName() string {
	if n, ok := x.GetEndpointBTag()[KeyLinkEPMultiHomingName]; ok {
		return n
	}
	// default
	return ""
}

func (x *TopologyLink) GetLacpFallbackA() bool {
	if _, ok := x.GetEndpointATag()[KeyLinkEPLacpFallback]; ok {
		return x.GetTags()[KeyLinkEPLacpFallback] == "true"
	}
	// default
	return false
}

func (x *TopologyLink) GetLacpFallbackB() bool {
	if _, ok := x.GetEndpointBTag()[KeyLinkEPLacpFallback]; ok {
		return x.GetTags()[KeyLinkEPLacpFallback] == "true"
	}
	// default
	return false
}

func (x *TopologyLink) GetLagMember() bool {
	if _, ok := x.GetTags()[KeyLinkLagMember]; ok {
		return x.GetTags()[KeyLinkLagMember] == "true"
	}
	// default is false
	return false
}

func (x *TopologyLink) GetLag() bool {
	if _, ok := x.GetTags()[KeyLinkLag]; ok {
		return x.GetTags()[KeyLinkLag] == "true"
	}
	// default is false
	return false
}

func (x *TopologyLink) GetLacp() bool {
	if _, ok := x.GetTags()[KeyLinkLacp]; ok {
		return x.GetTags()[KeyLinkLacp] == "true"
	}
	// default is true
	return true
}

func (x *TopologyLink) GetLagAName() string {
	if n, ok := x.GetEndpointATag()[KeyLinkEPLagName]; ok {
		return n
	}
	return ""
}

func (x *TopologyLink) GetLagBName() string {
	if n, ok := x.GetEndpointBTag()[KeyLinkEPLagName]; ok {
		return n
	}
	return ""
}

func (x *TopologyLink) GetStatus() string {
	if x.Status.TopoTopologyLink != nil && x.Status.TopoTopologyLink.State != nil && x.Status.TopoTopologyLink.State.Status != nil {
		return *x.Status.TopoTopologyLink.State.Status
	}
	return "unknown"
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

func (x *TopologyLink) SetNodeEndpoint(nodeName string, ep *NddrTopologyTopologyLinkStateNodeEndpoint) {
	for _, node := range x.Status.TopoTopologyLink.State.Node {
		if *node.Name == nodeName {
			for _, nodeep := range node.Endpoint {
				if *nodeep.Name == *ep.Name {
					nodeep.Lag = ep.Lag
					nodeep.LagMemberLink = ep.LagMemberLink
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

func (x *TopologyLink) GetNodeEndpoints() []*NddrTopologyTopologyLinkStateNode {
	if x.Status.TopoTopologyLink != nil && x.Status.TopoTopologyLink.State != nil && x.Status.TopoTopologyLink.State.Node != nil {
		return x.Status.TopoTopologyLink.State.Node
	}
	return make([]*NddrTopologyTopologyLinkStateNode, 0)
}

func (x *TopologyLink) SetKind(s string) {
	for _, tag := range x.Status.TopoTopologyLink.State.Tag {
		if *tag.Key == KeyLinkKind {
			tag.Value = &s
			return
		}
	}
	x.Status.TopoTopologyLink.State.Tag = append(x.Status.TopoTopologyLink.State.Tag, &NddrTopologyTopologyLinkStateTag{
		Key:   utils.StringPtr(KeyLinkKind),
		Value: &s,
	})
}

func (x *TopologyLink) GetKind() string {
	if x.Status.TopoTopologyLink != nil && x.Status.TopoTopologyLink.State != nil && x.Status.TopoTopologyLink.State.Tag != nil {
		for _, tag := range x.Status.TopoTopologyLink.State.Tag {
			if *tag.Key == KeyLinkKind {
				return *tag.Value
			}
		}

	}
	return LinkEPKindUnknown.String()
}

func (x *TopologyLink) AddEndPointATag(key string, value string) {
	for _, tag := range x.Spec.TopoTopologyLink.Endpoints[0].Tag {
		if *tag.Key == key {
			tag.Value = &value
			return
		}
	}
	// if not found append
	x.Spec.TopoTopologyLink.Endpoints[0].Tag = append(x.Spec.TopoTopologyLink.Endpoints[0].Tag,
		&TopoTopologyLinkEndpointsTag{
			Key:   &key,
			Value: &value,
		})
}

func (x *TopologyLink) AddEndPointBTag(key string, value string) {
	for _, tag := range x.Spec.TopoTopologyLink.Endpoints[1].Tag {
		if *tag.Key == key {
			tag.Value = &value
			return
		}
	}
	// if not found append
	x.Spec.TopoTopologyLink.Endpoints[1].Tag = append(x.Spec.TopoTopologyLink.Endpoints[1].Tag,
		&TopoTopologyLinkEndpointsTag{
			Key:   &key,
			Value: &value,
		})
}

func (x *TopologyLink) DeleteEndPointATag(key string, value string) {
	found := false
	var idx int
	for i, tag := range x.Spec.TopoTopologyLink.Endpoints[0].Tag {
		if *tag.Key == key && *tag.Value == value {
			idx = i
			found = true
		}
	}
	if found {
		x.Spec.TopoTopologyLink.Endpoints[0].Tag = append(x.Spec.TopoTopologyLink.Endpoints[0].Tag[:idx], x.Spec.TopoTopologyLink.Endpoints[0].Tag[idx+1:]...)
	}
}

func (x *TopologyLink) DeleteEndPointBTag(key string, value string) {
	found := false
	var idx int
	for i, tag := range x.Spec.TopoTopologyLink.Endpoints[1].Tag {
		if *tag.Key == key && *tag.Value == value {
			idx = i
			found = true
		}
	}
	if found {
		x.Spec.TopoTopologyLink.Endpoints[1].Tag = append(x.Spec.TopoTopologyLink.Endpoints[1].Tag[:idx], x.Spec.TopoTopologyLink.Endpoints[1].Tag[idx+1:]...)
	}
}
