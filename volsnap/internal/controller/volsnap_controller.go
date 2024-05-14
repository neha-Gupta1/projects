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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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

	var (
		customVolsnapshot       volv1.Volsnap
		customSnapFinalizerName = "nehagupta1/finalizer"
	)

	if err := r.Get(ctx, req.NamespacedName, &customVolsnapshot); err != nil {
		log.Log.Error(err, "getting Volsnap")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !customVolsnapshot.ObjectMeta.DeletionTimestamp.IsZero() &&
		controllerutil.ContainsFinalizer(&customVolsnapshot, customSnapFinalizerName) {
		err := r.handleFinalizer(ctx, customVolsnapshot, customSnapFinalizerName)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// registering finalizer
	if !controllerutil.ContainsFinalizer(&customVolsnapshot, customSnapFinalizerName) {
		controllerutil.AddFinalizer(&customVolsnapshot, customSnapFinalizerName)
		if err := r.Update(ctx, &customVolsnapshot); err != nil {
			return ctrl.Result{}, err
		}
	}

	snapClient, err := r.createSnapshotClient()
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = snapClient.VolumeSnapshots(req.NamespacedName.Namespace).Get(
		ctx,
		customVolsnapshot.Name,
		metav1.GetOptions{})
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

		volumeSnapshot, err := snapClient.VolumeSnapshots(req.NamespacedName.Namespace).Create(ctx, &newSnapshot, metav1.CreateOptions{})
		if err != nil && !errors.IsNotFound(err) {
			log.Log.Error(err, "creating snapshot")
			customVolsnapshot.Status = volv1.VolsnapStatus{
				RunningStatus: "Failed",
			}

			err = r.Status().Update(context.Background(), &customVolsnapshot)
			if err != nil {
				return reconcile.Result{}, err
			}

			return ctrl.Result{}, err
		}

		// snapshot not yet created. Lets wait for it to be created
		if err != nil {
			log.Log.Info("Snapshot not yet created. Requeuing")
			customVolsnapshot.Status = volv1.VolsnapStatus{
				RunningStatus: "Pending",
			}

			err = r.Status().Update(context.Background(), &customVolsnapshot)
			if err != nil {
				return reconcile.Result{}, err
			}

			return ctrl.Result{Requeue: true}, nil
		}

		// update snapshot info in the volsnap
		customVolsnapshot.Spec.VolumeName = *volumeSnapshot.Spec.Source.PersistentVolumeClaimName
		customVolsnapshot.Spec.SnapshotName = volumeSnapshot.Name

		if err := r.Update(ctx, &customVolsnapshot); err != nil {
			return ctrl.Result{}, err
		}
	}

	//status start
	customVolsnapshot.Status = volv1.VolsnapStatus{
		RunningStatus: "Created",
	}

	err = r.Status().Update(context.Background(), &customVolsnapshot)
	if err != nil {
		return reconcile.Result{}, err
	}

	//status end

	log.Log.Info("Every thing ran fine!!!")

	return ctrl.Result{}, nil
}

func (r *VolsnapReconciler) createSnapshotClient() (client *snapclientv1.SnapshotV1Client, err error) {
	client, err = snapclientv1.NewForConfigAndClient(r.Config, r.HTTPClient)
	if err != nil {
		log.Log.Error(err, "getting volumesnapshot client")
		return nil, err
	}

	return client, err
}

func (r *VolsnapReconciler) handleFinalizer(ctx context.Context, customVolsnapshot volv1.Volsnap, customSnapFinalizerName string) (err error) {
	if err = r.deleteVolumeSnapshot(ctx, customVolsnapshot); err != nil {
		// if fail to delete the external dependency here, return with error
		// so that it can be retried.
		return err
	}

	controllerutil.RemoveFinalizer(&customVolsnapshot, customSnapFinalizerName)
	if err := r.Update(ctx, &customVolsnapshot); err != nil {
		log.Log.Error(err, "Error removing finalizer")
		return err
	}

	return nil
}

func (r *VolsnapReconciler) deleteVolumeSnapshot(ctx context.Context, customVolsnapshot volv1.Volsnap) error {
	snapClient, err := r.createSnapshotClient()
	if err != nil {
		return err
	}

	volumesnapshot, err := snapClient.VolumeSnapshots(customVolsnapshot.Namespace).Get(ctx, customVolsnapshot.Name, metav1.GetOptions{})
	if err != nil {
		log.Log.Error(err, "getting snapshot")
		return err
	}

	err = snapClient.VolumeSnapshots(customVolsnapshot.Namespace).Delete(ctx, volumesnapshot.Name, metav1.DeleteOptions{})
	if err != nil {
		log.Log.Error(err, "deleting snapshot")
		return err
	}

	return nil
}

func ignoreUpdationPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolsnapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&volv1.Volsnap{}).
		WithEventFilter(ignoreUpdationPredicate()).
		Complete(r)
}
