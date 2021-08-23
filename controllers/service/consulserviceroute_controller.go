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
	"reflect"

	"github.com/go-logr/logr"
	consulk8s "github.com/hashicorp/consul-k8s/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/NativeChat/consul-merge-controller/apis/service/v1alpha1"
	servicev1alpha1 "github.com/NativeChat/consul-merge-controller/apis/service/v1alpha1"
	"github.com/NativeChat/consul-merge-controller/pkg/finalizers"
	controllerlabels "github.com/NativeChat/consul-merge-controller/pkg/labels"
	"github.com/NativeChat/consul-merge-controller/pkg/reconcile"
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
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.0/pkg/reconcile
func (r *ConsulServiceRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("consulserviceroute", req.NamespacedName)

	crdService := services.NewCRDService(
		r.Client,
		r.Client,
		log,
		finalizers.ConsulServiceRouteFinalizerName,
		reflect.TypeOf(v1alpha1.ConsulServiceRoute{}),
		reflect.TypeOf(v1alpha1.ConsulServiceRouteList{}),
	)
	merger := services.NewMerger(
		r.Client,
		r.Client,
		log,
		nil,
		"Routes",
		"Route",
		reflect.TypeOf(consulk8s.ServiceRouter{}),
	)
	reconciler := reconcile.NewReconciler(
		r,
		crdService,
		merger,
		log,
		controllerlabels.ServiceRouter,
	)

	res, err := reconciler.Reconcile(ctx, req)

	return res, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConsulServiceRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&servicev1alpha1.ConsulServiceRoute{}).
		Owns(&consulk8s.ServiceRouter{}).
		Complete(r)
}
