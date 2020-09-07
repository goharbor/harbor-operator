package graph_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/goharbor/harbor-operator/pkg/graph"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Walk a dependency manager", func() {
	var rm Manager
	var ctx context.Context

	BeforeEach(func() {
		rm, ctx = setupTest(context.TODO())
	})

	Context("With no resource", func() {
		It("Should not call any function", func() {
			Expect(rm.Run(ctx)).To(Succeed())
		})
	})

	Context("With a single resource", func() {
		var counter int32

		BeforeEach(func() {
			counter = 0

			add1 := func(ctx context.Context, resource Resource) error {
				defer GinkgoRecover()

				atomic.AddInt32(&counter, 1)

				return nil
			}

			err := rm.AddResource(ctx, &corev1.Secret{}, nil, add1)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should call the function only once", func() {
			Expect(rm.Run(ctx)).To(Succeed())

			Expect(counter).To(BeEquivalentTo(1))
		})
	})

	Context("With 4 isolated resources", func() {
		var counter int32

		BeforeEach(func() {
			counterMap := sync.Map{}

			add1 := func(ctx context.Context, resource Resource) error {
				defer GinkgoRecover()

				_, exists := counterMap.LoadOrStore(resource, true)
				Expect(exists).To(BeFalse())

				atomic.AddInt32(&counter, 1)

				return nil
			}

			Expect(rm.AddResource(ctx, &corev1.Namespace{}, nil, add1)).
				To(Succeed())

			Expect(rm.AddResource(ctx, &corev1.Secret{}, nil, add1)).
				To(Succeed())

			Expect(rm.AddResource(ctx, &corev1.ConfigMap{}, nil, add1)).
				To(Succeed())

			Expect(rm.AddResource(ctx, &corev1.Node{}, nil, add1)).
				To(Succeed())
		})

		It("Should call the function exatly 4 times", func() {
			err := rm.Run(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(counter).To(BeEquivalentTo(4))
		})
	})

	FContext("With a dependencies tree", func() {
		var expected int32
		var counter int32

		BeforeEach(func() {
			counter = 0
			expected = 0

			countAndCheck := func(expectedResource Resource, validIndexes ...interface{}) func(ctx context.Context, resource Resource) error {
				return func(ctx context.Context, resource Resource) error {
					defer GinkgoRecover()

					Expect(resource).To(Equal(expectedResource))
					Expect(atomic.AddInt32(&counter, 1)).
						To(WithTransform(
							func(v int32) int { return 1 + int((v-1)%expected) },
							BeElementOf(validIndexes...),
						))

					return nil
				}
			}

			ns := &corev1.Namespace{}
			Expect(rm.AddResource(ctx, ns, nil, countAndCheck(ns, 1))).
				To(Succeed())
			expected++

			secret := &corev1.Secret{}
			Expect(rm.AddResource(ctx, secret, []Resource{ns}, countAndCheck(secret, 2, 3))).
				To(Succeed())
			expected++

			cm := &corev1.ConfigMap{}
			Expect(rm.AddResource(ctx, cm, []Resource{ns}, countAndCheck(cm, 2, 3))).
				To(Succeed())
			expected++

			no := &corev1.Node{}
			Expect(rm.AddResource(ctx, no, []Resource{secret, cm}, countAndCheck(no, 4))).
				To(Succeed())
			expected++
		})

		It("Should call the function in the right order", func() {
			Expect(rm.Run(ctx)).To(Succeed())

			Expect(counter).To(Equal(expected))
		})

		It("Should accept multiple runs", func() {
			const runCount = 3

			for i := int32(1); i <= runCount; i++ {
				Expect(rm.Run(ctx)).To(Succeed())

				Expect(counter).To(Equal(i * expected))
			}
		})
	})

	Context("With errored node", func() {
		var expectedError error
		var resource Resource

		BeforeEach(func() {
			expectedError = errors.New("test error")

			raiseError := func(ctx context.Context, resource Resource) error {
				return expectedError
			}

			resource = &corev1.Namespace{}

			Expect(rm.AddResource(ctx, resource, nil, raiseError)).
				To(Succeed())
		})

		It("Should return the right error", func() {
			err := rm.Run(ctx)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(expectedError))
		})

		Describe("Linear graph", func() {
			BeforeEach(func() {
				fail := func(ctx context.Context, resource Resource) error {
					defer GinkgoRecover()

					Fail("func should not be executed")

					return nil
				}

				Expect(rm.AddResource(ctx, &corev1.Secret{}, []Resource{resource}, fail)).
					To(Succeed())
			})

			It("Should not trigger child nodes", func() {
				Expect(rm.Run(ctx)).ToNot(Succeed())
			})
		})

		Describe("Parallel graph", func() {
			BeforeEach(func() {
				watch := func(ctx context.Context, resource Resource) error {
					defer GinkgoRecover()

					select {
					case <-time.After(1500 * time.Millisecond):
						Fail("context not canceled after timeout")
					case <-ctx.Done():
						fmt.Fprintf(GinkgoWriter, "context canceled")
					}

					return nil
				}

				Expect(rm.AddResource(ctx, &corev1.Secret{}, []Resource{resource}, watch)).
					To(Succeed())
			})

			It("Should cancel context of sibling nodes", func() {
				Expect(rm.Run(ctx)).ToNot(Succeed())
			})
		})
	})
})
