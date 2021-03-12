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

import "fmt"

const (
	// ServiceGroup is the group name for the service api.
	ServiceGroup = "service.consul.k8s.nativechat.com"

	// ServiceGroupVersion is the group version for the service api.
	ServiceGroupVersion = "v1alpha1"

	// DefaultK8sNamespace is the name of the default k8s namespace which will be used in the tests.
	DefaultK8sNamespace = "default"
)

var (
	// ServiceRouterLabel is the name of the label which stores the service router name.
	ServiceRouterLabel = fmt.Sprintf("%s/service-router", ServiceGroup)

	// ServiceIntentions is the name of the label which stores the service intentions name.
	ServiceIntentions = fmt.Sprintf("%s/service-intentions", ServiceGroup)

	// ServiceFinalizer is the name of the service finalizer.
	ServiceFinalizer = fmt.Sprintf("finalizer.%s", ServiceGroup)
)
