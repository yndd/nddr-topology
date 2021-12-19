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
	"fmt"
	"reflect"
	"strconv"

	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/yndd/ndd-runtime/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ TnList = &TopologyNodeList{}

// +k8s:deepcopy-gen=false
type TnList interface {
	client.ObjectList

	GetNodes() []Tn
}

func (x *TopologyNodeList) GetNodes() []Tn {
	xs := make([]Tn, len(x.Items))
	for i, r := range x.Items {
		r := r // Pin range variable so we can take its address.
		xs[i] = &r
	}
	return xs
}

var _ Tn = &TopologyNode{}

// +k8s:deepcopy-gen=false
type Tn interface {
	resource.Object
	resource.Conditioned

	GetTopologyName() string
	GetKindName() string
	GetAdminState() string
	GetDescription() string
	GetTags() map[string]string
	GetStateTags() map[string]string
	GetPlatform() string
	GetPosition() string
	GetNodeIndex() uint32
	GetStatus() string
	InitializeResource() error
	SetStatus(string)
	SetReason(string)
	SetPlatform(string)
	SetPosition(string)
	SetNodeEndpoint(ep *NddrTopologyTopologyLinkStateNodeEndpoint)
	GetNodeEndpoints() []*NddrTopologyTopologyNodeStateEndpoint
}

// GetCondition of this Network Node.
func (x *TopologyNode) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ct)
}

// SetConditions of the Network Node.
func (x *TopologyNode) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

func (x *TopologyNode) GetTopologyName() string {
	if reflect.ValueOf(x.Spec.TopologyName).IsZero() {
		return ""
	}
	return *x.Spec.TopologyName
}

func (x *TopologyNode) GetKindName() string {
	if reflect.ValueOf(x.Spec.TopoTopologyNode.KindName).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyNode.KindName
}

func (x *TopologyNode) GetAdminState() string {
	if reflect.ValueOf(x.Spec.TopoTopologyNode.AdminState).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyNode.AdminState
}

func (x *TopologyNode) GetDescription() string {
	if reflect.ValueOf(x.Spec.TopoTopologyNode.Description).IsZero() {
		return ""
	}
	return *x.Spec.TopoTopologyNode.Description
}

func (x *TopologyNode) GetTags() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(x.Spec.TopoTopologyNode.Tag).IsZero() {
		return s
	}
	for _, tag := range x.Spec.TopoTopologyNode.Tag {
		s[*tag.Key] = *tag.Value
	}
	return s
}

func (x *TopologyNode) GetStatus() string {
	if x.Status.TopoTopologyNode != nil && x.Status.TopoTopologyNode.State != nil && x.Status.TopoTopologyNode.State.Status != nil {
		return *x.Status.TopoTopologyNode.State.Status
	}
	return "unknown"
}

func (x *TopologyNode) GetStateTags() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(x.Status.TopoTopologyNode.State.Tag).IsZero() {
		return s
	}
	for _, tag := range x.Spec.TopoTopologyNode.Tag {
		s[*tag.Key] = *tag.Value
	}
	return s
}

func (x *TopologyNode) GetPlatform() string {
	if t, ok := x.GetStateTags()[KeyNodePlatform]; ok {
		return t
	}
	return ""
}

func (x *TopologyNode) GetPosition() string {
	fmt.Printf("getposition tags: %v\n", x.GetStateTags)
	if t, ok := x.GetStateTags()[KeyNodePosition]; ok {
		return t
	}
	return ""
}

func (x *TopologyNode) GetNodeIndex() uint32 {
	if t, ok := x.GetTags()[KeyNodeIndex]; ok {
		if i, err := strconv.Atoi(t); err == nil {
			return uint32(i)
		}
		return MaxUint32
	}
	return MaxUint32
}

