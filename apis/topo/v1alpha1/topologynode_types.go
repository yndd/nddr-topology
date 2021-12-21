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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// TopologyNodeFinalizer is the name of the finalizer added to
	// TopologyNode to block delete operations until the physical node can be
	// deprovisioned.
	TopologyNodeFinalizer string = "node.topo.nddr.yndd.io"
)

// TopologyNode struct
type TopoTopologyNode struct {
	// +kubebuilder:validation:Enum=`disable`;`enable`
	// +kubebuilder:default:="enable"
	AdminState *string `json:"admin-state,omitempty"`
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Description *string                `json:"description,omitempty"`
	KindName    *string                `json:"kind-name,omitempty"`
	Name        *string                `json:"name,omitempty"`
	Tag         []*TopoTopologyNodeTag `json:"tag,omitempty"`
}

// TopologyNodeTag struct
type TopoTopologyNodeTag struct {
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Key *string `json:"key"`
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Value *string `json:"value,omitempty"`
}

// A TopologyNodeSpec defines the desired state of a TopologyNode.
type TopologyNodeSpec struct {
	//nddv1.ResourceSpec `json:",inline"`
	TopologyName     *string          `json:"topology-name"`
	TopoTopologyNode TopoTopologyNode `json:"node,omitempty"`
}

// A TopologyNodeStatus represents the observed state of a TopologyNode.
type TopologyNodeStatus struct {
	nddv1.ConditionedStatus `json:",inline"`
	TopoTopologyNode        *NddrTopologyTopologyNode `json:"node,omitempty"`
}

// +kubebuilder:object:root=true

// TopoTopologyNode is the Schema for the TopologyNode API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNC",type="string",JSONPath=".status.conditions[?(@.kind=='Synced')].status"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.conditions[?(@.kind=='Ready')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type TopologyNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopologyNodeSpec   `json:"spec,omitempty"`
	Status TopologyNodeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopoTopologyNodeList contains a list of TopologyNodes
type TopologyNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TopologyNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TopologyNode{}, &TopologyNodeList{})
}

// TopologyNode type metadata.
var (
	TopologyNodeKindKind         = reflect.TypeOf(TopoTopologyNode{}).Name()
	TopologyNodeGroupKind        = schema.GroupKind{Group: Group, Kind: TopologyNodeKindKind}.String()
	TopologyNodeKindAPIVersion   = TopologyNodeKindKind + "." + GroupVersion.String()
	TopologyNodeGroupVersionKind = GroupVersion.WithKind(TopologyNodeKindKind)
)
