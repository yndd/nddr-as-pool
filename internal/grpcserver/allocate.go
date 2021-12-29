/*
Copyright 2021 NDDO.

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

package grpcserver

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/utils"
	"github.com/yndd/nddo-grpc/resource/resourcepb"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
)

func (r *server) ResourceGet(ctx context.Context, req *resourcepb.Request) (*resourcepb.Reply, error) {
	log := r.log.WithValues("Request", req)
	log.Debug("ResourceGet...")

	return &resourcepb.Reply{Ready: true}, nil
}

func (r *server) ResourceAlloc(ctx context.Context, req *resourcepb.Request) (*resourcepb.Reply, error) {
	log := r.log.WithValues("Request", req)
	log.Debug("ResourceAlloc...")

	namespace := req.GetNamespace()
	allocpoolName := req.GetResourceName()

	crName := strings.Join([]string{namespace, allocpoolName}, ".")

	aspool := r.newAsPool()
	if err := r.client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      allocpoolName}, aspool); err != nil {
		// can happen when the networkinstance is not found
		log.Debug("NetworkInstance not available")
		return &resourcepb.Reply{Ready: false}, errors.New("NetworkInstance not available")
	}

	if _, ok := r.pool[crName]; !ok {
		log.Debug("AS pool/tree not ready", "treename", crName)
		return &resourcepb.Reply{Ready: false}, errors.New("NetworkInstance not available")
	}

	// the selector is used in the pool to find the entry in the pool
	// we use the labels with src-tag in the key
	// TBD how do we handle annotation changes
	selector := labels.NewSelector()
	sourcetags := make(map[string]string)
	for key, val := range req.GetAlloc().GetSourceTag() {
		req, err := labels.NewRequirement(key, selection.In, []string{val})
		if err != nil {
			log.Debug("wrong object", "error", err)
			return &resourcepb.Reply{Ready: true}, err
		}
		selector = selector.Add(*req)
		sourcetags[key] = val
	}

	var as uint32
	var ases []uint32
	var idx int
	var err error
	// query the pool to see if an allocation was performed using the selector
	switch aspool.GetAllocationStrategy() {
	case "deterministic":
		if index, ok := req.GetAlloc().GetSelector()["index"]; ok {
			idx, err = strconv.Atoi(index)
			if err != nil {
				log.Debug("index conversion failed")
			}
			ases = r.pool[crName].QueryByIndex(idx)
		} else {
			log.Debug("when allocation strategy is deterministic a valid index as selector needs to be assigned in the spec")
			return &resourcepb.Reply{Ready: true}, errors.New("when allocation strategy is deterministic a valid index as selector needs to be assigned in the spec")
		}
	default:
		// first available allocation strategy
		ases = r.pool[crName].QueryByLabels(selector)
	}

	if len(ases) == 0 {
		// label/selector not found in the pool -> allocate AS in pool
		switch aspool.GetAllocationStrategy() {
		case "deterministic":
			if a, ok := r.pool[crName].Allocate(utils.Uint32Ptr(uint32(idx)), sourcetags); !ok {
				log.Debug("pool allocation failed")
				return &resourcepb.Reply{Ready: true}, errors.New("pool allocation failed")
			} else {
				as = a
			}
		default:
			if a, ok := r.pool[crName].Allocate(nil, sourcetags); !ok {
				log.Debug("pool allocation failed")
				return &resourcepb.Reply{Ready: true}, errors.New("pool allocation failed")
			} else {
				as = a
			}
		}
	} else {
		// label/selector found or allocated
		as = ases[0]
		if len(ases) > 1 {
			// this should never happen since the metalabels will point to the same entry
			// in the pool
			log.Debug("strange situation, as found in pool multiple times", "ases", ases)
		}
	}

	// set the as in the alloc object
	log.Debug("handleAppLogic allocated AS", "AS", as)

	return &resourcepb.Reply{
		Ready:      true,
		Timestamp:  time.Now().UnixNano(),
		ExpiryTime: time.Now().UnixNano(),
		Data: map[string]*resourcepb.TypedValue{
			"as": {Value: &resourcepb.TypedValue_IntVal{IntVal: int64(as)}},
		},
	}, nil
}

func (s *server) ResourceDeAlloc(ctx context.Context, req *resourcepb.Request) (*resourcepb.Reply, error) {
	log := s.log.WithValues("Request", req)
	log.Debug("ResourceDeAlloc...")

	return &resourcepb.Reply{Ready: true}, nil
}
