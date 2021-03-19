package config_test

import (
	. "github.com/goharbor/harbor-operator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ovh/configstore"
)

var _ = Describe("NewConfigWithDefaults", func() {
	It("Should works", func() {
		Expect(NewConfigWithDefaults()).ToNot(BeNil())
	})
})

var _ = Describe("GetItem", func() {
	var store *configstore.Store
	var key string
	var defaultValue string

	BeforeEach(func() {
		store = NewConfigWithDefaults()
		defaultValue = "the-default-value"
	})

	Context("With an unknown key", func() {
		BeforeEach(func() {
			key = "unknown-key"
		})

		It("Should return the default value in parameter", func() {
			Expect(GetItem(store, key, defaultValue)).To(WithTransform(func(item configstore.Item) string {
				v, _ := item.Value()

				return v
			}, Equal(defaultValue)))
		})
	})

	Context("With a default key", func() {
		BeforeEach(func() {
			key = HarborClassKey
		})

		It("Should return the value from the store", func() {
			Expect(GetItem(store, key, defaultValue)).To(WithTransform(func(item configstore.Item) string {
				v, _ := item.Value()

				return v
			}, Equal(DefaultHarborClass)))
		})
	})

	Context("With a new key", func() {
		var value string

		BeforeEach(func() {
			value = "the-expected-value"
			key = "the-new-key"
			store.InMemory("test").Add(configstore.NewItem(key, value, 10))
		})

		It("Should return the value from the store", func() {
			Expect(GetItem(store, key, defaultValue)).To(WithTransform(func(item configstore.Item) string {
				v, _ := item.Value()

				return v
			}, Equal(value)))
		})
	})
})
