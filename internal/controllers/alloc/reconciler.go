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

package alloc

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
	aspoolv1alpha1 "github.com/yndd/nddr-as-pool/apis/aspool/v1alpha1"
	"github.com/yndd/nddr-as-pool/internal/rpool"
	"github.com/yndd/nddr-as-pool/internal/shared"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	finalizerName = "finalizer.alloc.ipam.nddr.yndd.io"
	//
	reconcileTimeout = 1 * time.Minute
	longWait         = 1 * time.Minute
	mediumWait       = 30 * time.Second
	shortWait        = 15 * time.Second
	veryShortWait    = 5 * time.Second

	// Errors
	errGetK8sResource = "cannot get alloc resource"
	errUpdateStatus   = "cannot update status of alloc resource"

	// events
	reasonReconcileSuccess             event.Reason = "ReconcileSuccess"
	reasonCannotDelete                 event.Reason = "CannotDeleteResource"
	reasonCannotAddFInalizer           event.Reason = "CannotAddFinalizer"
	reasonCannotDeleteFInalizer        event.Reason = "CannotDeleteFinalizer"
	reasonCannotInitialize             event.Reason = "CannotInitializeResource"
	reasonCannotGetAllocations         event.Reason = "CannotGetAllocations"
	reasonAppLogicFailed               event.Reason = "ApplogicFailed"
	reasonCannotParseIpPrefix          event.Reason = "CannotParseIpPrefix"
	reasonCannotDeleteDueToAllocations event.Reason = "CannotDeleteIpPrefixDueToExistingAllocations"
)

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// Reconciler reconciles packages.
type Reconciler struct {
	client  resource.ClientApplicator
	log     logging.Logger
	record  event.Recorder
	managed mrManaged

	newAlloc func() aspoolv1alpha1.Aa

	pool map[string]rpool.Pool
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

func WithNewReourceFn(f func() aspoolv1alpha1.Aa) ReconcilerOption {
	return func(r *Reconciler) {
		r.newAlloc = f
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

// Setup adds a controller that reconciles ipam.
func Setup(mgr ctrl.Manager, o controller.Options, nddcopts *shared.NddControllerOptions) error {
	name := "nddr/" + strings.ToLower(aspoolv1alpha1.AllocGroupKind)
	fn := func() aspoolv1alpha1.Aa { return &aspoolv1alpha1.Alloc{} }

	r := NewReconciler(mgr,
		WithLogger(nddcopts.Logger.WithValues("controller", name)),
		WithNewReourceFn(fn),
		WithPool(nddcopts.Pool),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&aspoolv1alpha1.Alloc{}).
		//Watches(&source.Kind{Type: &ipamv1alpha1.Ipam{}}, ipamHandler).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
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
	}

	for _, f := range opts {
		f(r)
	}

	return r
}

// Reconcile ipam allocation.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling alloc", "NameSpace", req.NamespacedName)

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	cr := r.newAlloc()
	if err := r.client.Get(ctx, req.NamespacedName, cr); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug("Cannot get managed resource", "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetK8sResource)
	}
	record := r.record.WithAnnotations("name", cr.GetAnnotations()[cr.GetName()])

	treename := strings.Join([]string{cr.GetNamespace(), cr.GetAsPoolName()}, "/")

	log.Debug("TreeName", "Name", treename)

	if meta.WasDeleted(cr) {
		log = log.WithValues("deletion-timestamp", cr.GetDeletionTimestamp())

		// check allocations
		if _, ok := r.pool[treename]; ok {
			if as, ok := cr.HasAs(); ok {
				r.pool[treename].DeAllocate(as)
			}
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

		// We've successfully delete our resource (if necessary) and
		// removed our finalizer. If we assume we were the only controller that
		// added a finalizer to this resource then it should no longer exist and
		// thus there is no point trying to update its status.
		log.Debug("Successfully deleted resource")
		return reconcile.Result{Requeue: false}, nil
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

	/*
		if err := cr.InitializeResource(); err != nil {
			record.Event(cr, event.Warning(reasonCannotInitialize, err))
			log.Debug("Cannot initialize", "error", err)
			cr.SetConditions(nddv1.ReconcileError(err), ipamv1alpha1.NotReady())
			return reconcile.Result{Requeue: true}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
		}
	*/

	if err := r.handleAppLogic(ctx, cr, treename); err != nil {
		record.Event(cr, event.Warning(reasonAppLogicFailed, err))
		log.Debug("handle applogic failed", "error", err)
		cr.SetConditions(nddv1.ReconcileError(err), aspoolv1alpha1.NotReady())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}

	cr.SetConditions(nddv1.ReconcileSuccess(), aspoolv1alpha1.Ready())
	return reconcile.Result{}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
}

func (r *Reconciler) handleAppLogic(ctx context.Context, cr aspoolv1alpha1.Aa, treename string) error {
	log := r.log.WithValues("name", cr.GetName())
	log.Debug("handleAppLogic")
	// get the ipam -> we need this mainly for parent status
	aspool := &aspoolv1alpha1.AsPool{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Namespace: cr.GetNamespace(),
		Name:      cr.GetAsPoolName()}, aspool); err != nil {
		// can happen when the ipam is not found
		log.Debug("Aspool not available")
		return errors.Wrap(err, "Aspool not available")
	}

	if _, ok := r.pool[treename]; !ok {
		log.Debug("AS pool/tree not ready", "treename", treename)
		return errors.New(fmt.Sprintf("AS pool/tree not ready, treename: %s", treename))
	}

	// the selector is used in the pool to find the entry in the pool
	// we use the labels with src-tag in the key
	// TBD how do we handle annotation changes
	selector := labels.NewSelector()
	sourcetags := make(map[string]string)
	for key, val := range cr.GetSourceTag() {
		req, err := labels.NewRequirement(key, selection.In, []string{val})
		if err != nil {
			log.Debug("wrong object", "error", err)
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
		if index, ok := cr.GetSelector()["index"]; ok {
			idx, err = strconv.Atoi(index)
			if err != nil {
				log.Debug("index conversion failed")
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
		switch aspool.GetAllocationStrategy() {
		case "deterministic":
			if a, ok := r.pool[treename].Allocate(utils.Uint32Ptr(uint32(idx)), sourcetags); !ok {
				log.Debug("pool allocation failed")
				return errors.New("pool allocation failed")
			} else {
				as = a
			}
		default:
			if a, ok := r.pool[treename].Allocate(nil, sourcetags); !ok {
				log.Debug("pool allocation failed")
				return errors.New("pool allocation failed")
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
	cr.SetAs(as)

	return nil

}
