// +build !ignore_autogenerated

/*
Copyright 2021.

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

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KunInstallation) DeepCopyInto(out *KunInstallation) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KunInstallation.
func (in *KunInstallation) DeepCopy() *KunInstallation {
	if in == nil {
		return nil
	}
	out := new(KunInstallation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KunInstallation) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KunInstallationIngress) DeepCopyInto(out *KunInstallationIngress) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KunInstallationIngress.
func (in *KunInstallationIngress) DeepCopy() *KunInstallationIngress {
	if in == nil {
		return nil
	}
	out := new(KunInstallationIngress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KunInstallationList) DeepCopyInto(out *KunInstallationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KunInstallation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KunInstallationList.
func (in *KunInstallationList) DeepCopy() *KunInstallationList {
	if in == nil {
		return nil
	}
	out := new(KunInstallationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KunInstallationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KunInstallationMongodb) DeepCopyInto(out *KunInstallationMongodb) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KunInstallationMongodb.
func (in *KunInstallationMongodb) DeepCopy() *KunInstallationMongodb {
	if in == nil {
		return nil
	}
	out := new(KunInstallationMongodb)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KunInstallationRegistry) DeepCopyInto(out *KunInstallationRegistry) {
	*out = *in
	if in.Auth != nil {
		in, out := &in.Auth, &out.Auth
		*out = make([]KunInstallationRegistryAuth, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KunInstallationRegistry.
func (in *KunInstallationRegistry) DeepCopy() *KunInstallationRegistry {
	if in == nil {
		return nil
	}
	out := new(KunInstallationRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KunInstallationRegistryAuth) DeepCopyInto(out *KunInstallationRegistryAuth) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KunInstallationRegistryAuth.
func (in *KunInstallationRegistryAuth) DeepCopy() *KunInstallationRegistryAuth {
	if in == nil {
		return nil
	}
	out := new(KunInstallationRegistryAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KunInstallationServer) DeepCopyInto(out *KunInstallationServer) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
	in.Resources.DeepCopyInto(&out.Resources)
	out.Ingress = in.Ingress
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KunInstallationServer.
func (in *KunInstallationServer) DeepCopy() *KunInstallationServer {
	if in == nil {
		return nil
	}
	out := new(KunInstallationServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KunInstallationSpec) DeepCopyInto(out *KunInstallationSpec) {
	*out = *in
	in.Registry.DeepCopyInto(&out.Registry)
	out.Mongodb = in.Mongodb
	in.Server.DeepCopyInto(&out.Server)
	in.UI.DeepCopyInto(&out.UI)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KunInstallationSpec.
func (in *KunInstallationSpec) DeepCopy() *KunInstallationSpec {
	if in == nil {
		return nil
	}
	out := new(KunInstallationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KunInstallationStatus) DeepCopyInto(out *KunInstallationStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KunInstallationStatus.
func (in *KunInstallationStatus) DeepCopy() *KunInstallationStatus {
	if in == nil {
		return nil
	}
	out := new(KunInstallationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KunInstallationUI) DeepCopyInto(out *KunInstallationUI) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
	in.Resources.DeepCopyInto(&out.Resources)
	out.Ingress = in.Ingress
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KunInstallationUI.
func (in *KunInstallationUI) DeepCopy() *KunInstallationUI {
	if in == nil {
		return nil
	}
	out := new(KunInstallationUI)
	in.DeepCopyInto(out)
	return out
}
