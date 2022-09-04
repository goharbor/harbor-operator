package harbor

import (
	"context"
	"fmt"
	"time"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	ConfigRegistryEncryptionCostKey = "registry-encryption-cost"
)

const (
	RegistryAuthRealm = "harbor-registry-basic-realm"
)

const (
	defaultUploadPurgingAge      = time.Hour * 168
	defaultUploadPurgingInterval = time.Hour * 24
)

var (
	varTrue  = true
	varFalse = false
)

type RegistryAuthSecret graph.Resource

func (r *Reconciler) AddRegistryAuthenticationSecret(ctx context.Context, harbor *goharborv1.Harbor) (RegistryAuthSecret, error) {
	authSecret, err := r.GetRegistryAuthenticationSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	authSecretRes, err := r.AddSecretToManage(ctx, authSecret)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return RegistryAuthSecret(authSecretRes), nil
}

func (r *Reconciler) AddRegistryConfigurations(ctx context.Context, harbor *goharborv1.Harbor, tlsIssuer InternalTLSIssuer) (RegistryInternalCertificate, RegistryAuthSecret, RegistryHTTPSecret, error) {
	certificate, err := r.AddRegistryInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "certificate")
	}

	authSecret, err := r.AddRegistryAuthenticationSecret(ctx, harbor)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "authentication secret")
	}

	httpSecret, err := r.AddRegistryHTTPSecret(ctx, harbor)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "http secret")
	}

	return certificate, authSecret, httpSecret, nil
}

type Registry graph.Resource

func (r *Reconciler) AddRegistry(ctx context.Context, harbor *goharborv1.Harbor, certificate RegistryInternalCertificate, authSecret RegistryAuthSecret, httpSecret RegistryHTTPSecret) (Registry, error) {
	registry, err := r.GetRegistry(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	registryRes, err := r.AddBasicResource(ctx, registry, certificate, authSecret, httpSecret)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return Registry(registryRes), nil
}

type RegistryHTTPSecret graph.Resource

func (r *Reconciler) AddRegistryHTTPSecret(ctx context.Context, harbor *goharborv1.Harbor) (RegistryHTTPSecret, error) {
	httpSecret, err := r.GetRegistryHTTPSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	httpSecretRes, err := r.AddSecretToManage(ctx, httpSecret)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return RegistryHTTPSecret(httpSecretRes), nil
}

type RegistryInternalCertificate graph.Resource

func (r *Reconciler) AddRegistryInternalCertificate(ctx context.Context, harbor *goharborv1.Harbor, tlsIssuer InternalTLSIssuer) (RegistryInternalCertificate, error) {
	cert, err := r.GetInternalTLSCertificate(ctx, harbor, harbormetav1.RegistryTLS)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return RegistryInternalCertificate(certRes), nil
}

const (
	// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/utils/configs.py#L14
	RegistryAuthenticationUsername = "harbor_registry_user"

	RegistryAuthenticationPasswordLength      = 32
	RegistryAuthenticationPasswordNumDigits   = 10
	RegistryAuthenticationPasswordNumSpecials = 10
)

func (r *Reconciler) GetRegistryAuthenticationSecret(ctx context.Context, harbor *goharborv1.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.Registry.String(), "basicauth")
	namespace := harbor.GetNamespace()

	password, err := password.Generate(RegistryAuthenticationPasswordLength, RegistryAuthenticationPasswordNumDigits, RegistryAuthenticationPasswordNumSpecials, false, true)
	if err != nil {
		return nil, errors.Wrap(err, "generate password")
	}

	cost, err := r.ConfigStore.GetItemValueInt(ConfigRegistryEncryptionCostKey)
	if err != nil {
		if !config.IsNotFound(err, ConfigRegistryEncryptionCostKey) {
			return nil, errors.Wrap(err, "cannot get encryption cost")
		}

		cost = int64(bcrypt.DefaultCost)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), int(cost))
	if err != nil {
		return nil, errors.Wrap(err, "cannot encrypt password")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varFalse,
		Type:      harbormetav1.SecretTypeHTPasswd,
		StringData: map[string]string{
			harbormetav1.HTPasswdFileName: fmt.Sprintf("%s:%s", RegistryAuthenticationUsername, string(hashedPassword)),
			harbormetav1.SharedSecretKey:  password,
		},
	}, nil
}

const (
	RegistrySecretPasswordLength      = 128
	RegistrySecretPasswordNumDigits   = 16
	RegistrySecretPasswordNumSpecials = 48
)

