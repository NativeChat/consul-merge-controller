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
	"github.com/NativeChat/consul-merge-controller/pkg/finalizers"
	controllerlabels "github.com/NativeChat/consul-merge-controller/pkg/labels"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type consulServiceRouteService struct {
	reader client.Reader
	writer client.Writer
	log    logr.Logger
}

func (c *consulServiceRouteService) GetConsulServiceRouteFromReq(ctx context.Context, req ctrl.Request) (*servicev1alpha1.ConsulServiceRoute, *ctrl.Result, error) {
	consulServiceRoute := new(servicev1alpha1.ConsulServiceRoute)
	err := c.reader.Get(ctx, req.NamespacedName, consulServiceRoute)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			// Return and don't requeue.
			c.log.Info("consul service route resource not found, object must be deleted")
			return nil, &ctrl.Result{}, nil
		}

		c.log.Error(err, "failed to get consul service route")

		return nil, &ctrl.Result{}, err
	}

	return consulServiceRoute, nil, nil
}

func (c *consulServiceRouteService) GetServiceRoutesForServiceRouter(ctx context.Context, serviceRouterName, namespace string) ([]servicev1alpha1.ConsulServiceRoute, error) {
	consulServiceRoutes := new(servicev1alpha1.ConsulServiceRouteList)

	requirement, err := labels.NewRequirement(controllerlabels.ServiceRouter, selection.Equals, []string{serviceRouterName})
	if err != nil {
		panic(err)
	}
	err = c.reader.List(ctx, consulServiceRoutes, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.Everything().Add(*requirement),
	})

	if err != nil {
		return nil, err
	}

	notMarkedForDeletion := []servicev1alpha1.ConsulServiceRoute{}
	for _, item := range consulServiceRoutes.Items {
		if !c.isDeleted(item) {
			notMarkedForDeletion = append(notMarkedForDeletion, item)
		}
	}

	return notMarkedForDeletion, nil
}

func (c *consulServiceRouteService) UpdateFinalizer(ctx context.Context, serviceRoute *servicev1alpha1.ConsulServiceRoute) error {
	if c.isDeleted(*serviceRoute) {
		controllerutil.RemoveFinalizer(serviceRoute, finalizers.ConsulServiceRouteFinalizerName)
	} else {
		controllerutil.AddFinalizer(serviceRoute, finalizers.ConsulServiceRouteFinalizerName)
	}

	err := c.writer.Update(ctx, serviceRoute)

	return err
}

func (c *consulServiceRouteService) isDeleted(serviceRoute servicev1alpha1.ConsulServiceRoute) bool {
	isDeleted := serviceRoute.GetDeletionTimestamp() != nil

	return isDeleted
}

// NewConsulServiceRouteService returns new consul service route service implementation.
func NewConsulServiceRouteService(reader client.Reader, writer client.Writer, log logr.Logger) ConsulServiceRouteService {
	svc := new(consulServiceRouteService)
	svc.reader = reader
	svc.writer = writer
	svc.log = log

	return svc
}
