/*
Copyright 2022 K-rays.

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
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/k1tests/basic-controller/api/v1beta1"
	k1v1beta1 "github.com/k1tests/basic-controller/api/v1beta1"
	"gopkg.in/yaml.v2"
	v1batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WatcherReconciler reconciles a Watcher object
type WatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// ServiceAccount needed to run informer job
const ServiceAccount = "k1-ready"

// Namespace where the jobs will be excuted
const Namespace = "default"

//const Namespace = "k1-watcher"

//+kubebuilder:rbac:groups=k1.kubefirst.io,resources=watchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k1.kubefirst.io,resources=watchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k1.kubefirst.io,resources=watchers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Watcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *WatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	// TODO(user): your logic here
	log.Log.Info(fmt.Sprintf("Called: %#v", req))
	log.Log.Info(fmt.Sprintf("Called: %#v", req.NamespacedName))

	//Desired State
	instance := &v1beta1.Watcher{}

	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// CRD was removed (DELETE EVENT)
			// How to remove the objects?
			r.deleteWatcher(req.Name, Namespace)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	//Other Events:
	//Created and Updated
	//How to check if it is is an update?
	// "Create Again" the Object and Compare with existing one
	// If not matchs it is a important update.
	desiredJob, desiredConfigMap, err := createWatcherJob(instance)
	log.Log.Info(fmt.Sprintf("Called: %#v", desiredJob))
	log.Log.Info(fmt.Sprintf("Called: %#v", desiredConfigMap))
	if err != nil {
		return reconcile.Result{}, err
	}

	currentStateJob, currentStateConfigMap, err := r.getCurrentState(instance)
	log.Log.Info(fmt.Sprintf("Called: %#v", currentStateJob))
	log.Log.Info(fmt.Sprintf("Called: %#v", currentStateConfigMap))
	if err != nil {
		return reconcile.Result{}, err
	}
	//Get Live Resources:
	// ConfigMap
	// Job
	// if any is missing, destroy and create again.
	// if both are missing, just create.
	// if both match, do nothing.

	//missing job, creating one
	if currentStateConfigMap == nil {
		err = r.Create(context.TODO(), desiredConfigMap)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	if currentStateJob == nil {
		err = r.Create(context.TODO(), desiredJob)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	currentStateJob, currentStateConfigMap, err = r.getCurrentState(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	//Handle Updates:
	if currentStateJob == nil || currentStateConfigMap == nil {
		//nothing to be updates
		return ctrl.Result{}, nil
	}

	if !reflect.DeepEqual(desiredConfigMap.Data, currentStateConfigMap.Data) || !reflect.DeepEqual(desiredJob.Spec, currentStateJob.Spec) {
		r.deleteWatcher(req.Name, Namespace)
		err = r.Create(context.TODO(), desiredConfigMap)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.Create(context.TODO(), desiredJob)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *WatcherReconciler) deleteWatcher(name string, namespace string) error {
	jobName, configMapName := generateNames(name, namespace)
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: Namespace,
		},
	}
	err := r.Delete(context.TODO(), configMap)
	if err != nil {
		log.Log.Info(fmt.Sprintf("Error deleting Found Configmap %s/%s\n", Namespace, configMapName))
		return err

	}
	job := &v1batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: Namespace,
		},
		Spec: v1batch.JobSpec{},
	}
	err = r.Delete(context.TODO(), job)
	if err != nil {
		log.Log.Info(fmt.Sprintf("Error deleting Found job %s/%s\n", Namespace, job))
		return err

	}

	return nil
}
func (r *WatcherReconciler) createWatcher(job *v1batch.Job, configMap *v1.ConfigMap) error {
	return nil
}

func (r *WatcherReconciler) updateWatcher(job *v1batch.Job, configMap *v1.ConfigMap) error {
	return nil
}
func (r *WatcherReconciler) getCurrentState(crd *k1v1beta1.Watcher) (*v1batch.Job, *v1.ConfigMap, error) {
	jobName, configMapName := generateNames(crd.Name, Namespace)
	configMapFound := &v1.ConfigMap{}
	jobFound := &v1batch.Job{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: configMapName, Namespace: Namespace}, configMapFound)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info(fmt.Sprintf("Not Found Configmap %s/%s\n", Namespace, configMapName))
		configMapFound = nil
	}
	err = r.Get(context.TODO(), types.NamespacedName{Name: jobName, Namespace: Namespace}, jobFound)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info(fmt.Sprintf("Not Found Job %s/%s\n", Namespace, jobName))
		jobFound = nil
	}
	return jobFound, configMapFound, nil
}

func createConfigWatcher(crd *k1v1beta1.Watcher) *k1v1beta1.WatcherConfig {
	configWatcher := &k1v1beta1.WatcherConfig{
		CrdName:      crd.Name,
		CrdNamespace: crd.Namespace,
		Kind:         crd.Kind,
		APIVersion:   crd.APIVersion,
		Group:        crd.GroupVersionKind().Group,
	}
	return configWatcher
}

func createWatcherJob(crd *k1v1beta1.Watcher) (*v1batch.Job, *v1.ConfigMap, error) {
	jobName, configMapName := generateNames(crd.Name, Namespace)
	watcherRules, _ := yaml.Marshal(crd.Spec)
	watcherConfig, _ := yaml.Marshal(createConfigWatcher(crd))
	log.Log.Info(fmt.Sprintf("Called: %s", watcherRules))
	dataSample := map[string]string{"check.yaml": string(watcherRules), "owner.yaml": string(watcherConfig)}
	//TODO: Improve logic to create ownership matching
	labels := map[string]string{"source": crd.GetObjectKind().GroupVersionKind().GroupKind().String(), "instance": crd.Name}

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: Namespace,
			Labels:    labels,
		},
		Data: dataSample,
	}
	serviceAccount := ServiceAccount
	var one int32
	volume := v1.Volume{
		Name: "k1-ready-config",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configMap.ObjectMeta.Name,
				},
			},
		},
	}
	container := v1.Container{
		Name:            "main",
		Image:           "6zar/k1test:a8a2386",
		ImagePullPolicy: v1.PullAlways,
		Command:         []string{"/usr/local/bin/k1-watcher"},
		Args:            []string{"watcher", "-c", "/k1-config/check.yaml", "-o", "/k1-config/owner.yaml"},
		VolumeMounts:    []v1.VolumeMount{{Name: volume.Name, MountPath: "/k1-config"}},
	}
	one = int32(1)
	job := &v1batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: Namespace,
			Labels:    labels,
		},
		Spec: v1batch.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					ServiceAccountName: serviceAccount,
					RestartPolicy:      v1.RestartPolicyNever,
					Volumes:            []v1.Volume{volume},
					Containers:         []v1.Container{container},
				},
			},
			BackoffLimit: &one,
			Completions:  &one,
		},
	}
	return job, configMap, nil
}

func generateNames(name string, namespace string) (string, string) {
	configMapName := name + "-cm"
	jobName := name + "-job"
	return jobName, configMapName
}

// SetupWithManager sets up the controller with the Manager.
func (r *WatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k1v1beta1.Watcher{}).
		Complete(r)
}
