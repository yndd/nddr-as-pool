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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/yndd/ndd-runtime/pkg/utils"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	aspoolv1alpha1 "github.com/yndd/nddr-as-pool/apis/aspool/v1alpha1"
	"github.com/yndd/nddr-as-pool/internal/rpool"
	"github.com/yndd/nddr-as-pool/internal/shared"
)

const (
	finalizerName = "finalizer.alloc.aspool.nddr.yndd.io"
	//
	reconcileTimeout = 1 * time.Minute
	longWait         = 1 * time.Minute
	mediumWait       = 30 * time.Second
	shortWait        = 15 * time.Second
	veryShortWait    = 5 * time.Second

	// Errors
	errGetK8sResource = "cannot get aspool resource"
	errUpdateStatus   = "cannot update status of aspool resource"

	// events
	reasonReconcileSuccess      event.Reason = "ReconcileSuccess"
	reasonCannotDelete          event.Reason = "CannotDeleteResource"
	reasonCannotAddFInalizer    event.Reason = "CannotAddFinalizer"
	reasonCannotDeleteFInalizer event.Reason = "CannotDeleteFinalizer"
	reasonCannotInitialize      event.Reason = "CannotInitializeResource"
	reasonCannotGetAllocations  event.Reason = "CannotGetAllocations"
	reasonAppLogicFailed        event.Reason = "ApplogicFailed"
	reasonCannotGarbageCollect  event.Reason = "CannotGarbageCollect"
)

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// Reconciler reconciles packages.
type Reconciler struct {
	client  resource.ClientApplicator
	log     logging.Logger
	record  event.Recorder
	managed mrManaged

	newAsPool func() aspoolv1alpha1.Ap
	pool      map[string]rpool.Pool
}

type mrManaged struct {
	resource.Finalizer
}

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.log = log
	}
}

func WithNewAsPoolFn(f func() aspoolv1alpha1.Ap) ReconcilerOption {
	return func(r *Reconciler) {
		r.newAsPool = f
	}
}

func WithPool(t map[string]rpool.Pool) ReconcilerOption {
	return func(r *Reconciler) {
		r.pool = t
	}
}

// WithRecorder specifies how the Reconciler should record Kubernetes events.
func WithRecorder(er event.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = er
	}
}

func defaultMRManaged(m ctrl.Manager) mrManaged {
	return mrManaged{
		Finalizer: resource.NewAPIFinalizer(m.GetClient(), finalizerName),
	}
}

