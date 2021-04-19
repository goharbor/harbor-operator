package config_test

import (
	"github.com/pkg/errors"

	. "github.com/goharbor/harbor-operator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ovh/configstore"
)

var _ = Describe("IsNotFound", func() {
	Context("With nil error", func() {
		It("Should return false", func() {
			Expect(IsNotFound(nil, "whatever")).To(BeFalse())
		})
	})

	Context("With random string error", func() {
		var err error

		BeforeEach(func() {
			err = errors.Errorf("a random error")
		})

		It("Should return false", func() {
			Expect(IsNotFound(err, "random")).To(BeFalse())
		})
	})

	Context("With string error", func() {
		var err error
		var key string

		BeforeEach(func() {
			key = "test"
			err = errors.Errorf("configstore: get '%s': no item found", key)
		})

		It("Should return false", func() {
			Expect(IsNotFound(err, key)).To(BeFalse())
		})
	})

	Context("With configstore", func() {
		var key string

		BeforeEach(func() {
			key = "test"
		})

		Describe("Getting item value", func() {
			It("Should return true", func() {
				_, err := configstore.GetItemValue(key)
				Expect(IsNotFound(err, key)).To(BeTrue())
			})
		})

		Describe("Getting item slice", func() {
			It("Should return true", func() {
				_, err := configstore.Filter().
					Slice(key).
					GetFirstItem()
				Expect(IsNotFound(err, key)).To(BeTrue())
			})
		})
	})
})
