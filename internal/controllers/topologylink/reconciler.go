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
	"fmt"
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
	corev1 "k8s.io/api/core/v1"
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
	reasonAppLogicError         event.Reason = "ApplogicError"
	reasonCannotDeleteTags      event.Reason = "CannotDeleteTagsOfLogicalLink"
)

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// Reconciler reconciles packages.
type Reconciler struct {
	client  resource.ClientApplicator
	log     logging.Logger
	record  event.Recorder
	managed mrManaged

	hooks Hooks

	newTopology     func() topov1alpha1.Tp
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

// WitHooks specifies how the Reconciler should deploy child resources
func WithHooks(h Hooks) ReconcilerOption {
	return func(r *Reconciler) {
		r.hooks = h
	}
}

func WithNewReourceFn(f func() topov1alpha1.Tl) ReconcilerOption {
	return func(r *Reconciler) {
		r.newTopologyLink = f
	}
}

func WithNewTopologyFn(f func() topov1alpha1.Tp) ReconcilerOption {
	return func(r *Reconciler) {
		r.newTopology = f
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
	tpfn := func() topov1alpha1.Tp { return &topov1alpha1.Topology{} }

	r := NewReconciler(mgr,
		WithLogger(nddcopts.Logger.WithValues("controller", name)),
		WithHooks(NewHook(resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
		}, nddcopts.Logger.WithValues("nodehook", name))),
		WithNewReourceFn(fn),
		WithNewTopologyFn(tpfn),
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

	topologyLinkHandler := &EnqueueRequestForAllTopologyLinks{
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

		// we need to delete the tag if the member link gets deleted
		if cr.GetLagMember() {
			topologyName := strings.Join([]string{cr.GetOrganizationName(), cr.GetDeploymentName(), cr.GetTopologyName()}, ".")
			logicalLink, err := r.hooks.Get(ctx, cr, topologyName)
			if err == nil {
				r.log.Debug("logical link exists", "Logical Link", logicalLink.GetName())
				//for the multi-homed case we need to delete the tags of the member links
				// that match the mh name
				if err := r.hooks.DeleteApply(ctx, cr, logicalLink); err != nil {
					record.Event(cr, event.Warning(reasonCannotDeleteTags, err))
					log.Debug("Cannot delete tags of a logical link", "error", err)
					cr.SetConditions(nddv1.ReconcileError(err), topov1alpha1.NotReady())
					return reconcile.Result{Requeue: true}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
				}
			}
		}

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

	msg, err := r.handleAppLogic(ctx, cr)
	if err != nil {
		record.Event(cr, event.Warning(reasonAppLogicError, err))
		log.Debug("handle applogic error", "error", err)
		cr.SetConditions(nddv1.ReconcileError(err), topov1alpha1.NotReady())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}
	if msg != nil {
		//record.Event(cr, event.(reasonAppLogicFailed, msg))
		log.Debug("handle applogic failed", "msg", msg)
		cr.SetConditions(nddv1.ReconcileSuccess(), topov1alpha1.NotReady())
		return reconcile.Result{}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
	}

	cr.SetConditions(nddv1.ReconcileSuccess(), topov1alpha1.Ready())
	// we don't need to requeue for topology
	return reconcile.Result{}, errors.Wrap(r.client.Status().Update(ctx, cr), errUpdateStatus)
}

func (r *Reconciler) handleAppLogic(ctx context.Context, cr topov1alpha1.Tl) (*string, error) {
	topologyName := strings.Join([]string{cr.GetOrganizationName(), cr.GetDeploymentName(), cr.GetTopologyName()}, ".")

	// get the topo
	topo := r.newTopology()
	if err := r.client.Get(ctx, types.NamespacedName{
		Namespace: cr.GetNamespace(),
		Name:      topologyName}, topo); err != nil {
		// can happen when the resource is not found
		cr.SetStatus("down")
		cr.SetReason("topology not found")
		return nil, errors.Wrap(err, "topology not found")
	}
	if topo.GetCondition(topov1alpha1.ConditionKindReady).Status != corev1.ConditionTrue {
		cr.SetStatus("down")
		cr.SetReason("topology not found or ready")
		return nil, errors.New("topology not ready")
	}

	// topology found and ready

	if err := r.handleStatus(ctx, cr, topo, topologyName); err != nil {
		return nil, err
	}

	return r.parseLink(ctx, cr, topologyName)
}

func (r *Reconciler) handleStatus(ctx context.Context, cr topov1alpha1.Tl, topo topov1alpha1.Tp, topologyName string) error {
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

func (r *Reconciler) parseLink(ctx context.Context, cr topov1alpha1.Tl, topologyName string) (*string, error) {
	// parse link

	// validates if the nodes if the links are present in the k8s api are not
	// if an error occurs during validation an error is returned
	msg, err := r.validateNodes(ctx, cr, topologyName)
	if err != nil {
		return msg, err
	}
	if msg != nil {
		return msg, nil
	}

	// for infra links we set the kind at the link level using the information from the spec
	if cr.GetEndPointAKind() == topov1alpha1.LinkEPKindInfra.String() && cr.GetEndPointBKind() == topov1alpha1.LinkEPKindInfra.String() {
		cr.SetKind(topov1alpha1.LinkEPKindInfra.String())
	}

	if cr.GetLag() {
		// this is a logical link (single homes or multihomed), we dont need to process it since the member links take care
		// of crud operation
		cr.SetOrganizationName(cr.GetOrganizationName())
		cr.SetDeploymentName(cr.GetDeploymentName())
		cr.SetTopologyName(cr.GetTopologyName())
		return nil, nil
	}

	// check if the link is part of a lag
	if cr.GetLagMember() {
		logicalLink, err := r.hooks.Get(ctx, cr, topologyName)
		if err != nil {
			if resource.IgnoreNotFound(err) != nil {
				return nil, err
			}
			if err := r.hooks.Create(ctx, cr, topologyName); err != nil {
				return nil, err
			}
			r.log.Debug("logical link created")
			cr.SetOrganizationName(cr.GetOrganizationName())
			cr.SetDeploymentName(cr.GetDeploymentName())
			cr.SetTopologyName(cr.GetTopologyName())
			return nil, nil

		}
		r.log.Debug("logical link exists", "Logical Link", logicalLink.GetName())

		// for the multi-homed case we need to add the tags of the other member links
		// that match the mh name
		if err := r.hooks.Apply(ctx, cr, logicalLink); err != nil {
			return nil, err
		}

	}
	cr.SetOrganizationName(cr.GetOrganizationName())
	cr.SetDeploymentName(cr.GetDeploymentName())
	cr.SetTopologyName(cr.GetTopologyName())
	return nil, nil
}

func (r *Reconciler) validateNodes(ctx context.Context, cr topov1alpha1.Tl, topologyName string) (*string, error) {
	for i := 0; i <= 1; i++ {
		var multihoming bool
		var nodeName string
		var tags map[string]string
		lag := cr.GetLag()
		switch i {
		case 0:
			nodeName = cr.GetEndpointANodeName()
			multihoming = cr.GetEndPointAMultiHoming()
			tags = cr.GetEndpointATag()
		case 1:
			nodeName = cr.GetEndpointBNodeName()
			multihoming = cr.GetEndPointBMultiHoming()
			tags = cr.GetEndpointBTag()
		}

		// lag are logical links which are created based on member links
		// for singlehomed logical links if the node no longer exists, we delete the sh-logical-link
		// for multi-homed logical links if a member node no longer exists, we delete the tags related to the node
		// for multi-homed logical links of all member nodes no longer exist, we delete the mh-logical link
		if lag {
			if multihoming {
				// node validation happens through the endpoint tags
				// a nodetag has a prefix of node:
				found := false
				for k, v := range tags {
					if strings.Contains(k, topov1alpha1.NodePrefix) {
						nodeName := strings.TrimPrefix(k, topov1alpha1.NodePrefix+":")
						node := &topov1alpha1.TopologyNode{}
						if err := r.client.Get(ctx, types.NamespacedName{
							Namespace: cr.GetNamespace(),
							Name:      strings.Join([]string{topologyName, nodeName}, ".")}, node); err != nil {
							if resource.IgnoreNotFound(err) != nil {
								return nil, err
							}
							r.log.Debug("mh-ep logical-link:: member node not found, delete the ep node tags", "nodeName", nodeName)
							// node no longer exists, we can delete the node tags from the logocal element
							if err := r.hooks.DeleteApplyNode(ctx, cr, 0, k, v); err != nil {
								return nil, err
							}
						} else {
							found = true
						}
					}
				}
				if !found {
					// when none of the mh nodes are found we can delete the logical link
					if err := r.hooks.Delete(ctx, cr); err != nil {
						return nil, err
					}
					r.log.Debug("mh-ep logical-link: none of the member nodes wwere found, delete the logical-link")
					return nil, nil
				}
			} else {
				node := &topov1alpha1.TopologyNode{}
				if err := r.client.Get(ctx, types.NamespacedName{
					Namespace: cr.GetNamespace(),
					Name:      strings.Join([]string{topologyName, nodeName}, ".")}, node); err != nil {
					if resource.IgnoreNotFound(err) != nil {
						return nil, err
					}
					r.log.Debug("sh-ep logical-link: node not found, delete the logical-link", "nodeName", nodeName)
					// node no longer exists, we can delete the logical element
					if err := r.hooks.Delete(ctx, cr); err != nil {
						return nil, err
					}
					// when delete is successfull we finish/return
					return nil, nil
				}
			}
		} else {
			// individual links
			node := &topov1alpha1.TopologyNode{}
			if err := r.client.Get(ctx, types.NamespacedName{
				Namespace: cr.GetNamespace(),
				Name:      strings.Join([]string{topologyName, nodeName}, ".")}, node); err != nil {
				r.log.Debug("individual link: node not found", "nodeName", nodeName)
				cr.SetStatus("down")
				cr.SetReason(fmt.Sprintf("node %d not found", i))
				return utils.StringPtr(fmt.Sprintf("node %d not found", i)), nil
			}
		}
	}
	return nil, nil
}
