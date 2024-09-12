module github.com/goharbor/harbor-operator

go 1.22.5

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/containers/image/v5 v5.16.1
	github.com/go-kit/kit v0.10.0
	github.com/go-logr/logr v1.2.4
	github.com/go-openapi/runtime v0.21.0
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/goharbor/go-client v0.26.2
	github.com/goharbor/harbor/src v0.0.0-20220526154154-b0506782b47d
	github.com/jaegertracing/jaeger-lib v2.2.0+incompatible
	github.com/jetstack/cert-manager v1.1.0
	github.com/markbates/pkger v0.17.1
	github.com/minio/minio-go/v6 v6.0.57
	github.com/onsi/ginkgo/v2 v2.6.0
	github.com/onsi/gomega v1.24.1
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/ovh/configstore v0.3.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.14.0
	github.com/sethvargo/go-password v0.1.3
	github.com/sirupsen/logrus v1.8.1
	github.com/spotahome/redis-operator v1.1.1
	github.com/stretchr/testify v1.8.2
	github.com/theupdateframework/notary v0.6.1
	github.com/uber/jaeger-client-go v2.24.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/umisama/go-regexpcache v0.0.0-20150417035358-2444a542492f
	github.com/zalando/postgres-operator v1.6.1
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.1.0
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.26.1
	k8s.io/apiextensions-apiserver v0.26.1
	k8s.io/apimachinery v0.26.1
	k8s.io/client-go v0.26.1
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.14.6
	sigs.k8s.io/kustomize/kstatus v0.0.2
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/beego/beego v1.12.9 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/go v0.0.0-20160303222718-d30aec9fd63c // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/analysis v0.20.1 // indirect
	github.com/go-openapi/errors v0.20.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/loads v0.21.0 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/strfmt v0.21.0 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-openapi/validate v0.20.3 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gobuffalo/here v0.6.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.9.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.1.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid v1.2.3 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.2 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2 // indirect
	github.com/minio/md5-simd v1.1.0 // indirect
	github.com/minio/sha256-simd v0.1.1 // indirect
	github.com/mitchellh/copystructure v1.1.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/motomux/pretty v0.0.0-20161209205251-b2aad2c9a95d // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.mongodb.org/mongo-driver v1.7.3 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.35.0 // indirect
	go.opentelemetry.io/otel v1.12.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.12.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.12.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.12.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.12.0 // indirect
	go.opentelemetry.io/otel/metric v0.31.0 // indirect
	go.opentelemetry.io/otel/sdk v1.12.0 // indirect
	go.opentelemetry.io/otel/trace v1.12.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/oauth2 v0.4.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/grpc v1.54.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/component-base v0.26.1 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280 // indirect
	k8s.io/utils v0.0.0-20221128185143-99ec85e7a448 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)
