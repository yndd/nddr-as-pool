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

// NddrAsPool struct
type NddrAsPool struct {
	AsPool []*NddrAsPoolAsPool `json:"as-pool,omitempty"`
}

// NddrAsPoolAsPool struct
type NddrAsPoolAsPool struct {
	AdminState         *string                `json:"admin-state,omitempty"`
	AllocationStrategy *string                `json:"allocation-strategy,omitempty"`
	End                *uint32                `json:"as-end,omitempty"`
	Start              *uint32                `json:"as-start,omitempty"`
	Description        *string                `json:"description,omitempty"`
	Name               *string                `json:"name,omitempty"`
	State              *NddrAsPoolAsPoolState `json:"state,omitempty"`
}

// NddrAsPoolAsPoolState struct
type NddrAsPoolAsPoolState struct {
	Total     *int      `json:"total,omitempty"`
	Allocated *int      `json:"allocated,omitempty"`
	Available *int      `json:"available,omitempty"`
	Used      []*uint32 `json:"used,omitempty"`
}

// NddrAsPoolAsPoolStateAllocatedAsTag struct
//type NddrAsPoolAsPoolStateAllocatedAsTag struct {
//	Key   *string `json:"key"`
//	Value *string `json:"value,omitempty"`
//}

// Root is the root of the schema
type Root struct {
	AspoolNddrAsPool *NddrAsPool `json:"nddr-as-pool,omitempty"`
}
