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

package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	hwEventV1alpha1 "github.com/redhat-cne/hw-event-proxy-operator/api/v1alpha1"
	"github.com/redhat-cne/hw-event-proxy-operator/pkg/apply"
	"github.com/redhat-cne/hw-event-proxy-operator/pkg/names"
	"github.com/redhat-cne/hw-event-proxy-operator/pkg/render"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// HardwareEventReconciler reconciles a HardwareEvent object
type HardwareEventReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=event.redhat-cne.org,resources=hardwareevents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets;endpoints;pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=event.redhat-cne.org,resources=hardwareevents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=event.redhat-cne.org,resources=hardwareevents/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;update;patch;create
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=services,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create;delete;get;watch;list;update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings;roles;rolebindings,verbs=create;delete;get;watch;list;update
//+kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create;delete;get;watch;list;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HardwareEvent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *HardwareEventReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = log.FromContext(ctx)
	reqLogger := r.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling Hardware event proxy")
	instance := &hwEventV1alpha1.HardwareEvent{}
	err := r.Get(ctx, req.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			// For additional cleanup logic use finalizers. Return and don't requeue.
			reqLogger.Info("Instance of hardwareevents.event.redhat-cne.org not found")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "error listing instance of hardwareevents.event.redhat-cne.org ")
		return ctrl.Result{}, err
	}
	//check if secret is found
	_, _, err = r.getSecret(names.RedfishSecretName, req.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("redfish secret not found, please create a secret to access hardware events ")
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	// check if deployment is found
	deps := appsv1.Deployment{}
	err = r.Get(ctx, req.NamespacedName, &deps)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			// For additional cleanup logic use finalizers. Return and don't requeue.
			if err = r.syncHwEventProxy(ctx, req.Namespace, instance); err != nil {
				reqLogger.Error(err, "failed to sync hardware event proxy deployment ")
				return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
			}

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "error reading the deployment")
		return ctrl.Result{}, err
	}
	instance.Status.LastSynced = &v1.Time{Time: time.Now()}
	return ctrl.Result{}, nil
}

// syncPtpConfig synchronizes PtpConfig CR
func (r *HardwareEventReconciler) syncHwEventProxy(ctx context.Context, namespace string, instance *hwEventV1alpha1.HardwareEvent) error {
	var err error

	objs := []*uns.Unstructured{}

	data := render.MakeRenderData()
	data.Data["Image"] = os.Getenv("HW_EVENT_PROXY_IMAGE")
	data.Data["Namespace"] = namespace
	data.Data["LogLevel"] = instance.Spec.LogLevel
	if instance.Spec.LogLevel == "" {
		data.Data["LogLevel"] = "debug"
	}
	data.Data["MsgParserTimeOut"] = "10"
	data.Data["ReleaseVersion"] = os.Getenv("RELEASE_VERSION")
	data.Data["KubeRbacProxy"] = os.Getenv("KUBE_RBAC_PROXY_IMAGE")
	data.Data["SideCar"] = os.Getenv("CLOUD_EVENT_PROXY_IMAGE")
	data.Data["EventTransportHost"] = instance.Spec.TransportHost

	objs, err = render.RenderDir(filepath.Join(names.ManifestDir, "hw-event-proxy"), &data)
	if err != nil {
		return fmt.Errorf("failed to render hardware event proxy deployment manifest: %v", err)
	}

	for _, obj := range objs {
		if err = apply.ApplyObject(ctx, r.Client, obj); err != nil {
			return fmt.Errorf("failed to apply %s object %v with err: %v", obj.GetName(), obj, err)
		}
		if err = controllerutil.SetControllerReference(instance, obj, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %v", err)
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HardwareEventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hwEventV1alpha1.HardwareEvent{}).
		Complete(r)
}

func (r *HardwareEventReconciler) getSecret(secretName string, secretNamespace string) (*corev1.Secret, string, error) {
	secret := &corev1.Secret{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret)
	if err != nil {
		return nil, "", err
	}

	secretHash, err := ObjectHash(secret)
	if err != nil {
		return nil, "", fmt.Errorf("error calculating configuration hash: %v", err)
	}
	return secret, secretHash, nil
}

// ObjectHash creates a deep object hash and return it as a safe encoded string
func ObjectHash(i interface{}) (string, error) {
	// Convert the hashSource to a byte slice so that it can be hashed
	hashBytes, err := json.Marshal(i)
	if err != nil {
		return "", fmt.Errorf("unable to convert to JSON: %v", err)
	}
	hash := sha256.Sum256(hashBytes)
	return rand.SafeEncodeString(fmt.Sprint(hash)), nil
}
