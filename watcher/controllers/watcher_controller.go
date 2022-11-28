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
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/k1tests/basic-controller/api/v1beta1"
	k1v1beta1 "github.com/k1tests/basic-controller/api/v1beta1"
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
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
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
	eventType := EventUpdate
	// TODO(user): your logic here
	//log.Log.Info(fmt.Sprintf("Called: %#v", req))
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
			r.deleteJob(req.Name, Namespace)
			eventType = EventDelete
			log.Log.Info(fmt.Sprintf("Event: %s =>  %#v", eventType, req.NamespacedName))
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	// Check if Job was already created:
	if instance.Status.Instanced {
		// nothing to be done
		return reconcile.Result{}, nil
	}

	desiredJob, err := createWatcherJob(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	currentStateJob, err := r.getCurrentState(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	checkJob(currentStateJob)

	//missing job, creating one
	if currentStateJob == nil {
		eventType = EventCreate
		log.Log.Info(fmt.Sprintf("Event: %s =>  %#v", eventType, req.NamespacedName))
	}
	if currentStateJob == nil {
		err = r.Create(context.TODO(), desiredJob)
		if err != nil {
			return reconcile.Result{}, err
		}
		instance.Status.Status = "Started"
		instance.Status.Instanced = true
		err = r.Status().Update(context.Background(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if eventType == EventUpdate && currentStateJob != nil {
		if !reflect.DeepEqual(desiredJob.Spec, currentStateJob.Spec) {
			r.deleteJob(req.Name, Namespace)
			err = r.Create(context.TODO(), desiredJob)
			if err != nil {
				return reconcile.Result{}, err
			}
			instance.Status.Status = "Started"
			instance.Status.Instanced = true
			err = r.Status().Update(context.Background(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	} else {
		log.Log.Info(fmt.Sprintf("Unkown State: %#v", eventType))
	}
	return reconcile.Result{}, nil
}

func (r *WatcherReconciler) deleteJob(name string, namespace string) error {
	jobName, _ := generateNames(name, namespace)
	job := &v1batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: Namespace,
		},
		Spec: v1batch.JobSpec{},
	}
	err := r.Delete(context.TODO(), job)
	if err != nil {
		log.Log.Info(fmt.Sprintf("Error deleting Found job %s/%s\n", Namespace, job))
		return err

	}
	return nil
}

const (
	EventCreate string = "Create"
	EventUpdate string = "Update"
	EventDelete string = "Delete"
)

func (r *WatcherReconciler) getCurrentState(crd *k1v1beta1.Watcher) (*v1batch.Job, error) {
	jobName, _ := generateNames(crd.Name, Namespace)
	jobFound := &v1batch.Job{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: jobName, Namespace: Namespace}, jobFound)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info(fmt.Sprintf("Not Found Job %s/%s\n", Namespace, jobName))
		jobFound = nil
	}
	return jobFound, nil
}

func createWatcherJob(crd *k1v1beta1.Watcher) (*v1batch.Job, error) {
	jobName, _ := generateNames(crd.Name, Namespace)
	labels := map[string]string{"source": crd.GetObjectKind().GroupVersionKind().GroupKind().String(), "instance": crd.Name}
	//Adding the CRD labels, works, but argo-cd prune its as it is not defined on the git side.
	//maps.Copy(labels, crd.Labels)
	serviceAccount := ServiceAccount
	var one, five int32
	container := v1.Container{
		Name:            "main",
		Image:           "6zar/k1test:cf831de",
		ImagePullPolicy: v1.PullAlways,
		Command:         []string{"/usr/local/bin/k1-watcher"},
		Args:            []string{"watcher", "--crd-api-version", crd.APIVersion, "--crd-namespace", crd.Namespace, "--crd-instance", crd.Name},
	}
	one = int32(1)
	five = int32(5)
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
					Containers:         []v1.Container{container},
				},
			},
			BackoffLimit: &five,
			Completions:  &one,
		},
	}
	return job, nil
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
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldGeneration := e.ObjectOld.GetGeneration()
				newGeneration := e.ObjectNew.GetGeneration()
				// Generation is only updated on spec changes (also on deletion),
				// not metadata or status
				// Filter out events where the generation hasn't changed to
				// avoid being triggered by status updates
				return oldGeneration != newGeneration
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				// The reconciler adds a finalizer so we perform clean-up
				// when the delete timestamp is added
				// Suppress Delete events to avoid filtering them out in the Reconcile function
				return true
			},
		}).
		Complete(r)
}

func checkJob(job *v1batch.Job) {
	if job == nil {
		log.Log.Info("Job is nil")
	} else if reflect.DeepEqual(job, v1batch.Job{}) {
		log.Log.Info("Job is blank:" + job.Name)
	} else {
		log.Log.Info("Job exist:" + job.Name)
	}
}
