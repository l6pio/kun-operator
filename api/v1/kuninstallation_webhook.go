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

package v1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var kunlog = logf.Log.WithName("kuninstallation-resource")

func (r *KunInstallation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-l6p-io-v1-kuninstallation,mutating=true,failurePolicy=fail,sideEffects=None,groups=l6p.io,resources=kuninstallations,verbs=create;update,versions=v1,name=mkuninstallation.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &KunInstallation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KunInstallation) Default() {
	kunlog.Info("default", "name", r.Name)

	if r.Spec.Hub == "" {
		r.Spec.Hub = "docker.io/l6pio"
	}

	if r.Spec.Server.Replicas == nil {
		r.Spec.Server.Replicas = new(int32)
		*r.Spec.Server.Replicas = 1
	}

	if r.Spec.Server.Port == nil {
		r.Spec.Server.Port = new(int32)
		*r.Spec.Server.Port = 80
	}

	if r.Spec.UI.Replicas == nil {
		r.Spec.UI.Replicas = new(int32)
		*r.Spec.UI.Replicas = 1
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-l6p-io-v1-kuninstallation,mutating=false,failurePolicy=fail,sideEffects=None,groups=l6p.io,resources=kuninstallations,verbs=create;update,versions=v1,name=vkuninstallation.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &KunInstallation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KunInstallation) ValidateCreate() error {
	kunlog.Info("validate create", "name", r.Name)
	return r.ValidateInstallation()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KunInstallation) ValidateUpdate(old runtime.Object) error {
	kunlog.Info("validate update", "name", r.Name)
	return r.ValidateInstallation()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KunInstallation) ValidateDelete() error {
	kunlog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *KunInstallation) ValidateInstallation() error {
	var allErrs field.ErrorList
	if r.Spec.Mongodb.Addr == "" {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec").Child("mongodb").Child("addr"),
			r.Spec.Mongodb.Addr,
			"cannot be empty",
		))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "kun.l6p.io", Kind: "KunInstallation"},
		r.Name, allErrs)
}
