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

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	consulk8s "github.com/hashicorp/consul-k8s/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	servicev1alpha1 "github.com/NativeChat/consul-merge-controller/apis/service/v1alpha1"
	e "github.com/NativeChat/consul-merge-controller/pkg/errors"
	controllerlabels "github.com/NativeChat/consul-merge-controller/pkg/labels"
	"github.com/NativeChat/consul-merge-controller/pkg/services"
)

// ConsulServiceRouteReconciler reconciles a ConsulServiceRoute object
type ConsulServiceRouteReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=service.consul.k8s.nativechat.com,resources=consulserviceroutes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=service.consul.k8s.nativechat.com,resources=consulserviceroutes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=service.consul.k8s.nativechat.com,resources=consulserviceroutes/finalizers,verbs=update

// +kubebuilder:rbac:groups=consul.hashicorp.com,resources=servicerouters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=consul.hashicorp.com,resources=servicerouters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=consul.hashicorp.com,resources=servicerouters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ConsulServiceRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("consulserviceroute", req.NamespacedName)

	log.Info("starting reconcile")

	consulServiceRouteService := services.NewConsulServiceRouteService(r, r, log)
	serviceRouterService := services.NewServiceRouterService(r, r, log)

	consulServiceRoute, res, err := consulServiceRouteService.GetConsulServiceRouteFromReq(ctx, req)
	if err != nil || res != nil {
		return *res, err
	}

	reconcileAction := "triggered by dependency change"
	isChanged := consulServiceRouteService.IsChanged(*consulServiceRoute)
	isDeleted := consulServiceRouteService.IsDeleted(*consulServiceRoute)
	if consulServiceRouteService.IsNew(*consulServiceRoute) {
		reconcileAction = "create"
	} else if isChanged {
		reconcileAction = "change"
	} else if isDeleted {
		reconcileAction = "delete"
	}

	log.Info(fmt.Sprintf("reconcile action is: %s", reconcileAction))

	serviceRouterName, ok := consulServiceRoute.Labels[controllerlabels.ServiceRouter]
	if !ok || len(serviceRouterName) == 0 {
		return ctrl.Result{}, apierrors.NewBadRequest(fmt.Sprintf("%s label is required", controllerlabels.ServiceRouter))
	}

	namespace := req.Namespace
	consulServiceRoutes, err := consulServiceRouteService.GetServiceRoutesForServiceRouter(ctx, serviceRouterName, namespace)
	if err != nil {
		log.Error(err, "failed to get consul service routes")
		if errors.Is(err, e.ErrReconcile) {
			return ctrl.Result{Requeue: err.(*e.ReconcileError).ShouldRequeue}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	res, err = serviceRouterService.WriteServiceRouter(ctx, serviceRouterName, namespace, consulServiceRoutes)
	if err != nil || res != nil {
		return *res, err
	}

	err = consulServiceRouteService.UpdateFinalizer(ctx, consulServiceRoute)
	if err != nil {
		log.Error(err, "failed to update the finalizer for the consul service route")

		return ctrl.Result{Requeue: true}, err
	}

	if !isDeleted && isChanged {
		consulServiceRoute.Status.UpdatedAt = time.Now().String()
		consulServiceRoute.Status.ContentSHA = consulServiceRouteService.GetContentSHA(*consulServiceRoute)

		log.Info("updating the status of the consul service route")
		err = r.Status().Update(ctx, consulServiceRoute)
		if err != nil {
			log.Error(err, "failed to update the status of the consul service route")

			return ctrl.Result{Requeue: true}, err
		}

		log.Info("successfully updated the status of the consul service route")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConsulServiceRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&servicev1alpha1.ConsulServiceRoute{}).
		Owns(&consulk8s.ServiceRouter{}).
		Complete(r)
}
