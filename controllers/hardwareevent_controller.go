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
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	kscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// HardwareEventReconciler reconciles a HardwareEvent object
type HardwareEventReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

const (
	AmqScheme            = "amqp"
	DefaultTransportHost = "http://hw-event-publisher-service.openshift-bare-metal-events.svc.cluster.local:9043"
	DefaultStorageType   = "emptyDir"
	PVCNamePrefix        = "cloud-event-proxy-store"
)

//+kubebuilder:rbac:groups=event.redhat-cne.org,resources=hardwareevents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets;endpoints;pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=event.redhat-cne.org,resources=hardwareevents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=event.redhat-cne.org,resources=hardwareevents/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;update;patch;create
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=services,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create;delete;get;watch;list;update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings;roles;rolebindings,verbs=create;delete;get;watch;list;update
//+kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create;delete;get;watch;list;update
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list;watch

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

	if err = r.syncHwEventProxy(ctx, req.Namespace, instance); err != nil {
		reqLogger.Error(err, "failed to sync hardware event proxy deployment ")
		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	instance.Status.LastSynced = &v1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, instance); err != nil {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// syncHwEventProxy synchronizes HardwareEvent CR
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
	if instance.Spec.MsgParserTimeout <= 0 {
		data.Data["MsgParserTimeOut"] = 10
	} else {
		data.Data["MsgParserTimeOut"] = instance.Spec.MsgParserTimeout
	}
	data.Data["ReleaseVersion"] = os.Getenv("RELEASE_VERSION")
	data.Data["KubeRbacProxy"] = os.Getenv("KUBE_RBAC_PROXY_IMAGE")
	data.Data["SideCar"] = os.Getenv("CLOUD_EVENT_PROXY_IMAGE")

	if instance.Spec.TransportHost == "" {
		data.Data["EventTransportHost"] = DefaultTransportHost
	} else {
		data.Data["EventTransportHost"] = instance.Spec.TransportHost
	}

	// storageType is already checked in webhook
	data.Data["StorageType"] = instance.Spec.StorageType

	objs, err = render.RenderDir(filepath.Join(names.ManifestDir, "hw-event-proxy"), &data)
	if err != nil {
		return fmt.Errorf("failed to render hardware event proxy deployment manifest: %v", err)
	}

	err = r.cleanupPvc(ctx, namespace, fmt.Sprintf("%s", data.Data["StorageType"]))
	if err != nil {
		return err
	}

	for _, obj := range objs {
		if obj, err = r.setNodeSelector(instance, obj); err != nil {
			return fmt.Errorf("failed to apply %s object %v with node selector err: %v", obj.GetName(), obj, err)
		}
		if obj.GetKind() == "PersistentVolumeClaim" {
			obj, err = r.syncPvc(ctx, namespace, obj)
			if err != nil {
				return err
			}
		}
		if err = controllerutil.SetControllerReference(instance, obj, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %v", err)
		}
		if err = apply.ApplyObject(ctx, r.Client, obj); err != nil {
			return fmt.Errorf("failed to apply %s object %v with err: %v", obj.GetName(), obj, err)
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HardwareEventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hwEventV1alpha1.HardwareEvent{}).
		Owns(&appsv1.Deployment{}).
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

// setNodeSelector synchronizes deployment
func (r *HardwareEventReconciler) setNodeSelector(
	defaultCfg *hwEventV1alpha1.HardwareEvent,
	obj *uns.Unstructured,
) (*uns.Unstructured, error) {
	var err error
	if obj.GetKind() == "Deployment" && defaultCfg.Spec.NodeSelector != nil && len(defaultCfg.Spec.NodeSelector) > 0 {
		scheme := kscheme.Scheme
		deps := &appsv1.Deployment{}
		err = scheme.Convert(obj, deps, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to convert obj hw-event-proxy to appsv1.Deployment: %v", err)
		}
		deps.Spec.Template.Spec.NodeSelector = defaultCfg.Spec.NodeSelector
		err = scheme.Convert(deps, obj, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to convert appsv1.Deployment to hw-event-proxy obj: %v", err)
		}
	}
	return obj, nil
}

// syncPvc update PersistentVolumeClaim
func (r *HardwareEventReconciler) syncPvc(ctx context.Context, namespace string, obj *uns.Unstructured) (*uns.Unstructured, error) {
	var err error
	scheme := kscheme.Scheme
	pvc := &corev1.PersistentVolumeClaim{}
	err = scheme.Convert(obj, pvc, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to convert obj to corev1.PersistentVolumeClaim: %v", err)
	}

	// update the VolumeName of PVC when the PVC is bound to a PV
	if pvcDeployed := r.getPvc(obj.GetName(), namespace); pvcDeployed != nil {
		if pvcDeployed.Spec.VolumeName != pvc.Spec.VolumeName && pvc.Spec.VolumeName == "" {
			log.Printf("pvc %s is Bound, updating VolumeName to %s", obj.GetName(), pvcDeployed.Spec.VolumeName)
			pvc.Spec.VolumeName = pvcDeployed.Spec.VolumeName
		}
	}

	err = scheme.Convert(pvc, obj, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to convert corev1.PersistentVolumeClaim to obj: %v", err)
	}

	return obj, nil
}

// cleanupPvc clean up obsolete PVCs not mounted to current hw-event-proxy pod
func (r *HardwareEventReconciler) cleanupPvc(ctx context.Context, namespace string, storageType string) error {
	var err error
	pvcList := &corev1.PersistentVolumeClaimList{}
	opts := []client.ListOption{
		client.InNamespace(namespace),
	}
	err = r.List(context.TODO(), pvcList, opts...)
	if err != nil {
		return err
	}

	pvcName := fmt.Sprintf("%s-%s", PVCNamePrefix, storageType)

	for _, p := range pvcList.Items {
		if strings.HasPrefix(p.ObjectMeta.Name, PVCNamePrefix) {
			if p.ObjectMeta.Name != pvcName || storageType == DefaultStorageType {
				if err := r.Client.Delete(ctx, &p); err != nil {
					log.Printf("fail to delete obsolete pvc %s err: %v", p.ObjectMeta.Name, err)
				} else {
					log.Printf("garbage collection: successfully deleted obsolete pvc %s", p.ObjectMeta.Name)
				}
			}
		}

	}
	return nil
}

func (r *HardwareEventReconciler) getPvc(pvcName string, namespace string) *corev1.PersistentVolumeClaim {

	pvcList := &corev1.PersistentVolumeClaimList{}
	opts := []client.ListOption{
		client.InNamespace(namespace),
	}
	err := r.List(context.TODO(), pvcList, opts...)
	if err != nil {
		return nil
	}
	for _, p := range pvcList.Items {
		if p.ObjectMeta.Name == pvcName {
			return &p
		}

	}
	return nil
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
