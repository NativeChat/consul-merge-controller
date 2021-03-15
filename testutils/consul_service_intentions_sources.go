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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetConsulServiceIntentionsSource ...
func GetConsulServiceIntentionsSource(ctx context.Context, k8sClient client.Client, name string) (*v1alpha1.ConsulServiceIntentionsSource, error) {
	csis := new(v1alpha1.ConsulServiceIntentionsSource)
	exists, err := getK8sObject(ctx, k8sClient, name, csis)
	if !exists {
		csis = nil
	}

	return csis, err
}

// ExpectConsulServiceIntentionsSource ...
func ExpectConsulServiceIntentionsSource(ctx context.Context, k8sClient client.Client, name string, expected *consulk8s.SourceIntention) {
	csis, err := GetConsulServiceIntentionsSource(ctx, k8sClient, name)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	gomega.Expect(csis).NotTo(gomega.BeNil())
	gomega.Expect(csis.Finalizers).To(gomega.ContainElement(ServiceFinalizer))

	gomega.Expect(csis.Status.UpdatedAt).NotTo(gomega.BeEmpty())
	gomega.Expect(csis.Status.ContentSHA).To(gomega.Equal(getResourceContentSHA(csis)))

	gomega.Expect(csis.Spec.Source).To(gomega.Equal(expected))
}

// CreateConsulServiceIntentionsSource ...
func CreateConsulServiceIntentionsSource(ctx context.Context, k8sClient client.Client, serviceName string, source *consulk8s.SourceIntention) (string, error) {
	name := fmt.Sprintf("%s-%s-to-%s", source.Action, source.Name, serviceName)

	csis := &v1alpha1.ConsulServiceIntentionsSource{
		TypeMeta: v1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.Version,
			Kind:       "ConsulServiceIntentionsSource",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: DefaultK8sNamespace,

			Labels: map[string]string{ServiceIntentions: serviceName},
		},
		Spec: v1alpha1.ConsulServiceIntentionsSourceSpec{
			Source: source,
		},
	}

	err := k8sClient.Create(ctx, csis)
	if err != nil {
		return "", err
	}

	err = waitForConsulServiceIntentionsSourceToBeUpToDate(ctx, k8sClient, csis)
	if err != nil {
		return "", err
	}

	return name, nil
}

// CreateConsulServiceIntentionsSources ...
func CreateConsulServiceIntentionsSources(ctx context.Context, k8sClient client.Client, serviceName string, sources []*consulk8s.SourceIntention) []string {
	result := []string{}

	for _, source := range sources {
		name, err := CreateConsulServiceIntentionsSource(ctx, k8sClient, serviceName, source)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		result = append(result, name)
	}

	err := WaitForServiceIntentionsToBeCreated(ctx, k8sClient, serviceName)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return result
}

// UpdateConsulServiceIntentionsSource ...
func UpdateConsulServiceIntentionsSource(ctx context.Context, k8sClient client.Client, updated *v1alpha1.ConsulServiceIntentionsSource) error {
	err := k8sClient.Update(ctx, updated)
	if err != nil {
		return err
	}

	err = waitForConsulServiceIntentionsSourceToBeUpToDate(ctx, k8sClient, updated)

	return err
}

// DeleteConsulServiceIntentionsSources ...
func DeleteConsulServiceIntentionsSources(ctx context.Context, k8sClient client.Client, serviceName string) {
	requirement, err := labels.NewRequirement(ServiceIntentions, selection.Equals, []string{serviceName})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	sources := new(v1alpha1.ConsulServiceIntentionsSourceList)
	err = k8sClient.List(ctx, sources, &client.ListOptions{
		Namespace:     DefaultK8sNamespace,
		LabelSelector: labels.Everything().Add(*requirement),
	})

	for _, source := range sources.Items {
		err := DeleteConsulServiceIntentionsSource(ctx, k8sClient, source.Name)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	serviceIntentions, err := GetServiceIntentions(ctx, k8sClient, serviceName)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	if serviceIntentions == nil {
		return
	}

	err = deleteK8sObject(ctx, k8sClient, serviceName, serviceIntentions)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

// DeleteConsulServiceIntentionsSource ...
func DeleteConsulServiceIntentionsSource(ctx context.Context, k8sClient client.Client, name string) error {
	csr := new(v1alpha1.ConsulServiceIntentionsSource)
	err := deleteK8sObject(ctx, k8sClient, name, csr)

	return err
}

func waitForConsulServiceIntentionsSourceToBeUpToDate(ctx context.Context, k8sClient client.Client, expected *v1alpha1.ConsulServiceIntentionsSource) error {
	expectedSHA := getResourceContentSHA(expected)

	hasTimedOut := retryWithSleep(func() bool {
		existing, _ := GetConsulServiceIntentionsSource(ctx, k8sClient, expected.Name)
		result := existing.Status.ContentSHA == expectedSHA

		return result
	})

	if hasTimedOut {
		return fmt.Errorf("ConsulServiceIntentionsSource sync timeout exceeded")
	}

	return nil
}
