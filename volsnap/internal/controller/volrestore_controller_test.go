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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	volv1 "neha-gupta1/volsnap/api/v1"
)

var _ = Describe("Volrestore Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		const notpresent = "notpresent"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		volrestore := &volv1.Volrestore{}
		volsnap := &volv1.Volsnap{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Volrestore")

			// create Volsnap
			typeSnapshot := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			err := k8sClient.Get(ctx, typeSnapshot, volsnap)
			if err != nil && errors.IsNotFound(err) {
				resource := &volv1.Volsnap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: volv1.VolsnapSpec{
						VolumeName: "csi-pvc",
					},
				}

				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			err = k8sClient.Get(ctx, typeNamespacedName, volrestore)
			if err != nil && errors.IsNotFound(err) {
				resource := &volv1.Volrestore{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: volv1.VolrestoreSpec{
						VolSnapName: typeSnapshot.Name,
					},
					// TODO(user): Specify other spec details if needed.
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &volv1.Volrestore{}

			for _, name := range []string{resourceName, notpresent} {
				typeNamespacedName := types.NamespacedName{
					Name:      name,
					Namespace: "default",
				}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				if err != nil && errors.IsNotFound(err) {
					err = nil
					return
				} else if err != nil {
					Expect(err).NotTo(HaveOccurred())
				} else {

					By("Cleanup the specific resource instance Volrestore")
					Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
				}
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &VolrestoreReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail if volSnap not present", func() {
			By("providing invalid restore name")
			controllerReconciler := &VolrestoreReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			resource := &volv1.Volrestore{
				ObjectMeta: metav1.ObjectMeta{
					Name:      notpresent,
					Namespace: "default",
				},
				Spec: volv1.VolrestoreSpec{
					VolSnapName: notpresent,
				},
			}
			Expect(k8sClient.Create(ctx, resource)).NotTo(HaveOccurred())

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      notpresent,
					Namespace: "default",
				},
			})

			Expect(err).Should(HaveOccurred())
		})

		It("should fail if namespace not provided", func() {
			By("request without namespace")
			controllerReconciler := &VolrestoreReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			resource := &volv1.Volrestore{
				ObjectMeta: metav1.ObjectMeta{
					Name:      notpresent,
					Namespace: "default",
				},
				Spec: volv1.VolrestoreSpec{
					VolSnapName: notpresent,
				},
			}
			Expect(k8sClient.Create(ctx, resource)).NotTo(HaveOccurred())

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: notpresent,
				},
			})

			Expect(err).Should(HaveOccurred())
		})
	})
})
