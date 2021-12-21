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
	// TopologyLinkFinalizer is the name of the finalizer added to
	// TopologyLink to block delete operations until the physical node can be
	// deprovisioned.
	TopologyLinkFinalizer string = "link.topo.nddr.yndd.io"
)

// TopoTopologyLink struct
type TopoTopologyLink struct {
	// +kubebuilder:validation:Enum=`disable`;`enable`
	// +kubebuilder:default:="enable"
	AdminState *string `json:"admin-state,omitempty"`
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Description *string                      `json:"description,omitempty"`
	Endpoints   []*TopoTopologyLinkEndpoints `json:"endpoints,omitempty"`
	Name        *string                      `json:"name,omitempty"`
	Tag         []*TopoTopologyLinkTag       `json:"tag,omitempty"`
}

// TopologyLinkEndpoints struct
type TopoTopologyLinkEndpoints struct {
	// kubebuilder:validation:MinLength=3
	// kubebuilder:validation:MaxLength=20
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`int-([1-9](\d){0,1}(/[abcd])?(/[1-9](\d){0,1})?/(([1-9](\d){0,1})|(1[0-1]\d)|(12[0-8])))|`
	InterfaceName *string                         `json:"interface-name"`
	NodeName      *string                         `json:"node-name"`
	Tag           []*TopoTopologyLinkEndpointsTag `json:"tag,omitempty"`
}

// TopologyLinkEndpointsTag struct
type TopoTopologyLinkEndpointsTag struct {
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

// TopologyLinkTag struct
type TopoTopologyLinkTag struct {
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

// A TopologyLinkSpec defines the desired state of a TopologyLink.
type TopologyLinkSpec struct {
	//nddv1.ResourceSpec `json:",inline"`
	TopologyName     *string           `json:"topology-name"`
	TopoTopologyLink *TopoTopologyLink `json:"link,omitempty"`
}

// A TopologyLinkStatus represents the observed state of a TopologyLink.
type TopologyLinkStatus struct {
	nddv1.ConditionedStatus `json:",inline"`
	TopoTopologyLink        *NddrTopologyTopologyLink `json:"link,omitempty"`
}

// +kubebuilder:object:root=true

// TopoTopologyLink is the Schema for the TopologyLink API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNC",type="string",JSONPath=".status.conditions[?(@.kind=='Synced')].status"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.conditions[?(@.kind=='Ready')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type TopologyLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopologyLinkSpec   `json:"spec,omitempty"`
	Status TopologyLinkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopoTopologyLinkList contains a list of TopologyLinks
type TopologyLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TopologyLink `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TopologyLink{}, &TopologyLinkList{})
}

// TopologyLink type metadata.
var (
	TopologyLinkKindKind         = reflect.TypeOf(TopologyLink{}).Name()
	TopologyLinkGroupKind        = schema.GroupKind{Group: Group, Kind: TopologyLinkKindKind}.String()
	TopologyLinkKindAPIVersion   = TopologyLinkKindKind + "." + GroupVersion.String()
	TopologyLinkGroupVersionKind = GroupVersion.WithKind(TopologyLinkKindKind)
)
