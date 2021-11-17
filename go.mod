module github.com/goharbor/harbor-operator

go 1.16

require (
	cloud.google.com/go v0.58.0 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/go-kit/kit v0.10.0
	github.com/go-logr/logr v0.4.0
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/goharbor/harbor/src v0.0.0-20211025104526-d4affc2eba6d
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/jaegertracing/jaeger-lib v2.2.0+incompatible
	github.com/jetstack/cert-manager v1.1.0
	github.com/markbates/pkger v0.17.1
	github.com/minio/minio-go/v6 v6.0.57
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/ovh/configstore v0.3.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/sethvargo/go-password v0.1.3
	github.com/sirupsen/logrus v1.8.1
	github.com/spotahome/redis-operator v1.0.0
	github.com/szlabs/redis-operator v1.0.1 // indirect
	github.com/theupdateframework/notary v0.6.1
	github.com/uber/jaeger-client-go v2.24.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/zalando/postgres-operator v1.6.1
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.3
	k8s.io/apiextensions-apiserver v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.10.2
	sigs.k8s.io/kustomize/kstatus v0.0.2
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/spotahome/redis-operator v1.0.0 => github.com/szlabs/redis-operator v1.0.1

replace github.com/szlabs/redis-operator v1.0.1 => github.com/spotahome/redis-operator v1.0.0
