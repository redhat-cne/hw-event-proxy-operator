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
	"net/url"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
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

// +kubebuilder:webhook:path=/validate-event-redhat-cne-org-v1alpha1-hardwareevent,mutating=false,failurePolicy=fail,sideEffects=None,groups=event.redhat-cne.org,resources=hardwareevents,verbs=create;update,versions=v1alpha1,name=vhardwareevent.kb.io,admissionReviewVersions=v1

const (
	AmqScheme            = "amqp"
	DefaultTransportHost = "http://hw-event-publisher-service.openshift-bare-metal-events.svc.cluster.local:9043"
	// storageTypeEmptyDir is used for developer tests to map pubsubstore volume to emptyDir
	storageTypeEmptyDir = "emptyDir"
)

var _ webhook.Validator = &HardwareEvent{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HardwareEvent) ValidateCreate() (admission.Warnings, error) {
	hardwareeventlog.Info("validate create", "name", r.Name)
	hwEventList := &HardwareEventList{}
	listOpts := []client.ListOption{
		client.InNamespace(r.Namespace),
	}
	err := webhookClient.List(context.TODO(), hwEventList, listOpts...)
	if err != nil {
		return admission.Warnings{}, err
	}
	if len(hwEventList.Items) >= 1 {
		return admission.Warnings{}, fmt.Errorf("only one Hardware Event instance is supported at this time")
	}

	return admission.Warnings{}, r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HardwareEvent) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	hardwareeventlog.Info("validate update", "name", r.Name)
	return admission.Warnings{}, r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HardwareEvent) ValidateDelete() (admission.Warnings, error) {
	hardwareeventlog.Info("validate delete", "name", r.Name)
	return admission.Warnings{}, nil
}

func (r *HardwareEvent) validate() error {

	eventConfig := r.Spec
	_, err := url.Parse(eventConfig.TransportHost)
	if eventConfig.TransportHost == "" || err != nil {
		hardwareeventlog.Info("transportHost is not valid, proceed as", "transportHost", DefaultTransportHost)
	}

	return nil
}
