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

package testutils

import (
	"context"
	"fmt"

	"github.com/onsi/gomega"

	"github.com/NativeChat/consul-merge-controller/apis/service/v1alpha1"
	consulk8s "github.com/hashicorp/consul-k8s/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateConsulServiceRoute ...
func CreateConsulServiceRoute(ctx context.Context, k8sClient client.Client, serviceRouter string, route consulk8s.ServiceRoute) error {
	name := route.Destination.Service

	csr := &v1alpha1.ConsulServiceRoute{
		TypeMeta: v1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.Version,
			Kind:       "ConsulServiceRoute",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: DefaultK8sNamespace,

			Labels: map[string]string{ServiceRouterLabel: serviceRouter},
		},
		Spec: v1alpha1.ConsulServiceRouteSpec{
			Route: route,
		},
	}

	err := k8sClient.Create(ctx, csr)
	if err != nil {
		return err
	}

	err = waitForConsulServiceRouteToBeUpToDate(ctx, k8sClient, csr)

	return err
}

// GetConsulServiceRoute ...
func GetConsulServiceRoute(ctx context.Context, k8sClient client.Client, service string) (*v1alpha1.ConsulServiceRoute, error) {
	csr := new(v1alpha1.ConsulServiceRoute)
	exists, err := getK8sObject(ctx, k8sClient, service, csr)
	if !exists {
		csr = nil
	}

	return csr, err
}

// DeleteConsulServiceRoute ...
func DeleteConsulServiceRoute(ctx context.Context, k8sClient client.Client, name string) error {
	csr := new(v1alpha1.ConsulServiceRoute)
	err := deleteK8sObject(ctx, k8sClient, name, csr)

	return err
}

// UpdateConsulServiceRoute ...
func UpdateConsulServiceRoute(ctx context.Context, k8sClient client.Client, updated *v1alpha1.ConsulServiceRoute) error {
	err := k8sClient.Update(ctx, updated)
	if err != nil {
		return err
	}

	err = waitForConsulServiceRouteToBeUpToDate(ctx, k8sClient, updated)

	return err
}

// ExpectConsulServiceRoute ...
func ExpectConsulServiceRoute(ctx context.Context, k8sClient client.Client, name string, expected consulk8s.ServiceRoute) {
	csr, err := GetConsulServiceRoute(ctx, k8sClient, name)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	gomega.Expect(csr).NotTo(gomega.BeNil())
	gomega.Expect(csr.Finalizers).To(gomega.ContainElement(ServiceFinalizer))

	gomega.Expect(csr.Status.UpdatedAt).NotTo(gomega.BeEmpty())
	gomega.Expect(csr.Status.ContentSHA).To(gomega.Equal(getResourceContentSHA(csr)))

	gomega.Expect(csr.Spec.Route).To(gomega.Equal(expected))
}

// CreateConsulServiceRoutes ...
func CreateConsulServiceRoutes(ctx context.Context, k8sClient client.Client, serviceRouterName string, routes []consulk8s.ServiceRoute) {
	for _, serviceRoute := range routes {
		err := CreateConsulServiceRoute(ctx, k8sClient, serviceRouterName, serviceRoute)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	err := WaitForServiceRouterToBeCreated(ctx, k8sClient, serviceRouterName)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

// DeleteConsulServiceRoutes ...
func DeleteConsulServiceRoutes(ctx context.Context, k8sClient client.Client, serviceRouterName string, routes []consulk8s.ServiceRoute) {
	for _, serviceRoute := range routes {
		err := DeleteConsulServiceRoute(ctx, k8sClient, serviceRoute.Destination.Service)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	serviceRouter, err := GetServiceRouter(ctx, k8sClient, serviceRouterName)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	if serviceRouter == nil {
		return
	}

	err = deleteK8sObject(ctx, k8sClient, serviceRouterName, serviceRouter)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func waitForConsulServiceRouteToBeUpToDate(ctx context.Context, k8sClient client.Client, expected *v1alpha1.ConsulServiceRoute) error {
	expectedSHA := getResourceContentSHA(expected)

	hasTimedOut := retryWithSleep(func() bool {
		existing, _ := GetConsulServiceRoute(ctx, k8sClient, expected.Name)
		if existing.Status.ContentSHA == expectedSHA {
			return true
		}

		return false
	})

	if hasTimedOut {
		return fmt.Errorf("ConsulServiceRoute sync timeout exceeded")
	}

	return nil
}
