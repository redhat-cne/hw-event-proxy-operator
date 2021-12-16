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

package v1alpha1

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var hardwareeventlog = logf.Log.WithName("hardwareevent-resource")
var webhookClient client.Client

func (r *HardwareEvent) SetupWebhookWithManager(mgr ctrl.Manager) error {

	if webhookClient == nil {
		webhookClient = mgr.GetClient()
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-event-redhat-cne-org-v1alpha1-hardwareevent,mutating=false,failurePolicy=fail,sideEffects=None,groups=event.redhat-cne.org,resources=hardwareevents,verbs=create;update,versions=v1alpha1,name=vhardwareevent.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &HardwareEvent{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HardwareEvent) ValidateCreate() error {
	hardwareeventlog.Info("validate create", "name", r.Name)
	hwEventList := &HardwareEventList{}
	listOpts := []client.ListOption{
		client.InNamespace(r.Namespace),
	}
	err := webhookClient.List(context.TODO(), hwEventList, listOpts...)
	if err != nil {
		return err
	}
	if len(hwEventList.Items) >= 1 {
		return fmt.Errorf("only one Hardware Event instance is supported at this time")
	}

	if r.Spec.TransportHost == "" {
		return fmt.Errorf("transport URL is required field in Hardware instance spec")
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HardwareEvent) ValidateUpdate(old runtime.Object) error {
	hardwareeventlog.Info("validate update", "name", r.Name)

	if r.Spec.TransportHost == "" {
		return fmt.Errorf("transport URL is required field in Hardware instance spec")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HardwareEvent) ValidateDelete() error {
	hardwareeventlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
