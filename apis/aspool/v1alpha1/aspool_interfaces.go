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
	"github.com/yndd/ndd-runtime/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ ApList = &AsPoolList{}

// +k8s:deepcopy-gen=false
type ApList interface {
	client.ObjectList

	GetAsPools() []Ap
}

func (x *AsPoolList) GetAsPools() []Ap {
	xs := make([]Ap, len(x.Items))
	for i, r := range x.Items {
		r := r // Pin range variable so we can take its address.
		xs[i] = &r
	}
	return xs
}

var _ Ap = &AsPool{}

// +k8s:deepcopy-gen=false
type Ap interface {
	resource.Object
	resource.Conditioned

	GetOrganizationName() string
	GetDeploymentName() string
	GetAsPoolName() string
	GetAllocationStrategy() string
	GetStart() uint32
	GetEnd() uint32
	GetAllocations() int
	GetAllocatedAses() []*uint32
	InitializeResource() error
	SetAs(uint32)
	UnSetAs(uint32)
	UpdateAs([]*uint32)
	FindAs(uint32) bool

	SetOrganizationName(string)
	SetDeploymentName(string)
	SetAsPoolName(string)
}

// GetCondition of this Network Node.
func (x *AsPool) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ct)
}

// SetConditions of the Network Node.
func (x *AsPool) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

func (x *AsPool) GetOrganizationName() string {
	split := strings.Split(x.GetName(), ".")
	if len(split) >= 2 {
		return split[0]
	}
	return ""
}

func (x *AsPool) GetDeploymentName() string {
	split := strings.Split(x.GetName(), ".")
	if len(split) >= 3 {
		return split[1]
	}
	return ""
}

func (x *AsPool) GetAsPoolName() string {
	split := strings.Split(x.GetName(), ".")
	if len(split) == 2 {
		return split[1]
	}
	if len(split) >= 3 {
		return split[2]
	}
	return ""
}

func (x *AsPool) GetAllocationStrategy() string {
	if reflect.ValueOf(x.Spec.AspoolAsPool.AllocationStrategy).IsZero() {
		return ""
	}
	return *x.Spec.AspoolAsPool.AllocationStrategy
}

func (x *AsPool) GetStart() uint32 {
	if reflect.ValueOf(x.Spec.AspoolAsPool.Start).IsZero() {
		return 0
	}
	return *x.Spec.AspoolAsPool.Start
}

func (x *AsPool) GetEnd() uint32 {
	if reflect.ValueOf(x.Spec.AspoolAsPool.End).IsZero() {
		return 0
	}
	return *x.Spec.AspoolAsPool.End
}

func (x *AsPool) GetAllocations() int {
	if x.Status.AspoolAsPool != nil && x.Status.AspoolAsPool.State != nil {
		return *x.Status.AspoolAsPool.State.Allocated
	}
	return 0
}

func (x *AsPool) GetAllocatedAses() []*uint32 {
	return x.Status.AspoolAsPool.State.Used
}

func (x *AsPool) InitializeResource() error {
	start := int(x.GetStart())
	end := int(x.GetEnd())

	//if start > end {
	// something went wrong with the input
	// maybe we can validate this a the input
	//}

	// check if the pool was already initialized
	if x.Status.AspoolAsPool != nil && x.Status.AspoolAsPool.State != nil {
		// pool was already initialiazed
		return nil
	}

	x.Status.AspoolAsPool = &NddrAsPoolAsPool{
		Start:       x.Spec.AspoolAsPool.Start,
		End:         x.Spec.AspoolAsPool.End,
		AdminState:  x.Spec.AspoolAsPool.AdminState,
		Description: x.Spec.AspoolAsPool.Description,
		State: &NddrAsPoolAsPoolState{
			Total:     utils.IntPtr(end - start + 1),
			Allocated: utils.IntPtr(0),
			Available: utils.IntPtr(end - start + 1),
			Used:      make([]*uint32, 0),
		},
	}
	return nil

}

func (x *AsPool) SetAs(as uint32) {
	*x.Status.AspoolAsPool.State.Allocated++
	*x.Status.AspoolAsPool.State.Available--

	x.Status.AspoolAsPool.State.Used = append(x.Status.AspoolAsPool.State.Used, &as)
}

func (x *AsPool) UnSetAs(as uint32) {
	*x.Status.AspoolAsPool.State.Allocated--
	*x.Status.AspoolAsPool.State.Available++

	found := false
	idx := 0
	for i, pas := range x.Status.AspoolAsPool.State.Used {
		if *pas == as {
			found = true
			idx = i
			break
		}
	}
	if found {
		x.Status.AspoolAsPool.State.Used = append(x.Status.AspoolAsPool.State.Used[:idx], x.Status.AspoolAsPool.State.Used[idx+1:]...)
	}
}

func (x *AsPool) UpdateAs(allocatedAses []*uint32) {
	*x.Status.AspoolAsPool.State.Available = *x.Status.AspoolAsPool.State.Total - len(allocatedAses)
	*x.Status.AspoolAsPool.State.Allocated = len(allocatedAses)
	x.Status.AspoolAsPool.State.Used = allocatedAses
}

func (x *AsPool) FindAs(as uint32) bool {
	for _, uas := range x.Status.AspoolAsPool.State.Used {
		if *uas == as {
			return true
		}
	}
	return false
}

func (x *AsPool) SetOrganizationName(s string) {
	x.Status.OrganizationName = &s
}

func (x *AsPool) SetDeploymentName(s string) {
	x.Status.DeploymentName = &s
}

func (x *AsPool) SetAsPoolName(s string) {
	x.Status.AsPoolName = &s
}
