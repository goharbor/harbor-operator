package config_test

import (
	"context"

	. "github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/config"
	"github.com/goharbor/harbor-operator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ovh/configstore"
)

var _ = Describe("New", func() {
	Context("With new store", func() {
		var store *configstore.Store
		var provider *configstore.InMemoryProvider

		BeforeEach(func() {
			store, provider = New(context.TODO(), "template", "name")
			Expect(store).ToNot(BeNil())
			Expect(provider).ToNot(BeNil())
		})

		Describe("Adding an config", func() {
			var key, value string

			BeforeEach(func() {
				key, value = "my-key", "my-value"
			})

			It("Should works", func() {
				provider.Add(configstore.NewItem(key, value, 10))

				defaultValue := ""
				Expect(defaultValue).ToNot(Equal(value))
				Expect(config.GetString(store, key, defaultValue)).To(Equal(value))
			})
		})
	})
})
