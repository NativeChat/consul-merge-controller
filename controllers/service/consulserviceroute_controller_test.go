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

package service_test

import (
	"context"
	"fmt"

	"github.com/NativeChat/consul-merge-controller/testutils"
	consulk8s "github.com/hashicorp/consul-k8s/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var serviceA = "service-a"
var serviceAV1 = fmt.Sprintf("%s-v1", serviceA)
var serviceAV2 = fmt.Sprintf("%s-v2", serviceA)
var serviceAV3 = fmt.Sprintf("%s-v3", serviceA)

var serviceB = "service-b"
var serviceBV1 = fmt.Sprintf("%s-v1", serviceB)
var serviceBV2 = fmt.Sprintf("%s-v2", serviceB)

var _ = Describe("Service", func() {
	var ctx context.Context
	BeforeEach(func() {
		ctx = context.Background()

		err := testutils.CreateServiceDefaults(ctx, k8sClient, serviceA)
		Expect(err).NotTo(HaveOccurred())

		err = testutils.CreateServiceDefaults(ctx, k8sClient, serviceB)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		for _, serviceName := range []string{serviceA, serviceB} {
			testutils.DeleteServiceDefaults(ctx, k8sClient, serviceName)
		}

		for _, csrName := range []string{serviceAV1, serviceAV2, serviceAV3, serviceBV1, serviceBV2} {
			testutils.DeleteConsulServiceRoute(ctx, k8sClient, csrName)
		}
	})

	Context("Merge", func() {
		It("should merge multiple services correctly", func() {
			serviceARoutes := []consulk8s.ServiceRoute{
				testutils.CreateHTTPPathPrefixRoute(serviceAV1, "/v1"),
				testutils.CreateHTTPPathPrefixRoute(serviceAV2, "/v2"),
			}

			serviceBRoutes := []consulk8s.ServiceRoute{
				testutils.CreateHTTPPathPrefixRoute(serviceBV1, "/v1"),
				testutils.CreateHTTPPathPrefixRoute(serviceBV2, "/v2"),
			}

			serviceRouters := map[string][]consulk8s.ServiceRoute{
				serviceA: serviceARoutes,
				serviceB: serviceBRoutes,
			}

			for serviceRouterName, serviceRoutes := range serviceRouters {
				for _, serviceRoute := range serviceRoutes {
					err := testutils.CreateConsulServiceRoute(ctx, k8sClient, serviceRouterName, serviceRoute)
					Expect(err).NotTo(HaveOccurred())

					testutils.ExpectConsulServiceRoute(ctx, k8sClient, serviceRoute.Destination.Service, serviceRoute)
				}

				serviceRouter, err := testutils.GetServiceRouter(ctx, k8sClient, serviceRouterName)
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceRouter).NotTo(BeNil())

				Expect(serviceRouter.Spec.Routes).To(ContainElements(serviceRoutes))
			}
		})
	})
})
