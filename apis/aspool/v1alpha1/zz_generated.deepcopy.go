//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Alloc) DeepCopyInto(out *Alloc) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Alloc.
func (in *Alloc) DeepCopy() *Alloc {
	if in == nil {
		return nil
	}
	out := new(Alloc)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Alloc) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AllocList) DeepCopyInto(out *AllocList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Alloc, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AllocList.
func (in *AllocList) DeepCopy() *AllocList {
	if in == nil {
		return nil
	}
	out := new(AllocList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AllocList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AllocSpec) DeepCopyInto(out *AllocSpec) {
	*out = *in
	if in.AsPoolName != nil {
		in, out := &in.AsPoolName, &out.AsPoolName
		*out = new(string)
		**out = **in
	}
	if in.Alloc != nil {
		in, out := &in.Alloc, &out.Alloc
		*out = new(AspoolAlloc)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AllocSpec.
func (in *AllocSpec) DeepCopy() *AllocSpec {
	if in == nil {
		return nil
	}
	out := new(AllocSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AllocStatus) DeepCopyInto(out *AllocStatus) {
	*out = *in
	in.ConditionedStatus.DeepCopyInto(&out.ConditionedStatus)
	if in.Alloc != nil {
		in, out := &in.Alloc, &out.Alloc
		*out = new(NddrAsPoolAlloc)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AllocStatus.
func (in *AllocStatus) DeepCopy() *AllocStatus {
	if in == nil {
		return nil
	}
	out := new(AllocStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AsPool) DeepCopyInto(out *AsPool) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AsPool.
func (in *AsPool) DeepCopy() *AsPool {
	if in == nil {
		return nil
	}
	out := new(AsPool)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AsPool) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AsPoolList) DeepCopyInto(out *AsPoolList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AsPool, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AsPoolList.
func (in *AsPoolList) DeepCopy() *AsPoolList {
	if in == nil {
		return nil
	}
	out := new(AsPoolList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AsPoolList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AsPoolSpec) DeepCopyInto(out *AsPoolSpec) {
	*out = *in
	if in.AspoolAsPool != nil {
		in, out := &in.AspoolAsPool, &out.AspoolAsPool
		*out = new(AspoolAsPool)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AsPoolSpec.
func (in *AsPoolSpec) DeepCopy() *AsPoolSpec {
	if in == nil {
		return nil
	}
	out := new(AsPoolSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AsPoolStatus) DeepCopyInto(out *AsPoolStatus) {
	*out = *in
	in.ConditionedStatus.DeepCopyInto(&out.ConditionedStatus)
	if in.AspoolAsPool != nil {
		in, out := &in.AspoolAsPool, &out.AspoolAsPool
		*out = new(NddrAsPoolAsPool)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AsPoolStatus.
func (in *AsPoolStatus) DeepCopy() *AsPoolStatus {
	if in == nil {
		return nil
	}
	out := new(AsPoolStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AspoolAlloc) DeepCopyInto(out *AspoolAlloc) {
	*out = *in
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = make([]*AspoolAllocSelectorTag, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(AspoolAllocSelectorTag)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.SourceTag != nil {
		in, out := &in.SourceTag, &out.SourceTag
		*out = make([]*AspoolAllocSourceTagTag, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(AspoolAllocSourceTagTag)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AspoolAlloc.
func (in *AspoolAlloc) DeepCopy() *AspoolAlloc {
	if in == nil {
		return nil
	}
	out := new(AspoolAlloc)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AspoolAllocSelectorTag) DeepCopyInto(out *AspoolAllocSelectorTag) {
	*out = *in
	if in.Key != nil {
		in, out := &in.Key, &out.Key
		*out = new(string)
		**out = **in
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AspoolAllocSelectorTag.
func (in *AspoolAllocSelectorTag) DeepCopy() *AspoolAllocSelectorTag {
	if in == nil {
		return nil
	}
	out := new(AspoolAllocSelectorTag)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AspoolAllocSourceTagTag) DeepCopyInto(out *AspoolAllocSourceTagTag) {
	*out = *in
	if in.Key != nil {
		in, out := &in.Key, &out.Key
		*out = new(string)
		**out = **in
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AspoolAllocSourceTagTag.
func (in *AspoolAllocSourceTagTag) DeepCopy() *AspoolAllocSourceTagTag {
	if in == nil {
		return nil
	}
	out := new(AspoolAllocSourceTagTag)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AspoolAsPool) DeepCopyInto(out *AspoolAsPool) {
	*out = *in
	if in.AdminState != nil {
		in, out := &in.AdminState, &out.AdminState
		*out = new(string)
		**out = **in
	}
	if in.AllocationStrategy != nil {
		in, out := &in.AllocationStrategy, &out.AllocationStrategy
		*out = new(string)
		**out = **in
	}
	if in.End != nil {
		in, out := &in.End, &out.End
		*out = new(uint32)
		**out = **in
	}
	if in.Start != nil {
		in, out := &in.Start, &out.Start
		*out = new(uint32)
		**out = **in
	}
	if in.Description != nil {
		in, out := &in.Description, &out.Description
		*out = new(string)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AspoolAsPool.
func (in *AspoolAsPool) DeepCopy() *AspoolAsPool {
	if in == nil {
		return nil
	}
	out := new(AspoolAsPool)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NddrAllocState) DeepCopyInto(out *NddrAllocState) {
	*out = *in
	if in.As != nil {
		in, out := &in.As, &out.As
		*out = new(uint32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NddrAllocState.
func (in *NddrAllocState) DeepCopy() *NddrAllocState {
	if in == nil {
		return nil
	}
	out := new(NddrAllocState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NddrAsPool) DeepCopyInto(out *NddrAsPool) {
	*out = *in
	if in.AsPool != nil {
		in, out := &in.AsPool, &out.AsPool
		*out = make([]*NddrAsPoolAsPool, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(NddrAsPoolAsPool)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NddrAsPool.
func (in *NddrAsPool) DeepCopy() *NddrAsPool {
	if in == nil {
		return nil
	}
	out := new(NddrAsPool)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NddrAsPoolAlloc) DeepCopyInto(out *NddrAsPoolAlloc) {
	*out = *in
	in.AspoolAlloc.DeepCopyInto(&out.AspoolAlloc)
	if in.State != nil {
		in, out := &in.State, &out.State
		*out = new(NddrAllocState)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NddrAsPoolAlloc.
func (in *NddrAsPoolAlloc) DeepCopy() *NddrAsPoolAlloc {
	if in == nil {
		return nil
	}
	out := new(NddrAsPoolAlloc)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NddrAsPoolAsPool) DeepCopyInto(out *NddrAsPoolAsPool) {
	*out = *in
	if in.AdminState != nil {
		in, out := &in.AdminState, &out.AdminState
		*out = new(string)
		**out = **in
	}
	if in.AllocationStrategy != nil {
		in, out := &in.AllocationStrategy, &out.AllocationStrategy
		*out = new(string)
		**out = **in
	}
	if in.End != nil {
		in, out := &in.End, &out.End
		*out = new(uint32)
		**out = **in
	}
	if in.Start != nil {
		in, out := &in.Start, &out.Start
		*out = new(uint32)
		**out = **in
	}
	if in.Description != nil {
		in, out := &in.Description, &out.Description
		*out = new(string)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.State != nil {
		in, out := &in.State, &out.State
		*out = new(NddrAsPoolAsPoolState)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NddrAsPoolAsPool.
func (in *NddrAsPoolAsPool) DeepCopy() *NddrAsPoolAsPool {
	if in == nil {
		return nil
	}
	out := new(NddrAsPoolAsPool)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NddrAsPoolAsPoolState) DeepCopyInto(out *NddrAsPoolAsPoolState) {
	*out = *in
	if in.Total != nil {
		in, out := &in.Total, &out.Total
		*out = new(int)
		**out = **in
	}
	if in.Allocated != nil {
		in, out := &in.Allocated, &out.Allocated
		*out = new(int)
		**out = **in
	}
	if in.Available != nil {
		in, out := &in.Available, &out.Available
		*out = new(int)
		**out = **in
	}
	if in.Used != nil {
		in, out := &in.Used, &out.Used
		*out = make([]*uint32, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(uint32)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NddrAsPoolAsPoolState.
func (in *NddrAsPoolAsPoolState) DeepCopy() *NddrAsPoolAsPoolState {
	if in == nil {
		return nil
	}
	out := new(NddrAsPoolAsPoolState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Root) DeepCopyInto(out *Root) {
	*out = *in
	if in.AspoolNddrAsPool != nil {
		in, out := &in.AspoolNddrAsPool, &out.AspoolNddrAsPool
		*out = new(NddrAsPool)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Root.
func (in *Root) DeepCopy() *Root {
	if in == nil {
		return nil
	}
	out := new(Root)
	in.DeepCopyInto(out)
	return out
}