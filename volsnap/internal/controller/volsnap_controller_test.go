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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	volv1 "neha-gupta1/volsnap/api/v1"
)

var _ = Describe("Volsnap Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource-1"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		volsnap := &volv1.Volsnap{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Volsnap")

			// create Volsnap
			err := k8sClient.Get(ctx, typeNamespacedName, volsnap)
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
				err := k8sClient.Create(ctx, resource)
				Expect(err).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &volv1.Volsnap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
			}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err != nil && errors.IsNotFound(err) {
				fmt.Println("Test successfully completed")
				err = nil
				return
			} else if err != nil {
				Expect(err).NotTo(HaveOccurred())
			} else {

				By("Cleanup the specific resource instance Volsnap")
				Eventually(k8sClient.Delete(ctx, resource)).Should(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")

			controllerReconciler := &VolsnapReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Get(ctx, typeNamespacedName, volsnap)).NotTo(HaveOccurred())
			SetDefaultEventuallyTimeout(time.Minute * 1)
			Eventually(volsnap.Status.RunningStatus).Should(Equal("Created"))
		},
		)

		It("should successfully reconcile the deleted resource", func() {
			By("Reconciling the deleted resource")

			resource := &volv1.Volsnap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
			}

			controllerReconciler := &VolsnapReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail if namespace not provided", func() {
			By("request without namespace")
			controllerReconciler := &VolsnapReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			resource := &volv1.Volsnap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nonamespace",
					Namespace: "default",
				},
			}

			err := k8sClient.Create(ctx, resource)
			if err != nil && errors.IsAlreadyExists(err) {
				err = nil
			}
			if err != nil {
				Expect(err).NotTo(HaveOccurred())
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: resourceName,
				},
			})

			Expect(err).Should(HaveOccurred())
		})
	})
})
