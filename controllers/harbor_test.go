/*
Copyright 2019 The Kubernetes Authors.

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
	"net/url"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gstruct "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

const (
	applyTimeoutInterval = 5 * time.Second
)

var _ = Context("Inside of a new namespace", func() {
	log := zap.LoggerTo(GinkgoWriter, true)
	ctx := logger.Context(log)
	ns := SetupTest(ctx)

	publicURL := url.URL{
		Scheme: "http",
		Host:   "the.dns",
	}

	Describe("Creating Harbor resources", func() {
		Context("with invalid version", func() {
			It("should raise an error", func() {
				harbor := &goharborv1alpha2.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "harbor-invalid-semver",
						Namespace: ns.Name,
					},
					Spec: goharborv1alpha2.HarborSpec{
						HarborVersion: "invalid-semver",
						PublicURL:     publicURL.String(),
					},
				}
				harbor.Default()

				Expect(k8sClient.Create(ctx, harbor)).Should(WithTransform(apierrs.IsInvalid, BeTrue()))
			})
		})

		Context("with invalid public url", func() {
			It("should raise an error", func() {
				harbor := &goharborv1alpha2.Harbor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "harbor-invalid-url",
						Namespace: ns.Name,
					},
					Spec: goharborv1alpha2.HarborSpec{
						HarborVersion: "1.9.1",
						PublicURL:     "123::bad::dns",
					},
				}
				harbor.Default()

				Expect(k8sClient.Create(ctx, harbor)).Should(WithTransform(apierrs.IsInvalid, BeTrue()))
			})
		})

		PContext("with valid spec", func() {
			It("should be handled", func() {
				harbor, key := newValidHarborTest(ns.Name)

				getHarbor := getResourceFunc(ctx, key, harbor, nil)

				Expect(k8sClient.Create(ctx, harbor)).To(Succeed())
				Eventually(getHarbor).Should(Succeed(), "harbor resource should exist")

				getConditions := func(harbor *goharborv1alpha2.Harbor) []goharborv1alpha2.Condition {
					return harbor.Status.Conditions
				}
				Eventually(getResourceFunc(ctx, key, harbor, getConditions), applyTimeoutInterval).Should(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Type":   BeEquivalentTo(goharborv1alpha2.AppliedConditionType),
					"Status": BeEquivalentTo(corev1.ConditionTrue),
				})), "harbor resource should be applied")
			})
		})
	})

	Describe("Handling resource events", func() {
		Context("Updating Harbor resource spec", func() {
			It("should update ObservedGeneration", func() {
				harbor, key := newValidHarborTest(ns.Name)

				getHarbor := getResourceFunc(ctx, key, harbor, nil)

				Expect(k8sClient.Create(ctx, harbor)).To(Succeed())
				Eventually(getHarbor).Should(Succeed(), "harbor resource should exist")

				Expect(k8sClient.Get(ctx, key, harbor)).To(Succeed())

				harbor.Spec.HarborVersion = fmt.Sprintf("%s-latest", harbor.Spec.HarborVersion)
				// Use Eventually since Operator may increase resourceVersion asynchronously
				Eventually(getUpdateFunc(ctx, harbor), applyTimeoutInterval).Should(Succeed(), "harbor resource should be updatable")

				Expect(k8sClient.Get(ctx, key, harbor)).To(Succeed())

				getObservedGeneration := func(harbor *goharborv1alpha2.Harbor) int64 {
					return harbor.Status.ObservedGeneration
				}
				Eventually(getResourceFunc(ctx, key, harbor, getObservedGeneration), applyTimeoutInterval).Should(BeNumerically(">=", harbor.GetGeneration()), "ObservedGeneration should math Generation")
			})

			It("should not update Generation", func() {
				const defaultGenerationNumber int64 = 1

				harbor, key := newValidHarborTest(ns.Name)

				Expect(k8sClient.Create(ctx, harbor)).To(Succeed())
				Consistently(getResourceFunc(ctx, key, harbor, metav1.Object.GetGeneration), applyTimeoutInterval).Should(Equal(defaultGenerationNumber), "harbor Generation should not be updated")
			})
		})
	})

	Describe("Deleting Harbor resource", func() {
		Context("with new resource", func() {
			It("should no more exists", func() {
				harbor, key := newValidHarborTest(ns.Name)

				getHarbor := getResourceFunc(ctx, key, harbor, nil)

				Expect(k8sClient.Create(ctx, harbor)).To(Succeed())
				Eventually(getHarbor).Should(Succeed(), "harbor resource should exist")

				Expect(k8sClient.Delete(ctx, harbor)).To(Succeed())
				Eventually(getHarbor).ShouldNot(Succeed(), "harbor resource should not exist")
			})
		})
	})
})

func getUpdateFunc(ctx context.Context, harbor *goharborv1alpha2.Harbor) func() error {
	return func() error {
		return k8sClient.Update(ctx, harbor)
	}
}

func getResourceFunc(ctx context.Context, key client.ObjectKey, obj runtime.Object, f interface{}) func() interface{} {
	fValue := reflect.ValueOf(f)

	return func() interface{} {
		err := k8sClient.Get(ctx, key, obj)

		if f == nil {
			return err
		}

		Expect(err).ToNot(HaveOccurred())

		return fValue.Call([]reflect.Value{reflect.ValueOf(obj)})[0].Interface()
	}
}

func newValidHarborTest(ns string) (*goharborv1alpha2.Harbor, client.ObjectKey) {
	name := newName("harbor")
	publicURL := url.URL{
		Scheme: "http",
		Host:   "the.dns",
	}

	harbor := &goharborv1alpha2.Harbor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.HarborSpec{
			HarborVersion:       "1.10.0",
			PublicURL:           publicURL.String(),
			AdminPasswordSecret: "admin-secret",
			Components: goharborv1alpha2.HarborComponents{
				Core: &goharborv1alpha2.CoreComponent{
					CoreConfig: goharborv1alpha2.CoreConfig{
						DatabaseSecret: "core-database-secret",
					},
				},
				JobService: &goharborv1alpha2.JobServiceComponent{
					JobServiceConfig: goharborv1alpha2.JobServiceConfig{
						RedisSecret: "jobservice-redis-secret",
					},
				},
				Portal:   &goharborv1alpha2.PortalComponent{},
				Registry: &goharborv1alpha2.RegistryComponent{},
			},
		},
	}
	harbor.Default()

	return harbor, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}
