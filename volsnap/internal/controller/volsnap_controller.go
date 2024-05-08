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
	"net/http"

	snapclientv1 "github.com/kubernetes-csi/external-snapshotter/client/v7/clientset/versioned/typed/volumesnapshot/v1"
	"k8s.io/client-go/rest"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v7/apis/volumesnapshot/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	volv1 "neha-gupta1/volsnap/api/v1"
)

// VolsnapReconciler reconciles a Volsnap object
type VolsnapReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Config     *rest.Config
	HTTPClient *http.Client
}

//+kubebuilder:rbac:groups=vol.nehagupta1,resources=volsnaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vol.nehagupta1,resources=volsnaps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vol.nehagupta1,resources=volsnaps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Volsnap object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *VolsnapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var customVolsnapshot volv1.Volsnap

	if err := r.Get(ctx, req.NamespacedName, &customVolsnapshot); err != nil {
		log.Log.Error(err, "getting Volsnap")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	snapClient, err := snapclientv1.NewForConfigAndClient(r.Config, r.HTTPClient)
	if err != nil {
		log.Log.Error(err, "getting volumesnapshot client")
		return ctrl.Result{}, err
	}

	_, err = snapClient.VolumeSnapshots(req.NamespacedName.Namespace).Get(ctx, customVolsnapshot.Name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if err != nil {
		log.Log.Info("Snapshot not present we will create one!")

		snapshotClassName := "csi-hostpath-snapclass"
		newSnapshot := snapshotv1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Name: customVolsnapshot.Name,
			},
			Spec: snapshotv1.VolumeSnapshotSpec{
				VolumeSnapshotClassName: &snapshotClassName,
				Source: snapshotv1.VolumeSnapshotSource{
					PersistentVolumeClaimName: &customVolsnapshot.Spec.VolumeName,
				}}}

		_, err := snapClient.VolumeSnapshots(req.NamespacedName.Namespace).Create(ctx, &newSnapshot, metav1.CreateOptions{})
		if err != nil {
			log.Log.Error(err, "creating snapshot")
			return ctrl.Result{}, err
		}
	}

	log.Log.Info("Every thing ran fine!!!")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolsnapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&volv1.Volsnap{}).
		Complete(r)
}
