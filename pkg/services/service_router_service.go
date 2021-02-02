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
	"reflect"

	servicev1alpha1 "github.com/NativeChat/consul-merge-controller/apis/service/v1alpha1"
	"github.com/go-logr/logr"
	consulk8s "github.com/hashicorp/consul-k8s/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type serviceRouterService struct {
	reader client.Reader
	writer client.Writer
	log    logr.Logger
}

func (s *serviceRouterService) WriteServiceRouter(ctx context.Context, serviceRouterName, namespace string, consulServiceRoutes []servicev1alpha1.ConsulServiceRoute) (*ctrl.Result, error) {
	expectedConsulServiceRouter := s.getConsulServiceRouterDefinition(serviceRouterName, namespace, consulServiceRoutes)

	actualConsulServiceRouter := new(consulk8s.ServiceRouter)
	err := s.reader.Get(ctx, types.NamespacedName{Namespace: namespace, Name: serviceRouterName}, actualConsulServiceRouter)
	if err != nil {
		if !errors.IsNotFound(err) {
			s.log.Error(err, "failed to get service router")

			return &ctrl.Result{Requeue: true}, nil
		}

		s.log.Info("creating service router...")

		err = s.writer.Create(ctx, expectedConsulServiceRouter)
		if err != nil {
			s.log.Error(err, "failed to create service router")

			return &ctrl.Result{Requeue: true}, nil
		}

		s.log.Info("service router created")
		return nil, nil
	}

	if reflect.DeepEqual(expectedConsulServiceRouter.Spec, actualConsulServiceRouter.Spec) {
		s.log.Info("service router is up to date")

		return nil, nil
	}

	if len(expectedConsulServiceRouter.Spec.Routes) == 0 {
		s.log.Info("no routes left for service router, it will be deleted")

		err = s.writer.Delete(ctx, actualConsulServiceRouter)

		if err != nil {
			return &ctrl.Result{}, err
		}

		return nil, nil
	}

	s.log.Info("updating service router...")

	actualConsulServiceRouter.Spec = expectedConsulServiceRouter.Spec
	actualConsulServiceRouter.SetOwnerReferences(expectedConsulServiceRouter.GetOwnerReferences())
	err = s.writer.Update(ctx, actualConsulServiceRouter)
	if err != nil {
		s.log.Error(err, "failed to update service router")
	}

	s.log.Info("service router updated")

	return nil, nil
}

func (s *serviceRouterService) getConsulServiceRouterDefinition(serviceRouterName, namespace string, routes []servicev1alpha1.ConsulServiceRoute) *consulk8s.ServiceRouter {
	serviceRouter := &consulk8s.ServiceRouter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceRouterName,
			Namespace: namespace,
		},
	}

	for _, serviceRoute := range routes {
		serviceRouter.Spec.Routes = append(serviceRouter.Spec.Routes, serviceRoute.Spec.Route)
		serviceRouter.SetOwnerReferences(append(serviceRoute.GetOwnerReferences(), metav1.OwnerReference{
			APIVersion: serviceRoute.APIVersion,
			Kind:       serviceRoute.Kind,
			Name:       serviceRoute.Name,
			UID:        serviceRoute.UID,
		}))
	}

	return serviceRouter
}

// NewServiceRouterService returns new service router service implementation.
func NewServiceRouterService(reader client.Reader, writer client.Writer, log logr.Logger) ServiceRouterService {
	svc := new(serviceRouterService)
	svc.reader = reader
	svc.writer = writer
	svc.log = log

	return svc
}
