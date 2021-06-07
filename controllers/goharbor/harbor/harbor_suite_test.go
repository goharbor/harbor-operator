package harbor_test

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harbor"
	"github.com/goharbor/harbor-operator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func TestHarbor(t *testing.T) {
	t.Parallel()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Harbor Suite")
}

func fileString(filePath string) string {
	content, err := ioutil.ReadFile(filePath)
	Expect(err).NotTo(HaveOccurred())

	return strings.TrimSpace(string(content))
}

func makeReconciler(ctx context.Context) *harbor.Reconciler {
	name := controllers.Harbor.String()
	configStore := config.NewConfigWithDefaults()
	configStore.Env(name)
	configStore.InitFromEnvironment()

	h, err := harbor.New(ctx, configStore)
	Expect(err).NotTo(HaveOccurred())

	r := h.(*harbor.Reconciler)

	sch := runtime.NewScheme()
	_ = goharborv1.AddToScheme(sch)

	r.Controller.Scheme = sch

	return r
}

func getSpec(file string) *goharborv1.Harbor {
	input := fileString(file)

	sch := runtime.NewScheme()
	_ = goharborv1.AddToScheme(sch)
	decoder := serializer.NewCodecFactory(sch).UniversalDeserializer()

	obj, _, err := decoder.Decode([]byte(input), nil, nil)
	Expect(err).NotTo(HaveOccurred())

	h, ok := obj.(*goharborv1.Harbor)
	Expect(ok).To(BeTrue())

	return h
}
