module github.com/goharbor/harbor-operator

go 1.16

require (
	cloud.google.com/go v0.58.0 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/go-kit/kit v0.10.0
	github.com/go-logr/logr v0.4.0
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/goharbor/harbor/src v0.0.0-20210705083811-ddb66197697f
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/jaegertracing/jaeger-lib v2.2.0+incompatible
	github.com/jetstack/cert-manager v1.1.0
	github.com/markbates/pkger v0.15.1
	github.com/minio/minio-go/v6 v6.0.57
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.12.0
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/ovh/configstore v0.3.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/sethvargo/go-password v0.1.3
	github.com/sirupsen/logrus v1.7.0
	github.com/spotahome/redis-operator v1.0.0
	github.com/theupdateframework/notary v0.6.1
	github.com/uber/jaeger-client-go v2.24.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/zalando/postgres-operator v1.6.1
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/tools v0.1.2 // indirect
	k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.9.0-beta.6
	sigs.k8s.io/kustomize/kstatus v0.0.2
	sigs.k8s.io/yaml v1.2.0
)
