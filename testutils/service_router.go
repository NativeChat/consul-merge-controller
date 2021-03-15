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

// GetServiceRouter ...
func GetServiceRouter(ctx context.Context, k8sClient client.Client, name string) (*consulk8s.ServiceRouter, error) {
	sr := new(consulk8s.ServiceRouter)
	exists, err := getK8sObject(ctx, k8sClient, name, sr)
	if !exists {
		sr = nil
	}

	return sr, err
}

// WaitForServiceRouterToBeCreated ...
func WaitForServiceRouterToBeCreated(ctx context.Context, k8sClient client.Client, name string) error {
	serviceRouter := new(consulk8s.ServiceRouter)

	hasTimedOut := retryWithSleep(func() bool {
		exists, _ := getK8sObject(ctx, k8sClient, name, serviceRouter)

		return exists
	})

	if hasTimedOut {
		return fmt.Errorf("ServiceRouter creation timeout exceeded")
	}

	return nil
}

// CreateHTTPPathPrefixRoute ...
func CreateHTTPPathPrefixRoute(service, pathPrefix string) consulk8s.ServiceRoute {
	result := consulk8s.ServiceRoute{
		Match:       &consulk8s.ServiceRouteMatch{HTTP: &consulk8s.ServiceRouteHTTPMatch{PathPrefix: pathPrefix}},
		Destination: &consulk8s.ServiceRouteDestination{Service: service},
	}

	return result
}