// SetupAsPool adds a controller that reconciles AsPools.
func Setup(mgr ctrl.Manager, o controller.Options, nddcopts *shared.NddControllerOptions) error {
	name := "nddr/" + strings.ToLower(aspoolv1alpha1.AsPoolGroupKind)
	ap := func() aspoolv1alpha1.Ap { return &aspoolv1alpha1.AsPool{} }

	r := NewReconciler(mgr,
		WithLogger(nddcopts.Logger.WithValues("controller", name)),
		WithNewAsPoolFn(ap),
		WithPool(nddcopts.Pool),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	allocHandler := &EnqueueRequestForAllAllocations{
		client: mgr.GetClient(),
		log:    nddcopts.Logger,
		ctx:    context.Background(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&aspoolv1alpha1.AsPool{}).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		Watches(&source.Kind{Type: &aspoolv1alpha1.Alloc{}}, allocHandler).
		//Watches(&source.Channel{Source: events}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

// NewReconciler creates a new reconciler.
func NewReconciler(mgr ctrl.Manager, opts ...ReconcilerOption) *Reconciler {

	r := &Reconciler{
		client: resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
		},
		log:     logging.NewNopLogger(),
		record:  event.NewNopRecorder(),
		managed: defaultMRManaged(mgr),
		pool:    make(map[string]rpool.Pool),
	}

	for _, f := range opts {
		f(r)
	}

	return r
}

// Reconcile ipam allocation.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling aspool", "NameSpace", req.NamespacedName)

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	cr := r.newAsPool()
	//cr := &aspoolv1alpha1.AsPool{}
	if err := r.client.Get(ctx, req.NamespacedName, cr); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug("Cannot get managed resource", "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetK8sResource)
	}
	record := r.record.WithAnnotations("name", cr.GetAnnotations()[cr.GetName()])

	treename := strings.Join([]string{cr.GetNamespace(), cr.GetName()}, ".")

	// initialize the status
	if meta.WasDeleted(cr) {
		log = log.WithValues("deletion-timestamp", cr.GetDeletionTimestamp())

		if cr.GetAllocations() != 0 {
			record.Event(cr, event.Normal(reasonCannotDelete, "allocations present"))
			log.Debug("Allocations present we cannot delete the as pool")

			if _, ok := r.pool[treename]; !ok {
				r.log.Debug("AS pool/tree init", "treename", treename)
				r.pool[treename] = rpool.New(cr.GetStart(), cr.GetEnd(), cr.GetAllocationStrategy())
			}

			if err := r.GarbageCollection(ctx, cr, treename); err != nil {
				record.Event(cr, event.Warning(reasonCannotGarbageCollect, err))
				log.Debug("Cannot perform garbage collection", "error", err)
				cr.SetConditions(nddv1.ReconcileError(err), aspoolv1alpha1.NotReady())
				return reconcile.Result{Requeue: true}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
			}

			return reconcile.Result{RequeueAfter: mediumWait}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
		}

		if err := r.managed.RemoveFinalizer(ctx, cr); err != nil {
			// If this is the first time we encounter this issue we'll be
			// requeued implicitly when we update our status with the new error
			// condition. If not, we requeue explicitly, which will trigger
			// backoff.
			record.Event(cr, event.Warning(reasonCannotDeleteFInalizer, err))
			log.Debug("Cannot remove managed resource finalizer", "error", err)
			cr.SetConditions(nddv1.ReconcileError(err), aspoolv1alpha1.NotReady())
			return reconcile.Result{Requeue: true}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
		}

		delete(r.pool, req.NamespacedName.String())

		// We've successfully deallocated our external resource (if necessary) and
		// removed our finalizer. If we assume we were the only controller that
		// added a finalizer to this resource then it should no longer exist and
		// thus there is no point trying to update its status.
		log.Debug("Successfully deallocated resource")
		return reconcile.Result{Requeue: false}, nil
	}

	if cr.GetStart() > cr.GetEnd() {
		log.Debug("wrong spec start should be <= end", "start", cr.GetStart(), "end", cr.GetEnd())
		cr.SetConditions(nddv1.ReconcileError(errors.New(fmt.Sprintf("wrong spec start %d should be <= end %d", cr.GetStart(), cr.GetEnd()))), aspoolv1alpha1.NotReady())
		return reconcile.Result{Requeue: true}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}

	if err := r.managed.AddFinalizer(ctx, cr); err != nil {
		// If this is the first time we encounter this issue we'll be requeued
		// implicitly when we update our status with the new error condition. If
		// not, we requeue explicitly, which will trigger backoff.
		record.Event(cr, event.Warning(reasonCannotAddFInalizer, err))
		log.Debug("Cannot add finalizer", "error", err)
		cr.SetConditions(nddv1.ReconcileError(err), aspoolv1alpha1.NotReady())
		return reconcile.Result{Requeue: true}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}

	// initialize resource
	if err := cr.InitializeResource(); err != nil {
		record.Event(cr, event.Warning(reasonCannotInitialize, err))
		log.Debug("Cannot initialize", "error", err)
		cr.SetConditions(nddv1.ReconcileError(err), aspoolv1alpha1.NotReady())
		return reconcile.Result{Requeue: true}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}

	if err := r.handleAppLogic(ctx, cr, treename); err != nil {
		record.Event(cr, event.Warning(reasonAppLogicFailed, err))
		log.Debug("handle applogic failed", "error", err)
		cr.SetConditions(nddv1.ReconcileError(err), aspoolv1alpha1.NotReady())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}

	if err := r.GarbageCollection(ctx, cr, treename); err != nil {
		record.Event(cr, event.Warning(reasonCannotGarbageCollect, err))
		log.Debug("Cannot perform garbage collection", "error", err)
		cr.SetConditions(nddv1.ReconcileError(err), aspoolv1alpha1.NotReady())
		return reconcile.Result{Requeue: true}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}

	cr.SetConditions(nddv1.ReconcileSuccess(), aspoolv1alpha1.Ready())
	return reconcile.Result{RequeueAfter: reconcileTimeout}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
}

func (r *Reconciler) handleAppLogic(ctx context.Context, cr aspoolv1alpha1.Ap, treename string) error {
	// initialize the pool
	if _, ok := r.pool[treename]; !ok {
		r.log.Debug("AS pool/tree init", "treename", treename)
		r.pool[treename] = rpool.New(cr.GetStart(), cr.GetEnd(), cr.GetAllocationStrategy())
	}

	cr.SetOrganizationName(cr.GetOrganizationName())
	cr.SetDeploymentName(cr.GetDeploymentName())
	cr.SetAsPoolName(cr.GetAsPoolName())

	return nil
}

func (r *Reconciler) GarbageCollection(ctx context.Context, cr aspoolv1alpha1.Ap, treename string) error {
	log := r.log.WithValues("function", "garbageCollection", "Name", cr.GetName())
	//log.Debug("entry")
	// get all allocations
	alloc := &aspoolv1alpha1.AllocList{}
	if err := r.client.List(ctx, alloc); err != nil {
		log.Debug("Cannot get allocations", "error", err)
		return err
	}

	poolname := cr.GetName()
	//namespace := cr.GetNamespace()
	//poolname := strings.Join([]string{namespace, aspoolname}, "/")

	// check if allocations dont have AS allocated
	// check if allocations match with allocated pool
	// -> alloc found in pool -> ok
	// -> alloc not found in pool -> assign in pool
	// we keep track of the allocated ASes to compare in a second stage
	allocAses := make([]*uint32, 0)
	for _, alloc := range alloc.Items {
		allocpoolName := strings.Join(strings.Split(alloc.GetName(), ".")[:len(strings.Split(alloc.GetName(), "."))-1], ".")
		// only garbage collect if the pool matches
		log.Debug("Alloc PoolName comparison", "AS Pool Name", poolname, "Alloc Pool Name", allocpoolName)
		if allocpoolName == poolname {
			allocAs, allocAsFound := alloc.HasAs()
			if !allocAsFound {
				log.Debug("Alloc", "Pool", poolname, "Name", alloc.GetName(), "AS found", allocAsFound)
			} else {
				log.Debug("Alloc", "Pool", poolname, "Name", alloc.GetName(), "AS found", allocAsFound, "allocAS", allocAs)
			}

			// the selector is used in the pool to find the entry in the pool
			// we use the annotations with src-tag in the key
			// TBD how do we handle annotation
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
			var ok bool

			var as uint32
			var ases []uint32
			var idx int
			var err error
			// query the pool to see if an allocation was performed using the selector
			switch cr.GetAllocationStrategy() {
			case "deterministic":
				if index, ok := alloc.GetSelector()["index"]; ok {
					idx, err = strconv.Atoi(index)
					if err != nil {
						log.Debug("index conversion failed")
						return err
					}
					ases = r.pool[treename].QueryByIndex(idx)
				} else {
					log.Debug("when allocation strategy is deterministic a valid index as selector needs to be assigned in the spec")
					return errors.New("when allocation strategy is deterministic a valid index as selector needs to be assigned in the spec")
				}
			default:
				// first available allocation strategy
				ases = r.pool[treename].QueryByLabels(selector)
			}

			if len(ases) == 0 {
				// label/selector not found in the pool -> allocate AS in pool
				// this can happen if the allocation was not assigned an AS
				if allocAsFound {
					// can happen during startup that an allocation had already an AS
					// we need to ensure we allocate the same AS again
					switch cr.GetAllocationStrategy() {
					case "deterministic":
						log.Debug("strange: With a determinsitic allocation the as should always be found")
						as, ok = r.pool[treename].Allocate(utils.Uint32Ptr(uint32(idx)), sourcetags)
					default:
						as, ok = r.pool[treename].Allocate(&allocAs, sourcetags)
					}
					log.Debug("strange situation: query not found in pool, but alloc has an AS (can happen during startup/restart)", "allocAS", allocAs, "new AS", as)
				} else {
					switch cr.GetAllocationStrategy() {
					case "deterministic":
						log.Debug("strange: With a determinsitic allocation the as should always be found")
						as, ok = r.pool[treename].Allocate(utils.Uint32Ptr(uint32(idx)), sourcetags)
					default:
						as, ok = r.pool[treename].Allocate(nil, sourcetags)
					}
				}
				if !ok {
					log.Debug("pool allocation failed")
					return errors.New("Pool allocation failed")
				}
				//alloc.SetAs(as)
				// find AS in the used list in the aspool resource
				if !cr.FindAs(as) {
					//cr.SetAs(as)
				} else {
					log.Debug("strange situation, query not found in pool, but AS was found in used allocations (can happen during startup/restart) ")
				}
				//if err := r.client.Status().Update(ctx, &alloc); err != nil {
				//	log.Debug("updating alloc status", "error", err)
				//}

			} else {
				// label/selector found in the pool
				as = ases[0]
				if len(ases) > 1 {
					// this should never happen since the metalabels will point to the same entry
					// in the pool
					log.Debug("strange situation, AS found in pool multiple times", "ases", ases)
				}
				switch {
				case !allocAsFound:
					log.Debug("strange situation, AS found in pool but alloc AS not found")
					//alloc.SetAs(as)
					//if err := r.client.Status().Update(ctx, &alloc); err != nil {
					//	log.Debug("updating alloc status", "error", err)
					//}
				case allocAsFound && as != allocAs:
					log.Debug("strange situation, AS found in pool but alloc AS had different AS", "pool AS", as, "alloc AS", allocAs)
					//alloc.SetAs(as)
					//if err := r.client.Status().Update(ctx, &alloc); err != nil {
					//	log.Debug("updating alloc status", "error", err)
					//}
				default:
					// do nothing, all ok
				}
			}
			allocAses = append(allocAses, &as)
		}
	}
	// based on the allocated ASes we collected before we can validate if the
	// pool had assigned other allocations
	found := false
	for _, pAs := range r.pool[treename].GetAllocated() {
		for _, allocAs := range allocAses {
			if *allocAs == pAs {
				found = true
				break
			}
		}
		if !found {
			log.Debug("as found in pool, but no alloc found -> deallocate from pool", "as", pAs)
			r.pool[treename].DeAllocate(pAs)
		}
	}
	// always update the status field in the aspool with the latest info
	cr.UpdateAs(allocAses)

	return nil
}
