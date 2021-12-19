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

package topologynode

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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	topov1alpha1 "github.com/yndd/nddr-topology/apis/topo/v1alpha1"
	"github.com/yndd/nddr-topology/internal/shared"
)

const (
	finalizerName = "finalizer.topologynode.topo.nddr.yndd.io"
	//
	reconcileTimeout = 1 * time.Minute
	longWait         = 1 * time.Minute
	mediumWait       = 30 * time.Second
	shortWait        = 15 * time.Second
	veryShortWait    = 5 * time.Second

	// Errors
	errGetK8sResource = "cannot get topologynode resource"
	errUpdateStatus   = "cannot update status of topologynode resource"

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

	newTopologyNode func() topov1alpha1.Tn
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

func WithNewReourceFn(f func() topov1alpha1.Tn) ReconcilerOption {
	return func(r *Reconciler) {
		r.newTopologyNode = f
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

// Setup adds a controller that reconciles topologynode.
func Setup(mgr ctrl.Manager, o controller.Options, nddcopts *shared.NddControllerOptions) error {
	name := "nddr/" + strings.ToLower(topov1alpha1.TopologyNodeGroupKind)
	fn := func() topov1alpha1.Tn { return &topov1alpha1.TopologyNode{} }

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

	topologyLinkHandler := &EnqueueRequestForAllTopologyLinks{
		client: mgr.GetClient(),
		log:    nddcopts.Logger,
		ctx:    context.Background(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&topov1alpha1.TopologyNode{}).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		Watches(&source.Kind{Type: &topov1alpha1.Topology{}}, topologyHandler).
		Watches(&source.Kind{Type: &topov1alpha1.TopologyLink{}}, topologyLinkHandler).
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
	log.Debug("Reconciling topologynode", "NameSpace", req.NamespacedName)

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	cr := r.newTopologyNode()
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
	// we don't need to requeue for topology node
	return reconcile.Result{}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
}

func (r *Reconciler) handleAppLogic(ctx context.Context, cr topov1alpha1.Tn) error {
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

	if err := r.setPosition(ctx, cr, topo); err != nil {
		return err
	}

	if err := r.setPlatform(ctx, cr, topo); err != nil {
		return err
	}

	/*
		if err := r.parseNode(ctx, cr); err != nil {
			return err
		}
	*/
	return nil
}

func (r *Reconciler) handleStatus(ctx context.Context, cr topov1alpha1.Tn, topo *topov1alpha1.Topology) error {
	r.log.Debug("handle node status", "topo admin status", topo.GetAdminState(), "topo status", topo.GetStatus())
	if topo.GetAdminState() == "disable" {
		cr.SetStatus("down")
		cr.SetReason("parent status down")
	} else {
		if cr.GetAdminState() == "disable" {
			cr.SetStatus("down")
			cr.SetReason("admin disabled")
		} else {
			cr.SetStatus("up")
			cr.SetReason("")
		}
	}
	return nil
}

func (r *Reconciler) setPosition(ctx context.Context, cr topov1alpha1.Tn, topo *topov1alpha1.Topology) error {
	r.log.Debug("SetPosition", "platform", cr.GetPlatform())

	switch cr.GetKindName() {
	case "sros", "srl":
		cr.SetPosition(topov1alpha1.PositionNetwork.String())
	default:
		cr.SetPosition(topov1alpha1.PositionEndpoint.String())
	}

	return nil
}

func (r *Reconciler) setPlatform(ctx context.Context, cr topov1alpha1.Tn, topo *topov1alpha1.Topology) error {
	r.log.Debug("Setflatform", "platform", cr.GetPlatform())
	if cr.GetPlatform() == "" && cr.GetPosition() == topov1alpha1.PositionNetwork.String() {
		// platform is not defined at node level
		p := topo.GetPlatformByKindName(cr.GetKindName())
		if p != "" {
			cr.SetPlatform(p)
			return nil
		}
		p = topo.GetPlatformFromDefaults()
		if p != "" {
			cr.SetPlatform(p)
			return nil
		}
		// platform is not defined we use the global default
		cr.SetPlatform("ixrd2")
		return nil

	}
	// all good since the platform is already set
	return nil
}

/*
func (r *Reconciler) parseNode(ctx context.Context, cr topov1alpha1.Tn) error {
	// list link
	// TODO we could potentially improve the parsing based on annotations; tbd
	link := &topov1alpha1.TopologyLinkList{}
	if err := r.client.List(ctx, link); err != nil {
		return err
	}

	for _, topolink := range link.Items {
		// only process the topology matches
		if topolink.GetTopologyName() == cr.GetTopologyName() {
			// only process if the link is up
			if topolink.GetStatus() == "up" {
				for _, n := range topolink.GetNodes() {
					if *n.Name == cr.GetName() {
						for _, ep := range n.Endpoint {
							cr.SetNodeEndpoint(ep)
						}
					}
				}
			}
		}
	}
	return nil

}
*/
