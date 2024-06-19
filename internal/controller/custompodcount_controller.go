/*
Copyright 2024.


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

package controller

import (
	"context"
	"fmt"
	"time"

	apiv1alpha1 "github.com/ShravaniVangur/custompodcount-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CustompodcountReconciler reconciles a Custompodcount object
type CustompodcountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.example.com,resources=custompodcounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.example.com,resources=custompodcounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.example.com,resources=custompodcounts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Custompodcount object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *CustompodcountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	//Step 1: Retrieve the defined details
	custompodcount := &apiv1alpha1.Custompodcount{}
	err := r.Get(ctx, req.NamespacedName, custompodcount)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "Failed to get Custompodcount ")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	podNamespace := req.Namespace
	podName := "custompod-count"

	//Step 2: retrieve the total number of pods running
	currentPodCount, err := r.countPods(ctx, podNamespace, podName)
	if err != nil {
		log.Error(err, "Failed to count pods")
		return ctrl.Result{}, err
	}

	expectedPodCount := custompodcount.Spec.SizePod

	//Step 3: check if the number differs
	if int32(currentPodCount) < expectedPodCount {
		for i := int32(currentPodCount); i < expectedPodCount; i++ {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", podName, time.Now().UnixNano()),
					Namespace: podNamespace,
					Labels: map[string]string{
						"app": podName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx-container",
							Image: "docker.io/library/nginx:latest",
						},
					},
				},
			}
			if err := r.Create(ctx, pod); err != nil {
				log.Error(err, "Failed to create pod", "pod", pod.Name)
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *CustompodcountReconciler) countPods(ctx context.Context, namespace, prefix string) (int, error) {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{"app": prefix},
	}

	if err := r.List(ctx, podList, listOpts...); err != nil {
		return 0, fmt.Errorf("failed to list pods: %v", err)
	}

	return len(podList.Items), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustompodcountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Custompodcount{}).
		Complete(r)
}
