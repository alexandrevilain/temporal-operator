// Licensed to Alexandre VILAIN under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Alexandre VILAIN licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package reconciler_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	"github.com/alexandrevilain/temporal-operator/pkg/discovery"
	"github.com/alexandrevilain/temporal-operator/pkg/fake"
	"github.com/alexandrevilain/temporal-operator/pkg/reconciler"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Reconciler", func() {
	Describe("BuildersReconciler", func() {
		var deploy *appsv1.Deployment
		var builders []resource.Builder
		var rec *reconciler.BuildersReconciler
		var owner *appsv1.Deployment

		BeforeEach(func() {
			builder := fake.NewDeploymentBuilder(fmt.Sprintf("deploy-%d", rand.Int31()), "default") //nolint:gosec
			deploy = builder.Build().(*appsv1.Deployment)
			_ = builder.Update(deploy)

			builders = []resource.Builder{
				builder,
			}

			owner = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-owner",
					Namespace: "default",
				},
			}

			discoveryManager, err := discovery.NewManager(cfg, c.Scheme())
			Expect(err).ToNot(HaveOccurred())

			rec = &reconciler.BuildersReconciler{
				Client:    c,
				Scheme:    c.Scheme(),
				Recorder:  record.NewFakeRecorder(512),
				Discovery: discoveryManager,
			}
		})

		It("creates a new object if one doesn't exists", func() {
			statuses, reconcileAfter, err := rec.Reconcile(context.TODO(), owner, builders)

			By("returning no error")
			Expect(err).NotTo(HaveOccurred())
			By("returning no reconcile after")
			Expect(reconcileAfter).To(Equal(time.Duration(0)))
			By("returning statuses")
			Expect(statuses).To(HaveLen(1))

			By("actually having the deployment created")
			fetched := &appsv1.Deployment{}
			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(deploy), fetched)).To(Succeed())

			By("being updated by builder")
			Expect(fetched.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(fetched.Spec.Template.Spec.Containers[0].Name).To(Equal(deploy.Spec.Template.Spec.Containers[0].Name))
			Expect(fetched.Spec.Template.Spec.Containers[0].Image).To(Equal(deploy.Spec.Template.Spec.Containers[0].Image))
		})

		It("updates existing object", func() {
			var scale int32 = 2
			statuses, reconcileAfter, err := rec.Reconcile(context.TODO(), owner, builders)
			By("returning no error")
			Expect(err).NotTo(HaveOccurred())
			By("returning no reconcile after")
			Expect(reconcileAfter).To(Equal(time.Duration(0)))
			By("returning statuses")
			Expect(statuses).To(HaveLen(1))

			fakeDepBuilder := builders[0].(*fake.DeploymentBuilder)
			fakeDepBuilder.MutateObject = func(o client.Object) {
				deploy := o.(*appsv1.Deployment)
				deploy.Spec.Replicas = &scale
			}

			statuses, reconcileAfter, err = rec.Reconcile(context.TODO(), owner, builders)
			By("returning no error")
			Expect(err).NotTo(HaveOccurred())
			By("returning no reconcile after")
			Expect(reconcileAfter).To(Equal(time.Duration(0)))
			By("returning statuses")
			Expect(statuses).To(HaveLen(1))

			By("actually having the deployment scaled")
			fetched := &appsv1.Deployment{}
			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(deploy), fetched)).To(Succeed())
			Expect(*fetched.Spec.Replicas).To(Equal(scale))
		})
	})
})
