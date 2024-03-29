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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceRouterService provides methods for working with service routers.
type ServiceRouterService interface {
	WriteServiceRouter(ctx context.Context, serviceRouterName, namespace string, consulServiceRoutes []servicev1alpha1.ConsulServiceRoute) (*ctrl.Result, error)
}

// CRDService provides methods for working with custom resources.
type CRDService interface {
	GetResourceFromRequest(ctx context.Context, req ctrl.Request) (client.Object, *ctrl.Result, error)
	GetAllResourcesForService(ctx context.Context, label, serviceName, namespace string) ([]client.Object, error)
	UpdateFinalizer(ctx context.Context, obj client.Object) error
	IsDeleted(obj client.Object) bool
	IsNew(obj client.Object) bool
	IsChanged(obj client.Object) bool
	GetContentSHA(obj client.Object) string
	SetUpdatedAt(obj client.Object, updatedAt string)
	SetContentSHA(obj client.Object, contentSHA string)
}

// Merger provides methods for merging items into a destination property of a k8s object.
type Merger interface {
	Merge(ctx context.Context, destinationResourceName, namespace string, items []client.Object) (*ctrl.Result, error)
}
