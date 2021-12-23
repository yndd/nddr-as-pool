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

	GetPoolName() string
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
}

// GetCondition of this Network Node.
func (x *AsPool) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ct)
}

// SetConditions of the Network Node.
func (x *AsPool) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

func (n *AsPool) GetPoolName() string {
	if reflect.ValueOf(n.Spec.AspoolAsPool.Name).IsZero() {
		return ""
	}
	return *n.Spec.AspoolAsPool.Name
}

func (n *AsPool) GetAllocationStrategy() string {
	if reflect.ValueOf(n.Spec.AspoolAsPool.AllocationStrategy).IsZero() {
		return ""
	}
	return *n.Spec.AspoolAsPool.AllocationStrategy
}

func (n *AsPool) GetStart() uint32 {
	if reflect.ValueOf(n.Spec.AspoolAsPool.Start).IsZero() {
		return 0
	}
	return *n.Spec.AspoolAsPool.Start
}

func (n *AsPool) GetEnd() uint32 {
	if reflect.ValueOf(n.Spec.AspoolAsPool.End).IsZero() {
		return 0
	}
	return *n.Spec.AspoolAsPool.End
}

func (n *AsPool) GetAllocations() int {
	if n.Status.AspoolAsPool != nil && n.Status.AspoolAsPool.State != nil {
		return *n.Status.AspoolAsPool.State.Allocated
	}
	return 0
}

func (n *AsPool) GetAllocatedAses() []*uint32 {
	return n.Status.AspoolAsPool.State.Used
}

func (n *AsPool) InitializeResource() error {
	start := int(n.GetStart())
	end := int(n.GetEnd())

	//if start > end {
	// something went wrong with the input
	// maybe we can validate this a the input
	//}

	// check if the pool was already initialized
	if n.Status.AspoolAsPool != nil && n.Status.AspoolAsPool.State != nil {
		// pool was already initialiazed
		return nil
	}

	n.Status.AspoolAsPool = &NddrAsPoolAsPool{
		Name:        n.Spec.AspoolAsPool.Name,
		Start:       n.Spec.AspoolAsPool.Start,
		End:         n.Spec.AspoolAsPool.End,
		AdminState:  n.Spec.AspoolAsPool.AdminState,
		Description: n.Spec.AspoolAsPool.Description,
		State: &NddrAsPoolAsPoolState{
			Total:     utils.IntPtr(end - start),
			Allocated: utils.IntPtr(0),
			Available: utils.IntPtr(end - start),
			Used:      make([]*uint32, 0),
		},
	}
	return nil

}

func (n *AsPool) SetAs(as uint32) {
	*n.Status.AspoolAsPool.State.Allocated++
	*n.Status.AspoolAsPool.State.Available--

	n.Status.AspoolAsPool.State.Used = append(n.Status.AspoolAsPool.State.Used, &as)
}

func (n *AsPool) UnSetAs(as uint32) {
	*n.Status.AspoolAsPool.State.Allocated--
	*n.Status.AspoolAsPool.State.Available++

	found := false
	idx := 0
	for i, pas := range n.Status.AspoolAsPool.State.Used {
		if *pas == as {
			found = true
			idx = i
			break
		}
	}
	if found {
		n.Status.AspoolAsPool.State.Used = append(n.Status.AspoolAsPool.State.Used[:idx], n.Status.AspoolAsPool.State.Used[idx+1:]...)
	}
}

func (n *AsPool) UpdateAs(allocatedAses []*uint32) {
	*n.Status.AspoolAsPool.State.Available = *n.Status.AspoolAsPool.State.Total - len(allocatedAses)
	*n.Status.AspoolAsPool.State.Allocated = len(allocatedAses)
	n.Status.AspoolAsPool.State.Used = allocatedAses
}

func (n *AsPool) FindAs(as uint32) bool {
	for _, uas := range n.Status.AspoolAsPool.State.Used {
		if *uas == as {
			return true
		}
	}
	return false
}