func (x *TopologyNode) InitializeResource() error {
	tags := make([]*NddrTopologyTopologyNodeTag, 0, len(x.Spec.TopoTopologyNode.Tag))
	for _, tag := range x.Spec.TopoTopologyNode.Tag {
		tags = append(tags, &NddrTopologyTopologyNodeTag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	if x.Status.TopoTopologyNode != nil && x.Status.TopoTopologyNode.State != nil {
		// pool was already initialiazed
		x.Status.TopoTopologyNode.AdminState = x.Spec.TopoTopologyNode.AdminState
		x.Status.TopoTopologyNode.Description = x.Spec.TopoTopologyNode.Description
		x.Status.TopoTopologyNode.KindName = x.Spec.TopoTopologyNode.KindName
		x.Status.TopoTopologyNode.Tag = tags
		return nil
	}

	x.Status.TopoTopologyNode = &NddrTopologyTopologyNode{
		AdminState:  x.Spec.TopoTopologyNode.AdminState,
		Description: x.Spec.TopoTopologyNode.Description,
		KindName:    x.Spec.TopoTopologyNode.KindName,
		Tag:         tags,
		State: &NddrTopologyTopologyNodeState{
			Status:   utils.StringPtr(""),
			Reason:   utils.StringPtr(""),
			Endpoint: make([]*NddrTopologyTopologyNodeStateEndpoint, 0),
			Tag:      make([]*NddrTopologyTopologyNodeStateTag, 0),
		},
	}
	return nil
}

func (x *TopologyNode) SetStatus(s string) {
	x.Status.TopoTopologyNode.State.Status = &s
}

func (x *TopologyNode) SetReason(s string) {
	x.Status.TopoTopologyNode.State.Reason = &s
}

func (x *TopologyNode) SetPlatform(s string) {
	for _, tag := range x.Status.TopoTopologyNode.State.Tag {
		if *tag.Key == KeyNodePlatform {
			tag.Value = &s
			return
		}
	}
	x.Status.TopoTopologyNode.State.Tag = append(x.Status.TopoTopologyNode.State.Tag, &NddrTopologyTopologyNodeStateTag{
		Key:   utils.StringPtr(KeyNodePlatform),
		Value: &s,
	})
}

func (x *TopologyNode) SetPosition(s string) {
	for _, tag := range x.Status.TopoTopologyNode.State.Tag {
		if *tag.Key == KeyNodePosition {
			tag.Value = &s
			return
		}
	}
	x.Status.TopoTopologyNode.State.Tag = append(x.Status.TopoTopologyNode.State.Tag, &NddrTopologyTopologyNodeStateTag{
		Key:   utils.StringPtr(KeyNodePosition),
		Value: &s,
	})
}

func (x *TopologyNode) SetNodeEndpoint(ep *NddrTopologyTopologyLinkStateNodeEndpoint) {
	if x.Status.TopoTopologyNode.State.Endpoint == nil {
		x.Status.TopoTopologyNode.State.Endpoint = make([]*NddrTopologyTopologyNodeStateEndpoint, 0)
	}
	for _, nodeep := range x.Status.TopoTopologyNode.State.Endpoint {
		if *nodeep.Name == *ep.Name {
			// endpoint exists, so we update the information
			nodeep = &NddrTopologyTopologyNodeStateEndpoint{
				Name:       ep.Name,
				Lag:        ep.Lag,
				LagSubLink: ep.LagMemberLink,
			}
			return
		}
	}
	// new endpoint
	x.Status.TopoTopologyNode.State.Endpoint = append(x.Status.TopoTopologyNode.State.Endpoint,
		&NddrTopologyTopologyNodeStateEndpoint{
			Name:       ep.Name,
			Lag:        ep.Lag,
			LagSubLink: ep.LagMemberLink,
		})
}

func (x *TopologyNode) GetNodeEndpoints() []*NddrTopologyTopologyNodeStateEndpoint {
	if x.Status.TopoTopologyNode != nil && x.Status.TopoTopologyNode.State != nil && x.Status.TopoTopologyNode.State.Endpoint != nil {
		return x.Status.TopoTopologyNode.State.Endpoint
	}
	return make([]*NddrTopologyTopologyNodeStateEndpoint, 0)
}
