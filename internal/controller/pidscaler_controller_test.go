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
	"github.com/timson/pidhpa-operator/internal/storage"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pidscalerv1 "github.com/timson/pidhpa-operator/api/v1"
)

var _ = Describe("PIDScaler Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		pidscaler := &pidscalerv1.PIDScaler{}
		var controllerReconciler *PIDScalerReconciler

		BeforeEach(func() {
			By("creating the custom resource for the Kind PIDScaler")
			err := k8sClient.Get(ctx, typeNamespacedName, pidscaler)
			if err != nil && errors.IsNotFound(err) {
				resource := &pidscalerv1.PIDScaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: pidscalerv1.PIDScalerSpec{
						Target: pidscalerv1.TargetSettings{
							Deployment:  "test-deployment",
							Namespace:   "default",
							MinReplicas: 1,
							MaxReplicas: 10,
						},
						PID: pidscalerv1.PIDSettings{
							Kp:              "0.1",
							Ki:              "0.01",
							Kd:              "0.001",
							ReferenceSignal: 100,
						},
						Kafka: pidscalerv1.KafkaSettings{
							Topic:   "test-topic",
							Group:   "test-group",
							Brokers: []string{"broker1:9092", "broker2:9092"},
						},
						CooldownTimeout: 60,
						Interval:        30,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
			By("initializing the PIDScalerReconciler")
			controllerReconciler = &PIDScalerReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				Log:             zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)),
				Storage:         storage.NewPIDScalerStorage(),
				OperatorContext: ctx,
				wg:              &sync.WaitGroup{},
			}
		})

		AfterEach(func() {
			resource := &pidscalerv1.PIDScaler{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance PIDScaler")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
		It("should handle missing PIDScaler resource", func() {
			req := reconcile.Request{NamespacedName: typeNamespacedName}
			_, err := controllerReconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should validate the created PIDScaler resource", func() {
			created := &pidscalerv1.PIDScaler{}
			err := k8sClient.Get(ctx, typeNamespacedName, created)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Spec.Target.MinReplicas).To(Equal(int32(1)))
			Expect(created.Spec.Target.MaxReplicas).To(Equal(int32(10)))
		})
	})
})
