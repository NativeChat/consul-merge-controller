// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsulServiceIntentionsSource) DeepCopyInto(out *ConsulServiceIntentionsSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsulServiceIntentionsSource.
func (in *ConsulServiceIntentionsSource) DeepCopy() *ConsulServiceIntentionsSource {
	if in == nil {
		return nil
	}
	out := new(ConsulServiceIntentionsSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConsulServiceIntentionsSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsulServiceIntentionsSourceList) DeepCopyInto(out *ConsulServiceIntentionsSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ConsulServiceIntentionsSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsulServiceIntentionsSourceList.
func (in *ConsulServiceIntentionsSourceList) DeepCopy() *ConsulServiceIntentionsSourceList {
	if in == nil {
		return nil
	}
	out := new(ConsulServiceIntentionsSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConsulServiceIntentionsSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsulServiceIntentionsSourceSpec) DeepCopyInto(out *ConsulServiceIntentionsSourceSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsulServiceIntentionsSourceSpec.
func (in *ConsulServiceIntentionsSourceSpec) DeepCopy() *ConsulServiceIntentionsSourceSpec {
	if in == nil {
		return nil
	}
	out := new(ConsulServiceIntentionsSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsulServiceIntentionsSourceStatus) DeepCopyInto(out *ConsulServiceIntentionsSourceStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsulServiceIntentionsSourceStatus.
func (in *ConsulServiceIntentionsSourceStatus) DeepCopy() *ConsulServiceIntentionsSourceStatus {
	if in == nil {
		return nil
	}
	out := new(ConsulServiceIntentionsSourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsulServiceRoute) DeepCopyInto(out *ConsulServiceRoute) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsulServiceRoute.
func (in *ConsulServiceRoute) DeepCopy() *ConsulServiceRoute {
	if in == nil {
		return nil
	}
	out := new(ConsulServiceRoute)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConsulServiceRoute) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsulServiceRouteList) DeepCopyInto(out *ConsulServiceRouteList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ConsulServiceRoute, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsulServiceRouteList.
func (in *ConsulServiceRouteList) DeepCopy() *ConsulServiceRouteList {
	if in == nil {
		return nil
	}
	out := new(ConsulServiceRouteList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConsulServiceRouteList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsulServiceRouteSpec) DeepCopyInto(out *ConsulServiceRouteSpec) {
	*out = *in
	in.Route.DeepCopyInto(&out.Route)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsulServiceRouteSpec.
func (in *ConsulServiceRouteSpec) DeepCopy() *ConsulServiceRouteSpec {
	if in == nil {
		return nil
	}
	out := new(ConsulServiceRouteSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsulServiceRouteStatus) DeepCopyInto(out *ConsulServiceRouteStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsulServiceRouteStatus.
func (in *ConsulServiceRouteStatus) DeepCopy() *ConsulServiceRouteStatus {
	if in == nil {
		return nil
	}
	out := new(ConsulServiceRouteStatus)
	in.DeepCopyInto(out)
	return out
}
