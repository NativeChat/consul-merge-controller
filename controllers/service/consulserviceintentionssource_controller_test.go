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
	"time"

	"github.com/NativeChat/consul-merge-controller/testutils"
	consulk8s "github.com/hashicorp/consul-k8s/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConsulServiceIntentionsSource controller", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		testutils.DeleteConsulServiceIntentionsSources(ctx, k8sClient, serviceA)
		testutils.DeleteConsulServiceIntentionsSources(ctx, k8sClient, serviceAV1)
		testutils.DeleteConsulServiceIntentionsSources(ctx, k8sClient, serviceAV2)
		testutils.DeleteConsulServiceIntentionsSources(ctx, k8sClient, serviceAV3)
		testutils.DeleteConsulServiceIntentionsSources(ctx, k8sClient, serviceB)
		testutils.DeleteConsulServiceIntentionsSources(ctx, k8sClient, serviceBV1)
		testutils.DeleteConsulServiceIntentionsSources(ctx, k8sClient, serviceBV2)
	})

	Context("Merge", func() {
		It("should merge multiple service intentions correctly", func() {
			serviceASources := []*consulk8s.SourceIntention{
				{Name: serviceBV1, Action: "allow"},
				{Name: serviceBV2, Action: "allow"},
			}
			serviceBSources := []*consulk8s.SourceIntention{
				{Name: serviceAV1, Action: "allow"},
				{Name: serviceAV2, Action: "allow"},
			}

			services := map[string][]*consulk8s.SourceIntention{
				serviceA: serviceASources,
				serviceB: serviceBSources,
			}

			for serviceName, intentionsSources := range services {
				for _, source := range intentionsSources {
					name, err := testutils.CreateConsulServiceIntentionsSource(ctx, k8sClient, serviceName, source)
					Expect(err).NotTo(HaveOccurred())

					testutils.ExpectConsulServiceIntentionsSource(ctx, k8sClient, name, source)
				}

				err := testutils.WaitForServiceIntentionsToBeCreated(ctx, k8sClient, serviceName)
				Expect(err).NotTo(HaveOccurred())

				serviceIntentions, err := testutils.GetServiceIntentions(ctx, k8sClient, serviceName)
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceIntentions).NotTo(BeNil())

				Expect(serviceIntentions.Spec.Sources).To(ContainElements(intentionsSources))
			}
		})
	})

	Context("Multiple sources in single service intentions", func() {
		var serviceIntentionsSources []*consulk8s.SourceIntention
		var serviceIntentionsNames []string

		BeforeEach(func() {
			serviceIntentionsSources = []*consulk8s.SourceIntention{
				{Name: serviceB, Action: "allow"},
				{Name: serviceBV1, Action: "allow"},
				{Name: serviceBV2, Action: "allow"},
			}

			serviceIntentionsNames = testutils.CreateConsulServiceIntentionsSources(ctx, k8sClient, serviceA, serviceIntentionsSources)
		})

		Describe("should update the correct item in the service intentions", func() {
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
					serviceIntentions, err := testutils.GetServiceIntentions(ctx, k8sClient, serviceA)
					Expect(err).NotTo(HaveOccurred())

					sources := serviceIntentions.Spec.Sources

					updated, err := testutils.GetConsulServiceIntentionsSource(ctx, k8sClient, serviceIntentionsNames[testCase.indexToUpdate])
					Expect(err).NotTo(HaveOccurred())

					updated.Spec.Source.Action = "deny"

					err = testutils.UpdateConsulServiceIntentionsSource(ctx, k8sClient, updated)
					Expect(err).NotTo(HaveOccurred())

					serviceIntentions, err = testutils.GetServiceIntentions(ctx, k8sClient, serviceA)
					Expect(err).NotTo(HaveOccurred())

					expected := []*consulk8s.SourceIntention{updated.Spec.Source}
					for i, source := range sources {
						if i != testCase.indexToUpdate {
							expected = append(expected, source)
						}
					}

					Expect(serviceIntentions.Spec.Sources).To(ContainElements(expected))
				})
			}
		})

		Describe("should delete the correct item from the service intentions", func() {
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
					serviceIntentions, err := testutils.GetServiceIntentions(ctx, k8sClient, serviceA)
					Expect(err).NotTo(HaveOccurred())

					sources := serviceIntentions.Spec.Sources

					toDelete, err := testutils.GetConsulServiceIntentionsSource(ctx, k8sClient, serviceIntentionsNames[testCase.indexToDelete])
					Expect(err).NotTo(HaveOccurred())

					err = testutils.DeleteConsulServiceIntentionsSource(ctx, k8sClient, toDelete.Name)
					Expect(err).NotTo(HaveOccurred())

					serviceIntentions, err = testutils.GetServiceIntentions(ctx, k8sClient, serviceA)
					Expect(err).NotTo(HaveOccurred())

					expected := []*consulk8s.SourceIntention{}
					for i, source := range sources {
						if i != testCase.indexToDelete {
							expected = append(expected, source)
						}
					}

					Expect(serviceIntentions.Spec.Sources).To(ContainElements(expected))
					Expect(serviceIntentions.Spec.Sources).NotTo(ContainElement(toDelete.Spec.Source))
					Expect(serviceIntentions.Spec.Sources).To(HaveLen(len(sources) - 1))
				})
			}
		})
	})

	It("should delete the service router if all routes for it are deleted.", func() {
		serviceASources := []*consulk8s.SourceIntention{
			{Name: serviceBV1, Action: "allow"},
			{Name: serviceBV2, Action: "allow"},
		}
		serviceBSources := []*consulk8s.SourceIntention{
			{Name: serviceAV1, Action: "allow"},
			{Name: serviceAV2, Action: "allow"},
		}

		services := map[string][]*consulk8s.SourceIntention{
			serviceA: serviceASources,
			serviceB: serviceBSources,
		}

		for serviceName, intentionsSources := range services {
			for _, source := range intentionsSources {
				name, err := testutils.CreateConsulServiceIntentionsSource(ctx, k8sClient, serviceName, source)
				Expect(err).NotTo(HaveOccurred())

				testutils.ExpectConsulServiceIntentionsSource(ctx, k8sClient, name, source)
			}
		}

		testutils.DeleteConsulServiceIntentionsSources(ctx, k8sClient, serviceA)

		time.Sleep(time.Second)

		serviceAServiceIntentions, err := testutils.GetServiceIntentions(ctx, k8sClient, serviceA)
		Expect(err).NotTo(HaveOccurred())
		Expect(serviceAServiceIntentions).To(BeNil())

		serviceBServiceIntentions, err := testutils.GetServiceIntentions(ctx, k8sClient, serviceB)
		Expect(err).NotTo(HaveOccurred())
		Expect(serviceBServiceIntentions).NotTo(BeNil())
	})
})
