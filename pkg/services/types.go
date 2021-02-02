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

package services

import (
	"context"

	servicev1alpha1 "github.com/NativeChat/consul-merge-controller/apis/service/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ConsulServiceRouteService provides methods for working with consul service routes.
type ConsulServiceRouteService interface {
	GetConsulServiceRouteFromReq(ctx context.Context, req ctrl.Request) (*servicev1alpha1.ConsulServiceRoute, *ctrl.Result, error)
	GetServiceRoutesForServiceRouter(ctx context.Context, serviceRouterName, namespace string) ([]servicev1alpha1.ConsulServiceRoute, error)
	UpdateFinalizer(ctx context.Context, serviceRoute *servicev1alpha1.ConsulServiceRoute) error
}

// ServiceRouterService provides methods for working with service routers.
type ServiceRouterService interface {
	WriteServiceRouter(ctx context.Context, serviceRouterName, namespace string, consulServiceRoutes []servicev1alpha1.ConsulServiceRoute) (*ctrl.Result, error)
}
