/*
Copyright 2021 Progress Software Corporation.

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

package reconcile

import (
	"context"
	"errors"
	"fmt"
	"time"

	e "github.com/NativeChat/consul-merge-controller/pkg/errors"
	"github.com/NativeChat/consul-merge-controller/pkg/services"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type reconciler struct {
	statusClient client.StatusClient
	crdService   services.CRDService
	merger       services.Merger
	log          logr.Logger
	queryLabel   string
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.log.Info("starting reconcile")

	obj, res, err := r.crdService.GetResourceFromRequest(ctx, req)
	if err != nil || res != nil {
		return *res, err
	}

	r.log = r.log.WithValues("resourceName", obj.GetName())

	reconcileAction := "triggered by dependency change"
	isChanged := r.crdService.IsChanged(obj)
	isDeleted := r.crdService.IsDeleted(obj)
	if r.crdService.IsNew(obj) {
		reconcileAction = "create"
	} else if isChanged {
		reconcileAction = "change"
	} else if isDeleted {
		reconcileAction = "delete"
	}

	r.log.Info(fmt.Sprintf("reconcile action is: %s", reconcileAction))

	queryValue, ok := obj.GetLabels()[r.queryLabel]
	if !ok || len(queryValue) == 0 {
		return ctrl.Result{}, apierrors.NewBadRequest(fmt.Sprintf("%s label is required", r.queryLabel))
	}

	namespace := req.Namespace
	resources, err := r.crdService.GetAllResourcesForService(ctx, r.queryLabel, queryValue, namespace)
	if err != nil {
		r.log.Error(err, "failed to get consul service routes")
		if errors.Is(err, e.ErrReconcile) {
			return ctrl.Result{Requeue: err.(*e.ReconcileError).ShouldRequeue}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	res, err = r.merger.Merge(ctx, queryValue, namespace, resources)
	if err != nil || res != nil {
		return *res, err
	}

	err = r.crdService.UpdateFinalizer(ctx, obj)
	if err != nil {
		r.log.Error(err, "failed to update the finalizer")

		return ctrl.Result{Requeue: true}, err
	}

	if !isDeleted && isChanged {
		r.crdService.SetContentSHA(obj, r.crdService.GetContentSHA(obj))
		r.crdService.SetUpdatedAt(obj, time.Now().String())

		r.log.Info("updating the status of the consul service route")
		err = r.statusClient.Status().Update(ctx, obj)
		if err != nil {
			r.log.Error(err, "failed to update the status of the consul service route")

			return ctrl.Result{Requeue: true}, err
		}

		r.log.Info("successfully updated the status of the consul service route")
	}

	return ctrl.Result{}, nil
}

// NewReconciler ...
func NewReconciler(
	statusClient client.StatusClient,
	crdService services.CRDService,
	merger services.Merger,
	log logr.Logger,
	queryLabel string,
) Reconciler {
	r := new(reconciler)
	r.statusClient = statusClient
	r.crdService = crdService
	r.merger = merger
	r.log = log
	r.queryLabel = queryLabel

	return r
}
