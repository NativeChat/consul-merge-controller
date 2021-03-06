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

package finalizers

import (
	"fmt"

	servicev1alpha1 "github.com/NativeChat/consul-merge-controller/apis/service/v1alpha1"
)

var (
	// ConsulServiceRouteFinalizerName is the name of the finalizer used for consul service routes.
	ConsulServiceRouteFinalizerName = fmt.Sprintf("finalizer.%s", servicev1alpha1.GroupVersion.Group)
)