func (r *Reconciler) GetRegistryHTTPSecret(ctx context.Context, harbor *goharborv1.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.Registry.String(), "http")
	namespace := harbor.GetNamespace()

	secret, err := password.Generate(RegistrySecretPasswordLength, RegistrySecretPasswordNumDigits, RegistrySecretPasswordNumSpecials, false, true)
	if err != nil {
		return nil, errors.Wrap(err, "generate secret")
	}

	encodedSecret, err := yaml.Marshal(secret)
	if err != nil {
		return nil, serrors.UnrecoverrableError(err, serrors.OperatorReason, "cannot encode secret")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varTrue,
		Type:      harbormetav1.SecretTypeRegistry,
		StringData: map[string]string{
			harbormetav1.RegistryHTTPSecret: string(encodedSecret),
		},
	}, nil
}

func (r *Reconciler) GetRegistry(ctx context.Context, harbor *goharborv1.Harbor) (*goharborv1.Registry, error) { //nolint:funlen
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	authenticationSecretName := r.NormalizeName(ctx, harbor.GetName(), controllers.Registry.String(), "basicauth")
	httpSecretName := r.NormalizeName(ctx, harbor.GetName(), controllers.Registry.String(), "http")

	redis := harbor.Spec.RedisConnection(harbormetav1.RegistryRedis)

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.RegistryTLS))

	var httpDebug *goharborv1.RegistryHTTPDebugSpec
	if harbor.Spec.Registry.Metrics.IsEnabled() {
		httpDebug = &goharborv1.RegistryHTTPDebugSpec{
			Port: harbor.Spec.Registry.Metrics.Port,
			Prometheus: goharborv1.RegistryHTTPDebugPrometheusSpec{
				Enabled: true,
				Path:    harbor.Spec.Registry.Metrics.Path,
			},
		}
	}

	return &goharborv1.Registry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: version.SetVersion(map[string]string{
				harbormetav1.NetworkPoliciesAnnotationName: harbormetav1.NetworkPoliciesAnnotationDisabled,
			}, harbor.Spec.Version),
		},
		Spec: goharborv1.RegistrySpec{
			ComponentSpec: harbor.GetComponentSpec(ctx, harbormetav1.RegistryComponent),
			RegistryConfig01: goharborv1.RegistryConfig01{
				Log: goharborv1.RegistryLogSpec{
					AccessLog: goharborv1.RegistryAccessLogSpec{
						Disabled: false,
					},
					Level: harbor.Spec.LogLevel.Registry(),
				},
				Authentication: goharborv1.RegistryAuthenticationSpec{
					HTPasswd: &goharborv1.RegistryAuthenticationHTPasswdSpec{
						Realm:     RegistryAuthRealm,
						SecretRef: authenticationSecretName,
					},
				},
				Validation: goharborv1.RegistryValidationSpec{
					Disabled: true,
				},
				Middlewares: goharborv1.RegistryMiddlewaresSpec{
					Storage: harbor.Spec.Registry.StorageMiddlewares,
				},
				HTTP: goharborv1.RegistryHTTPSpec{
					Debug:        httpDebug,
					RelativeURLs: harbor.Spec.Registry.RelativeURLs,
					SecretRef:    httpSecretName,
					TLS:          tls,
				},
				Storage: goharborv1.RegistryStorageSpec{
					Driver: r.RegistryStorage(ctx, harbor),
					Cache: goharborv1.RegistryStorageCacheSpec{
						Blobdescriptor: "redis",
					},
					Redirect: harbor.Spec.ImageChartStorage.Redirect,
					Maintenance: goharborv1.RegistryStorageMaintenanceSpec{
						UploadPurging: goharborv1.RegistryStorageMaintenanceUploadPurgingSpec{
							Enabled: true,
							Age: &metav1.Duration{
								Duration: defaultUploadPurgingAge,
							},
							Interval: &metav1.Duration{
								Duration: defaultUploadPurgingInterval,
							},
						},
					},
				},
				Redis: &goharborv1.RegistryRedisSpec{
					RedisConnection: redis,
				},
			},
			CertificateInjection: harbor.Spec.Registry.CertificateInjection,
			Proxy:                harbor.GetComponentProxySpec(harbormetav1.RegistryComponent),
			Network:              harbor.Spec.Network,
			Trace:                harbor.Spec.Trace,
		},
	}, nil
}
