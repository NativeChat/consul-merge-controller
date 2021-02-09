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
	"time"

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

var serviceDefaults = []string{
	serviceA, serviceB, serviceAV1, serviceAV2, serviceAV3, serviceBV1, serviceBV2,
}

var _ = Describe("Service", func() {
	Describe("ConsulServiceRoute controller", func() {
		var ctx context.Context

		BeforeEach(func() {
			ctx = context.Background()

			for _, serviceDefaultsName := range serviceDefaults {
				err := testutils.CreateServiceDefaults(ctx, k8sClient, serviceDefaultsName)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		AfterEach(func() {
			for _, csrName := range []string{serviceAV1, serviceAV2, serviceAV3, serviceBV1, serviceBV2} {
				err := testutils.DeleteConsulServiceRoute(ctx, k8sClient, csrName)
				Expect(err).NotTo(HaveOccurred())
			}

			for _, serviceDefaultsName := range serviceDefaults {
				err := testutils.DeleteServiceDefaults(ctx, k8sClient, serviceDefaultsName)
				Expect(err).NotTo(HaveOccurred())
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

					err := testutils.WaitForServiceRouterToBeCreated(ctx, k8sClient, serviceRouterName)
					Expect(err).NotTo(HaveOccurred())

					serviceRouter, err := testutils.GetServiceRouter(ctx, k8sClient, serviceRouterName)
					Expect(err).NotTo(HaveOccurred())
					Expect(serviceRouter).NotTo(BeNil())

					Expect(serviceRouter.Spec.Routes).To(ContainElements(serviceRoutes))
				}
			})
		})

		Context("Multiple routes in single service router", func() {
			var serviceRoutes []consulk8s.ServiceRoute

			BeforeEach(func() {
				serviceRoutes = []consulk8s.ServiceRoute{
					testutils.CreateHTTPPathPrefixRoute(serviceAV1, "/v1"),
					testutils.CreateHTTPPathPrefixRoute(serviceAV2, "/v2"),
					testutils.CreateHTTPPathPrefixRoute(serviceAV3, "/v3"),
				}

				testutils.CreateConsulServiceRoutes(ctx, k8sClient, serviceA, serviceRoutes)
			})

			AfterEach(func() {
				testutils.DeleteConsulServiceRoutes(ctx, k8sClient, serviceA, serviceRoutes)
			})

			Describe("should update the correct item in the service router", func() {
				updateTestCases := []struct {
					name          string
					indexToUpdate int
				}{
					{name: "when the first item is updated", indexToUpdate: 0},
					{name: "when the second item is updated", indexToUpdate: 1},
					{name: "when the last item is updated", indexToUpdate: 2},
				}

				for _, testCase := range updateTestCases {
					It(testCase.name, func() {
						serviceRouter, err := testutils.GetServiceRouter(ctx, k8sClient, serviceA)
						Expect(err).NotTo(HaveOccurred())

						routes := serviceRouter.Spec.Routes

						updated, err := testutils.GetConsulServiceRoute(ctx, k8sClient, routes[testCase.indexToUpdate].Destination.Service)
						Expect(err).NotTo(HaveOccurred())

						updated.Spec.Route.Match.HTTP.PathPrefix = "/updated"

						err = testutils.UpdateConsulServiceRoute(ctx, k8sClient, updated)
						Expect(err).NotTo(HaveOccurred())

						serviceRouter, err = testutils.GetServiceRouter(ctx, k8sClient, serviceA)
						Expect(err).NotTo(HaveOccurred())

						expectedRoutes := []consulk8s.ServiceRoute{updated.Spec.Route}
						for i, route := range routes {
							if i != testCase.indexToUpdate {
								expectedRoutes = append(expectedRoutes, route)
							}
						}

						Expect(serviceRouter.Spec.Routes).To(ContainElements(expectedRoutes))
					})
				}
			})

			Describe("should delete the correct item from the service router", func() {
				deleteTestCases := []struct {
					name          string
					indexToDelete int
				}{
					{name: "when the first item is deleted", indexToDelete: 0},
					{name: "when the second item is deleted", indexToDelete: 1},
					{name: "when the last item is deleted", indexToDelete: 2},
				}

				for _, testCase := range deleteTestCases {
					It(testCase.name, func() {
						serviceRouter, err := testutils.GetServiceRouter(ctx, k8sClient, serviceA)
						Expect(err).NotTo(HaveOccurred())

						routes := serviceRouter.Spec.Routes

						toDelete, err := testutils.GetConsulServiceRoute(ctx, k8sClient, routes[testCase.indexToDelete].Destination.Service)
						Expect(err).NotTo(HaveOccurred())

						err = testutils.DeleteConsulServiceRoute(ctx, k8sClient, toDelete.Name)
						Expect(err).NotTo(HaveOccurred())

						serviceRouter, err = testutils.GetServiceRouter(ctx, k8sClient, serviceA)
						Expect(err).NotTo(HaveOccurred())

						expectedRoutes := []consulk8s.ServiceRoute{}
						for i, route := range routes {
							if i != testCase.indexToDelete {
								expectedRoutes = append(expectedRoutes, route)
							}
						}

						Expect(serviceRouter.Spec.Routes).To(ContainElements(expectedRoutes))
						Expect(serviceRouter.Spec.Routes).NotTo(ContainElement(toDelete.Spec.Route))
						Expect(serviceRouter.Spec.Routes).To(HaveLen(len(routes) - 1))
					})
				}
			})
		})

		It("should delete the service router if all routes for it are deleted.", func() {
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
				}
			}

			serviceAServiceRouter, err := testutils.GetServiceRouter(ctx, k8sClient, serviceA)
			Expect(err).NotTo(HaveOccurred())

			for _, route := range serviceAServiceRouter.Spec.Routes {
				err = testutils.DeleteConsulServiceRoute(ctx, k8sClient, route.Destination.Service)
				Expect(err).NotTo(HaveOccurred())
			}

			time.Sleep(time.Second)

			serviceAServiceRouter, err = testutils.GetServiceRouter(ctx, k8sClient, serviceA)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceAServiceRouter).To(BeNil())

			serviceBServiceRouter, err := testutils.GetServiceRouter(ctx, k8sClient, serviceB)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceBServiceRouter).NotTo(BeNil())
		})
	})
})
