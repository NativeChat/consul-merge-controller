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

	consulk8s "github.com/hashicorp/consul-k8s/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateServiceDefaults ...
func CreateServiceDefaults(ctx context.Context, k8sClient client.Client, service string) error {
	serviceProtocol := "http"

	sd := &consulk8s.ServiceDefaults{
		TypeMeta: v1.TypeMeta{
			APIVersion: consulk8s.SchemeBuilder.GroupVersion.Version,
			Kind:       "ServiceDefaults",
		},
		ObjectMeta: v1.ObjectMeta{Name: service, Namespace: DefaultK8sNamespace},
		Spec:       consulk8s.ServiceDefaultsSpec{Protocol: serviceProtocol},
	}

	err := k8sClient.Create(ctx, sd)
	if err != nil {
		return err
	}

	hasTimedOut := retryWithSleep(func() bool {
		sd := new(consulk8s.ServiceDefaults)
		_, err = getK8sObject(ctx, k8sClient, service, sd)
		isSynced := len(sd.Status.Conditions) > 0 && sd.Status.Conditions[0].Status == "True"

		return isSynced
	})

	if hasTimedOut {
		return fmt.Errorf("create service defaults timeout exceeded")
	}

	return nil
}

// DeleteServiceDefaults ...
func DeleteServiceDefaults(ctx context.Context, k8sClient client.Client, service string) error {
	sd := new(consulk8s.ServiceDefaults)
	err := deleteK8sObject(ctx, k8sClient, service, sd)

	return err
}
