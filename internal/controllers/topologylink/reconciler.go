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

package topologylink

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/yndd/ndd-runtime/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	topov1alpha1 "github.com/yndd/nddr-topology/apis/topo/v1alpha1"
	"github.com/yndd/nddr-topology/internal/shared"
)

const (
	finalizerName = "finalizer.topologylink.topo.nddr.yndd.io"
	//
	reconcileTimeout = 1 * time.Minute
	longWait         = 1 * time.Minute
	mediumWait       = 30 * time.Second
	shortWait        = 15 * time.Second
	veryShortWait    = 5 * time.Second

	// Errors
	errGetK8sResource = "cannot get topologylink resource"
	errUpdateStatus   = "cannot update status of topologylink resource"

	// events
	reasonReconcileSuccess      event.Reason = "ReconcileSuccess"
	reasonCannotDelete          event.Reason = "CannotDeleteResource"
	reasonCannotAddFInalizer    event.Reason = "CannotAddFinalizer"
	reasonCannotDeleteFInalizer event.Reason = "CannotDeleteFinalizer"
	reasonCannotInitialize      event.Reason = "CannotInitializeResource"
	reasonCannotGetAllocations  event.Reason = "CannotGetAllocations"
	reasonAppLogicFailed        event.Reason = "ApplogicFailed"
)

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// Reconciler reconciles packages.
type Reconciler struct {
	client  resource.ClientApplicator
	log     logging.Logger
	record  event.Recorder
	managed mrManaged

	newTopologyLink func() topov1alpha1.Tl
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

func WithNewReourceFn(f func() topov1alpha1.Tl) ReconcilerOption {
	return func(r *Reconciler) {
		r.newTopologyLink = f
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

// Setup adds a controller that reconciles topologylink.
func Setup(mgr ctrl.Manager, o controller.Options, nddcopts *shared.NddControllerOptions) error {
	name := "nddr/" + strings.ToLower(topov1alpha1.TopologyLinkGroupKind)
	fn := func() topov1alpha1.Tl { return &topov1alpha1.TopologyLink{} }

	r := NewReconciler(mgr,
		WithLogger(nddcopts.Logger.WithValues("controller", name)),
		WithNewReourceFn(fn),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	topologyHandler := &EnqueueRequestForAllTopologies{
		client: mgr.GetClient(),
		log:    nddcopts.Logger,
		ctx:    context.Background(),
	}

	topologyNodeHandler := &EnqueueRequestForAllTopologyNodes{
		client: mgr.GetClient(),
		log:    nddcopts.Logger,
		ctx:    context.Background(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&topov1alpha1.TopologyLink{}).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		Watches(&source.Kind{Type: &topov1alpha1.Topology{}}, topologyHandler).
		Watches(&source.Kind{Type: &topov1alpha1.TopologyNode{}}, topologyNodeHandler).
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

// Reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling topologylink", "NameSpace", req.NamespacedName)

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	cr := r.newTopologyLink()
	if err := r.client.Get(ctx, req.NamespacedName, cr); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug("Cannot get managed resource", "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetK8sResource)
	}
	record := r.record.WithAnnotations("name", cr.GetAnnotations()[cr.GetName()])

	if meta.WasDeleted(cr) {
		log = log.WithValues("deletion-timestamp", cr.GetDeletionTimestamp())

		if err := r.managed.RemoveFinalizer(ctx, cr); err != nil {
			// If this is the first time we encounter this issue we'll be
			// requeued implicitly when we update our status with the new error
			// condition. If not, we requeue explicitly, which will trigger
			// backoff.
			record.Event(cr, event.Warning(reasonCannotDeleteFInalizer, err))
			log.Debug("Cannot remove managed resource finalizer", "error", err)
			cr.SetConditions(nddv1.ReconcileError(err), topov1alpha1.NotReady())
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
		cr.SetConditions(nddv1.ReconcileError(err), topov1alpha1.NotReady())
		return reconcile.Result{Requeue: true}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}

	if err := cr.InitializeResource(); err != nil {
		record.Event(cr, event.Warning(reasonCannotInitialize, err))
		log.Debug("Cannot initialize", "error", err)
		cr.SetConditions(nddv1.ReconcileError(err), topov1alpha1.NotReady())
		return reconcile.Result{Requeue: true}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}

	if err := r.handleAppLogic(ctx, cr); err != nil {
		record.Event(cr, event.Warning(reasonAppLogicFailed, err))
		log.Debug("handle applogic failed", "error", err)
		cr.SetConditions(nddv1.ReconcileError(err), topov1alpha1.NotReady())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}

	cr.SetConditions(nddv1.ReconcileSuccess(), topov1alpha1.Ready())
	// we don't need to requeue for topology
	return reconcile.Result{}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
}

func (r *Reconciler) handleAppLogic(ctx context.Context, cr topov1alpha1.Tl) error {
	// get the topology
	topo := &topov1alpha1.Topology{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Namespace: cr.GetNamespace(),
		Name:      cr.GetTopologyName()}, topo); err != nil {
		// can happen when the topology is not found
		return err
	}
	// topology found

	if err := r.handleStatus(ctx, cr, topo); err != nil {
		return err
	}

	if err := r.parseLink(ctx, cr); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) handleStatus(ctx context.Context, cr topov1alpha1.Tl, topo *topov1alpha1.Topology) error {
	// topology found

	if topo.GetStatus() == "down" {
		cr.SetStatus("down")
		cr.SetReason("parent status down")
	} else {
		if cr.GetAdminState() == "disable" {
			cr.SetStatus("down")
			cr.SetReason("admin disable")
		} else {
			cr.SetStatus("up")
			cr.SetReason("")
		}
	}
	return nil
}

func (r *Reconciler) parseLink(ctx context.Context, cr topov1alpha1.Tl) error {
	// parse link
	nameNodeA := cr.GetEndpointANodeName()
	interfaceNameA := cr.GetEndpointAInterfaceName()
	nodeA := &topov1alpha1.TopologyNode{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Namespace: cr.GetNamespace(),
		Name:      nameNodeA}, nodeA); err != nil {
		// node A is not found
		cr.SetStatus("down")
		cr.SetReason("nodeA not found")
		return err
	}
	nameNodeB := cr.GetEndpointBNodeName()
	interfaceNameB := cr.GetEndpointBInterfaceName()
	nodeB := &topov1alpha1.TopologyNode{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Namespace: cr.GetNamespace(),
		Name:      nameNodeB}, nodeB); err != nil {
		// node A is not found
		cr.SetStatus("down")
		cr.SetReason("nodeB not found")
		return err
	}

	// check if link is part of a lag
	if cr.GetLag() {
		lagNameA := cr.GetLagAName()
		lagNameB := cr.GetLagBName()

		cr.SetNodeEndpoint(nodeA.GetName(), &topov1alpha1.NddrTopologyTopologyLinkStateNodeEndpoint{
			Name:       utils.StringPtr(lagNameA),
			Lag:        utils.BoolPtr(true),
			LagSubLink: utils.BoolPtr(false),
		})

		cr.SetNodeEndpoint(nodeB.GetName(), &topov1alpha1.NddrTopologyTopologyLinkStateNodeEndpoint{
			Name:       utils.StringPtr(lagNameB),
			Lag:        utils.BoolPtr(true),
			LagSubLink: utils.BoolPtr(false),
		})

		cr.SetNodeEndpoint(nodeA.GetName(), &topov1alpha1.NddrTopologyTopologyLinkStateNodeEndpoint{
			Name:       utils.StringPtr(interfaceNameA),
			Lag:        utils.BoolPtr(false),
			LagSubLink: utils.BoolPtr(true),
		})

		cr.SetNodeEndpoint(nodeB.GetName(), &topov1alpha1.NddrTopologyTopologyLinkStateNodeEndpoint{
			Name:       utils.StringPtr(interfaceNameB),
			Lag:        utils.BoolPtr(false),
			LagSubLink: utils.BoolPtr(true),
		})

		return nil
	}

	cr.SetNodeEndpoint(nodeA.GetName(), &topov1alpha1.NddrTopologyTopologyLinkStateNodeEndpoint{
		Name:       utils.StringPtr(interfaceNameA),
		Lag:        utils.BoolPtr(false),
		LagSubLink: utils.BoolPtr(false),
	})

	cr.SetNodeEndpoint(nodeB.GetName(), &topov1alpha1.NddrTopologyTopologyLinkStateNodeEndpoint{
		Name:       utils.StringPtr(interfaceNameB),
		Lag:        utils.BoolPtr(false),
		LagSubLink: utils.BoolPtr(false),
	})
	return nil
}
