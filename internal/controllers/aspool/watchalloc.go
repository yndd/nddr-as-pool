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

package aspool

import (
	"context"
	"strconv"
	"strings"

	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/utils"
	aspoolv1alpha1 "github.com/yndd/nddr-as-pool/apis/aspool/v1alpha1"
	"github.com/yndd/nddr-as-pool/internal/rpool"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type adder interface {
	Add(item interface{})
}

type EnqueueRequestForAllAllocations struct {
	client client.Client
	log    logging.Logger
	pool   map[string]rpool.Pool
	ctx    context.Context
}

// Create enqueues a request for all infrastructures which pertains to the allocation.
func (e *EnqueueRequestForAllAllocations) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	e.handleEvent(evt.Object, q, "add")
}

// Create enqueues a request for all infrastructures which pertains to the allocation.
func (e *EnqueueRequestForAllAllocations) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	e.handleEvent(evt.ObjectOld, q, "add")
	e.handleEvent(evt.ObjectNew, q, "add")
}

// Create enqueues a request for all infrastructures which pertains to the allocation.
func (e *EnqueueRequestForAllAllocations) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	e.handleEvent(evt.Object, q, "delete")
}

// Create enqueues a request for all infrastructures which pertains to the allocation.
func (e *EnqueueRequestForAllAllocations) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.handleEvent(evt.Object, q, "add")
}

func (e *EnqueueRequestForAllAllocations) handleEvent(obj runtime.Object, queue adder, action string) {
	alloc, ok := obj.(*aspoolv1alpha1.Alloc)
	if !ok {
		e.log.Debug("wrong alloc object in watch")
		return
	}
	allocpoolname := alloc.GetAsPoolName()
	allocnamespace := alloc.GetNamespace()
	poolname := strings.Join([]string{allocnamespace, allocpoolname}, "/")

	log := e.log.WithValues("function", "watch alloc", "name", alloc.GetName(), "aspool", allocpoolname, "action", action)
	log.Debug("handleEvent")
	aspools := &aspoolv1alpha1.AsPoolList{}
	if err := e.client.List(context.TODO(), aspools); err != nil {
		return
	}

	// walk through all pool, process the resources that have pool name matches
	for _, aspool := range aspools.Items {
		aspoolname := aspool.GetName()
		// process if the pool name matches
		if allocpoolname == aspoolname {
			// the selector is used in the pool to find the entry in the pool
			// we use the labels with src-tag in the key
			// TBD how do we handle annotation changes
			selector := labels.NewSelector()
			sourcetags := make(map[string]string)
			for key, val := range alloc.GetSourceTag() {
				req, err := labels.NewRequirement(key, selection.In, []string{val})
				if err != nil {
					log.Debug("wrong object", "error", err)
				}
				selector = selector.Add(*req)
				sourcetags[key] = val
			}

			// validate if the pool was initialized, can happen during startup
			// that the pool was not initialized
			// garbage collection will take care of the consistency
			if _, ok := e.pool[poolname]; ok {
				var as uint32
				var ases []uint32
				var idx int
				var err error
				// query the pool to see if an allocation was performed using the selector
				switch aspool.GetAllocationStrategy() {
				case "deterministic":
					if index, ok := alloc.GetSelector()["index"]; ok {
						idx, err = strconv.Atoi(index)
						if err != nil {
							log.Debug("index conversion failed")
						}
						ases = e.pool[poolname].QueryByIndex(idx)
					} else {
						log.Debug("when allocation strategy is deterministic a valid index as selector needs to be assigned in the spec")
						return
					}
				default:
					// first available allocation strategy
					ases = e.pool[poolname].QueryByLabels(selector)
				}

				switch action {
				case "add":
					if len(ases) == 0 {
						// label/selector not found in the pool -> allocate AS in pool
						switch aspool.GetAllocationStrategy() {
						case "deterministic":
							log.Debug("strange: With a determinsitic allocation the as should always be found")
							as, ok = e.pool[poolname].Allocate(utils.Uint32Ptr(uint32(idx)), sourcetags)
						default:
							as, ok = e.pool[poolname].Allocate(nil, sourcetags)
						}
						if !ok {
							log.Debug("pool allocation failed")
							return
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
					// set the status in the aspool used objects and update the relevant counters
					//aspool.SetAs(as)
					// set the as in the alloc object
					alloc.SetAs(as)

					// update alloc status
					if err := e.client.Status().Update(e.ctx, alloc); err != nil {
						log.Debug("updating alloc status", "error", err)
					}
					// update status in aspool resource
					//if err := e.client.Status().Update(e.ctx, &aspool); err != nil {
					//	log.Debug("updating aspool status", "error", err)
					//}
				case "delete":
					if len(ases) == 0 {
						// label/selector not found in the pool
						// this should only happen if the alloc never got an allocation
						if allocAs, allocAsFound := alloc.HasAs(); allocAsFound {
							// this would be strange if this occurs, query results in not found,
							// but AS exists in allocate object
							log.Debug("delete alloc without an allocated as", "as", allocAs)
						}
						return
					}
					as = ases[0]
					// label/selector -> deallocate AS in the  pool
					e.pool[poolname].DeAllocate(as)
					// update status in aspool resource
					//aspool.UnSetAs(as)

					//if err := e.client.Status().Update(e.ctx, &aspool); err != nil {
					//	log.Debug("updating aspool status", "error", err)
					//}
				}
				queue.Add(reconcile.Request{NamespacedName: types.NamespacedName{
					Namespace: aspool.GetNamespace(),
					Name:      aspool.GetName()}})
			} else {
				log.Debug("pool not ready")
			}
		}
	}
}
