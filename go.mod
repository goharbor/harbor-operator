module github.com/goharbor/harbor-operator

go 1.14

require (
	cloud.google.com/go v0.58.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/alicebob/miniredis/v2 v2.14.1
	github.com/aws/aws-sdk-go v1.32.5
	github.com/go-kit/kit v0.10.0
	github.com/go-logr/logr v0.2.1-0.20200730175230-ee2de8da5be6
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/goharbor/harbor/src v0.0.0-20200728051037-1d361a89a056
	github.com/google/go-cmp v0.5.1
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/jackc/pgx/v4 v4.8.1
	github.com/jaegertracing/jaeger-lib v2.2.0+incompatible
	github.com/jetstack/cert-manager v0.16.1
	github.com/markbates/pkger v0.15.1
	github.com/minio/minio-go/v6 v6.0.57
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/ncw/swift v1.0.49
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/ovh/configstore v0.3.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/sethvargo/go-password v0.1.3
	github.com/sirupsen/logrus v1.6.0
	github.com/spotahome/redis-operator v1.0.0
	github.com/uber/jaeger-client-go v2.24.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	golang.org/x/tools v0.0.0-20200626171337-aa94e735be7f // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	k8s.io/api v0.19.0-rc.3
	k8s.io/apimachinery v0.19.0-rc.3
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.1-0.20200804124940-17eebbff0d48
	sigs.k8s.io/kustomize/kstatus v0.0.2
	sigs.k8s.io/yaml v1.2.0
)

replace k8s.io/client-go v11.0.0+incompatible => k8s.io/client-go v0.0.0-20200813012017-e7a1d9ada0d5
