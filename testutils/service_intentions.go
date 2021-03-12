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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetServiceIntentions ...
func GetServiceIntentions(ctx context.Context, k8sClient client.Client, name string) (*consulk8s.ServiceIntentions, error) {
	serviceIntentions := new(consulk8s.ServiceIntentions)
	exists, err := getK8sObject(ctx, k8sClient, name, serviceIntentions)
	if !exists {
		serviceIntentions = nil
	}

	return serviceIntentions, err
}

// WaitForServiceIntentionsToBeCreated ...
func WaitForServiceIntentionsToBeCreated(ctx context.Context, k8sClient client.Client, name string) error {
	serviceIntentions := new(consulk8s.ServiceIntentions)

	hasTimedOut := retryWithSleep(func() bool {
		exists, _ := getK8sObject(ctx, k8sClient, name, serviceIntentions)
		if exists {
			return true
		}

		return false
	})

	if hasTimedOut {
		return fmt.Errorf("ServiceIntentions creation timeout exceeded")
	}

	return nil
}
