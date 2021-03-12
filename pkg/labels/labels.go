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

package labels

import (
	"fmt"

	servicev1alpha1 "github.com/NativeChat/consul-merge-controller/apis/service/v1alpha1"
)

var (
	// ServiceRouter is the name of the label which stores the service router name.
	ServiceRouter = fmt.Sprintf("%s/service-router", servicev1alpha1.GroupVersion.Group)

	// ServiceIntentions is the name of the label which stores the service intentions name.
	ServiceIntentions = fmt.Sprintf("%s/service-intentions", servicev1alpha1.GroupVersion.Group)
)
