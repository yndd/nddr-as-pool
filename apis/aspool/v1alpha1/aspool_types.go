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
	// AsPoolFinalizer is the name of the finalizer added to
	// AsPool to block delete operations until the physical node can be
	// deprovisioned.
	AsPoolFinalizer string = "asPool.aspool.nddr.yndd.io"
)

// AsPool struct
type AspoolAsPool struct {
	// +kubebuilder:validation:Enum=`disable`;`enable`
	// +kubebuilder:default:="enable"
	AdminState *string `json:"admin-state,omitempty"`
	// +kubebuilder:validation:Enum=`first-available`;`deterministic`
	// +kubebuilder:default:="first-available"
	AllocationStrategy *string `json:"allocation-strategy,omitempty"`
	// kubebuilder:validation:Minimum=1
	// kubebuilder:validation:Maximum=4294967295
	End *uint32 `json:"end"`
	// kubebuilder:validation:Minimum=1
	// kubebuilder:validation:Maximum=4294967295
	Start *uint32 `json:"start"`
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Description *string `json:"description,omitempty"`
	Name        *string `json:"name"`
}

// A AsPoolSpec defines the desired state of a AsPool.
type AsPoolSpec struct {
	//nddv1.ResourceSpec `json:",inline"`
	AspoolAsPool *AspoolAsPool `json:"as-pool,omitempty"`
}

// A AsPoolStatus represents the observed state of a AsPool.
type AsPoolStatus struct {
	nddv1.ConditionedStatus `json:",inline"`
	AspoolAsPool            *NddrAsPoolAsPool `json:"as-pool,omitempty"`
}

// +kubebuilder:object:root=true

// AsPool is the Schema for the AsPool API
// +kubebuilder:subresource:status
type AsPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AsPoolSpec   `json:"spec,omitempty"`
	Status AsPoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AsPoolList contains a list of AsPools
type AsPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AsPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AsPool{}, &AsPoolList{})
}

// AsPool type metadata.
var (
	AsPoolKindKind         = reflect.TypeOf(AsPool{}).Name()
	AsPoolGroupKind        = schema.GroupKind{Group: Group, Kind: AsPoolKindKind}.String()
	AsPoolKindAPIVersion   = AsPoolKindKind + "." + GroupVersion.String()
	AsPoolGroupVersionKind = GroupVersion.WithKind(AsPoolKindKind)
)
