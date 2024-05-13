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
	"net/http"

	snapclientv1 "github.com/kubernetes-csi/external-snapshotter/client/v7/clientset/versioned/typed/volumesnapshot/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pvcClient "k8s.io/api/core/v1"

	volv1 "neha-gupta1/volsnap/api/v1"
)

// VolrestoreReconciler reconciles a Volrestore object
type VolrestoreReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Config     *rest.Config
	HTTPClient *http.Client
}

//+kubebuilder:rbac:groups=vol.nehagupta1,resources=volrestores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vol.nehagupta1,resources=volrestores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vol.nehagupta1,resources=volrestores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Volrestore object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *VolrestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var (
		volRestore volv1.Volrestore
	)

	if err := r.Get(ctx, req.NamespacedName, &volRestore); err != nil {
		fmt.Println("getting Volrestore", err)
		log.Log.Info("getting Volrestore not found	")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	snapClient, err := r.createSnapshotClient()
	if err != nil {
		return ctrl.Result{}, err
	}

	existingVolsnapshot, err := snapClient.VolumeSnapshots(req.NamespacedName.Namespace).Get(ctx, volRestore.Spec.VolumeSnapName, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		log.Log.Error(err, "no volume snapshot present")
		return ctrl.Result{}, err
	}

	snapshotClassName := "csi-hostpath-sc"
	snapshotGroup := existingVolsnapshot.GroupVersionKind().Group

	vol := &pvcClient.PersistentVolumeClaim{ // Instantiate a new PersistentVolumeClaim object
		TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s%s", "restore", volRestore.Spec.VolumeSnapName),
			Namespace: req.NamespacedName.Namespace,
		},
		Spec: pvcClient.PersistentVolumeClaimSpec{
			StorageClassName: &snapshotClassName,
			// DataSource:
			AccessModes: []pvcClient.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			DataSource: &pvcClient.TypedLocalObjectReference{
				Name:     existingVolsnapshot.Name,
				Kind:     existingVolsnapshot.Kind,
				APIGroup: &snapshotGroup,
			},
			Resources: pvcClient.VolumeResourceRequirements{
				Requests: pvcClient.ResourceList{
					pvcClient.ResourceStorage: *resource.NewQuantity(2, resource.BinarySI),
				},
			},
		},
	}

	if err := r.Create(ctx, vol); err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Println("already exist")
			err = nil
		} else {
			log.Log.Error(err, "failed to create PVC")
			fmt.Println("failed to create PVC", err)

			return ctrl.Result{}, err
		}
	}

	name := types.NamespacedName{Namespace: req.NamespacedName.Namespace, Name: vol.Name}

	if err := r.Get(ctx, name, vol); err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Was unable to find. Will requeue")

			return ctrl.Result{Requeue: true}, nil
		} else {
			volRestore.Status.Phase = "failed" //pvcClient.PersistentVolumeClaimPhase{}

			return ctrl.Result{}, err
		}
	}

	volRestore.Status.Phase = vol.Status.Phase

	copy(volRestore.Status.Condition, vol.Status.Conditions)
	err = r.Status().Update(context.Background(), &volRestore)
	if err != nil {
		return reconcile.Result{}, err
	}

	// just to update the status
	if volRestore.Status.Phase == "Pending" {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *VolrestoreReconciler) createSnapshotClient() (client *snapclientv1.SnapshotV1Client, err error) {
	client, err = snapclientv1.NewForConfigAndClient(r.Config, r.HTTPClient)
	if err != nil {
		log.Log.Error(err, "getting volumesnapshot client")
		return nil, err
	}

	return client, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolrestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&volv1.Volrestore{}).
		Complete(r)
}