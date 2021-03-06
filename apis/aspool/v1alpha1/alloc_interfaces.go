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
	"strings"

	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ AaList = &AllocList{}

// +k8s:deepcopy-gen=false
type AaList interface {
	client.ObjectList

	GetAllocs() []Aa
}

func (x *AllocList) GetAllocs() []Aa {
	allocs := make([]Aa, len(x.Items))
	for i, r := range x.Items {
		r := r // Pin range variable so we can take its address.
		allocs[i] = &r
	}
	return allocs
}

var _ Aa = &Alloc{}

// +k8s:deepcopy-gen=false
type Aa interface {
	resource.Object
	resource.Conditioned

	GetCondition(ct nddv1.ConditionKind) nddv1.Condition
	SetConditions(c ...nddv1.Condition)
	GetOrganizationName() string
	GetDeploymentName() string
	GetAsPoolName() string
	GetSourceTag() map[string]string
	GetSelector() map[string]string
	SetAs(as uint32)
	HasAs() (uint32, bool)

	SetOrganizationName(string)
	SetDeploymentName(string)
	SetAsPoolName(string)
}

// GetCondition of this Network Node.
func (x *Alloc) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ct)
}

// SetConditions of the Network Node.
func (x *Alloc) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

func (x *Alloc) GetOrganizationName() string {
	split := strings.Split(x.GetName(), ".")
	if len(split) >= 2 {
		return split[0]
	}
	return ""
}

func (x *Alloc) GetDeploymentName() string {
	split := strings.Split(x.GetName(), ".")
	if len(split) >= 4 {
		return split[1]
	}
	return ""
}

func (x *Alloc) GetAsPoolName() string {
	split := strings.Split(x.GetName(), ".")
	if len(split) == 3 {
		return split[1]
	}
	if len(split) >= 4 {
		return split[2]
	}
	return ""
}

func (n *Alloc) GetSourceTag() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(n.Spec.Alloc.SourceTag).IsZero() {
		return s
	}
	for _, tag := range n.Spec.Alloc.SourceTag {
		s[*tag.Key] = *tag.Value
	}
	return s
}

func (n *Alloc) GetSelector() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(n.Spec.Alloc.Selector).IsZero() {
		return s
	}
	for _, tag := range n.Spec.Alloc.Selector {
		s[*tag.Key] = *tag.Value
	}
	return s
}

func (n *Alloc) SetAs(as uint32) {
	n.Status = AllocStatus{
		Alloc: &NddrAsPoolAlloc{
			State: &NddrAllocState{
				As: &as,
			},
		},
	}
}

func (n *Alloc) HasAs() (uint32, bool) {
	if n.Status.Alloc != nil && n.Status.Alloc.State != nil && n.Status.Alloc.State.As != nil {
		return *n.Status.Alloc.State.As, true
	}
	return 0, false
}

func (x *Alloc) SetOrganizationName(s string) {
	x.Status.OrganizationName = &s
}

func (x *Alloc) SetDeploymentName(s string) {
	x.Status.DeploymentName = &s
}

func (x *Alloc) SetAsPoolName(s string) {
	x.Status.AsPoolName = &s
}
