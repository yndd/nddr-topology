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

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	//nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/resource"

	topov1alpha1 "github.com/yndd/nddr-topology/apis/topo/v1alpha1"
)

const (
	errApplyLink  = "cannot apply link"
	errGetLink    = "cannot get link"
	errDeleteLink = "cannot delete link"
)

// A Hooks performs operations to desploy/destroy .
type Hooks interface {
	// Get performs operations to validate the child resources
	Get(ctx context.Context, cr topov1alpha1.Tl) (*topov1alpha1.TopologyLink, error)

	// Create performs operations to deploy the child resources
	Create(ctx context.Context, cr topov1alpha1.Tl) error

	// Delete performs operations to deploy the child resources
	Delete(ctx context.Context, cr topov1alpha1.Tl) error

	// apply performs operations to update the resource
	Apply(ctx context.Context, cr topov1alpha1.Tl, mhtl *topov1alpha1.TopologyLink) error

	// apply performs operations to update the resource
	DeleteApply(ctx context.Context, cr topov1alpha1.Tl, mhtl *topov1alpha1.TopologyLink) error
}

// DeviceDriverHooks performs operations to deploy the device driver.
type Hook struct {
	client resource.ClientApplicator
	log    logging.Logger
}

func NewHook(client resource.ClientApplicator, log logging.Logger) Hooks {
	return &Hook{
		client: client,
		log:    log,
	}
}

func (h *Hook) Get(ctx context.Context, cr topov1alpha1.Tl) (*topov1alpha1.TopologyLink, error) {
	link := buildLogicalTopologyLink(cr)
	h.log.Debug("hook get", "logical link name", link.GetName())
	if err := h.client.Get(ctx, types.NamespacedName{Namespace: cr.GetNamespace(), Name: link.GetName()}, link); err != nil {
		return nil, errors.Wrap(err, errGetLink)
	}
	return link, nil
}

func (h *Hook) Create(ctx context.Context, cr topov1alpha1.Tl) error {
	link := buildLogicalTopologyLink(cr)
	if err := h.client.Apply(ctx, link); err != nil {
		return errors.Wrap(err, errApplyLink)
	}
	return nil
}

func (h *Hook) Delete(ctx context.Context, cr topov1alpha1.Tl) error {
	link := buildLogicalTopologyLink(cr)
	if err := h.client.Delete(ctx, link); err != nil {
		return errors.Wrap(err, errDeleteLink)
	}

	return nil
}

func (h *Hook) Apply(ctx context.Context, cr topov1alpha1.Tl, mhtl *topov1alpha1.TopologyLink) error {
	tl := updateLogicalTopologyLink(cr, mhtl)
	if err := h.client.Apply(ctx, tl); err != nil {
		return errors.Wrap(err, errDeleteLink)
	}

	return nil
}

func (h *Hook) DeleteApply(ctx context.Context, cr topov1alpha1.Tl, mhtl *topov1alpha1.TopologyLink) error {
	tl := updateDeleteLogicalTopologyLink(cr, mhtl)
	if err := h.client.Apply(ctx, tl); err != nil {
		return errors.Wrap(err, errDeleteLink)
	}

	return nil
}
