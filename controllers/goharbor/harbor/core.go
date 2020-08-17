package harbor

import (
	"context"
	"net/url"
	"time"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

type CoreCSRF graph.Resource

func (r *Reconciler) AddCoreCSRF(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreCSRF, error) {
	csrf, err := r.GetCSRFSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	csrfRes, err := r.AddSecretToManage(ctx, csrf)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return CoreCSRF(csrfRes), nil
}

type CoreSecret graph.Resource

func (r *Reconciler) AddCoreSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreSecret, error) {
	secret, err := r.GetCoreSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	secretRes, err := r.AddSecretToManage(ctx, secret)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return CoreSecret(secretRes), nil
}

type CoreTokenCertificate graph.Resource

func (r *Reconciler) AddCoreTokenCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreTokenCertificate, error) {
	certificate, err := r.GetCoreTokenCertificate(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certificateRes, err := r.AddCertificateToManage(ctx, certificate)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return CoreTokenCertificate(certificateRes), nil
}

type CoreAdminPassword graph.Resource

func (r *Reconciler) AddCoreAdminPassword(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreAdminPassword, error) {
	adminPassword, err := r.GetCoreAdminPassword(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	adminPasswordRes, err := r.AddImmutableSecretToManage(ctx, adminPassword)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return CoreAdminPassword(adminPasswordRes), nil
}

type CoreEncryptionKey graph.Resource

func (r *Reconciler) AddCoreEncryptionKey(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreEncryptionKey, error) {
	encryptionKey, err := r.GetCoreEncryptionKey(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	encryptionKeyRes, err := r.AddSecretToManage(ctx, encryptionKey)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return CoreEncryptionKey(encryptionKeyRes), nil
}

type CoreInternalCertificate graph.Resource

func (r *Reconciler) AddCoreInternalCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (CoreInternalCertificate, error) {
	cert, err := r.GetInternalTLSCertificate(ctx, harbor, harbormetav1.CoreTLS)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return CoreInternalCertificate(certRes), nil
}

func (r *Reconciler) AddCoreConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (CoreInternalCertificate, CoreCSRF, CoreTokenCertificate, CoreSecret, CoreAdminPassword, CoreEncryptionKey, error) {
	certificate, err := r.AddCoreInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.Wrap(err, "certificate")
	}

	csrf, err := r.AddCoreCSRF(ctx, harbor)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.Wrap(err, "csrf")
	}

	secret, err := r.AddCoreSecret(ctx, harbor)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.Wrap(err, "secret")
	}

	tokenCertificate, err := r.AddCoreTokenCertificate(ctx, harbor)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.Wrap(err, "secret")
	}

	adminPassword, err := r.AddCoreAdminPassword(ctx, harbor)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.Wrap(err, "admin password")
	}

	encryptionKey, err := r.AddCoreEncryptionKey(ctx, harbor)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.Wrap(err, "encryption key")
	}

	return certificate, csrf, tokenCertificate, secret, adminPassword, encryptionKey, nil
}

type Core graph.Resource

func (r *Reconciler) AddCore(ctx context.Context, harbor *goharborv1alpha2.Harbor, coreCertificate CoreInternalCertificate, registryAuth RegistryAuthSecret, csrf CoreCSRF, tokenCertificate CoreTokenCertificate, secret CoreSecret, adminPassword CoreAdminPassword, encryptionKey CoreEncryptionKey) (Core, error) {
	core, err := r.GetCore(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	coreRes, err := r.AddBasicResource(ctx, core, coreCertificate, registryAuth, csrf, tokenCertificate, secret, adminPassword, encryptionKey)

	return Core(coreRes), errors.Wrap(err, "add")
}

const (
	CoreAdminPasswordLength      = 32
	CoreAdminPasswordNumDigits   = 5
	CoreAdminPasswordNumSpecials = 10
)

func (r *Reconciler) GetCoreAdminPassword(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "admin-password")
	namespace := harbor.GetNamespace()

	password, err := password.Generate(CoreAdminPasswordLength, CoreAdminPasswordNumDigits, CoreAdminPasswordNumSpecials, false, true)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate admin password")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varTrue,
		Type:      harbormetav1.SecretTypeSingle,
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: password,
		},
	}, nil
}

const (
	CoreSecretPasswordLength      = 128
	CoreSecretPasswordNumDigits   = 16
	CoreSecretPasswordNumSpecials = 48
)

func (r *Reconciler) GetCoreSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "secret")
	namespace := harbor.GetNamespace()

	secret, err := password.Generate(CoreSecretPasswordLength, CoreSecretPasswordNumDigits, CoreSecretPasswordNumSpecials, false, true)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate secret")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varTrue,
		Type:      harbormetav1.SecretTypeSingle,
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: secret,
			corev1.BasicAuthUsernameKey:  ChartMuseumAuthenticationUsername,
			corev1.BasicAuthPasswordKey:  secret,
		},
	}, nil
}

const (
	CoreTokenServiceDefaultKeySize             = 4096
	CoreTokenServiceDefaultCertificateDuration = 3 * 30 * 24 * time.Hour
)

func (r *Reconciler) GetCoreTokenCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*certv1.Certificate, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "tokencert")
	namespace := harbor.GetNamespace()

	publicDNS, err := url.Parse(harbor.Spec.ExternalURL)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse external url")
	}

	secretName := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "tokencert")

	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: certv1.CertificateSpec{
			Duration: &metav1.Duration{
				Duration: CoreTokenServiceDefaultCertificateDuration,
			},
			KeyAlgorithm: certv1.RSAKeyAlgorithm,
			KeySize:      CoreTokenServiceDefaultKeySize,
			DNSNames:     []string{publicDNS.Host},
			SecretName:   secretName,
			Usages:       []certv1.KeyUsage{certv1.UsageSigning},
			IssuerRef:    harbor.Spec.Core.TokenIssuer,
		},
	}, nil
}

const (
	EncryptionKeyLength      = 128
	EncryptionKeyNumDigits   = 16
	EncryptionKeyNumSpecials = 48
)

func (r *Reconciler) GetCoreEncryptionKey(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "encryptionkey")
	namespace := harbor.GetNamespace()

	key, err := password.Generate(CoreSecretPasswordLength, CoreSecretPasswordNumDigits, CoreSecretPasswordNumSpecials, false, true)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate key")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varTrue,
		Type:      harbormetav1.SecretTypeSingle,
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: key,
		},
	}, nil
}

const (
	CSRFSecretPasswordLength      = 32
	CSRFSecretPasswordNumDigits   = 5
	CSRFSecretPasswordNumSpecials = 0
)

func (r *Reconciler) GetCSRFSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "csrf")
	namespace := harbor.GetNamespace()

	csrf, err := password.Generate(CSRFSecretPasswordLength, CSRFSecretPasswordNumDigits, CSRFSecretPasswordNumSpecials, false, true)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate csrf")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varTrue,
		Type:      harbormetav1.SecretTypeSingle,
		StringData: map[string]string{
			harbormetav1.CSRFSecretKey: csrf,
		},
	}, nil
}

func (r *Reconciler) GetCore(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.Core, error) { // nolint:funlen
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	credentials := goharborv1alpha2.CoreComponentsRegistryCredentialsSpec{
		Username:    RegistryAuthenticationUsername,
		PasswordRef: r.NormalizeName(ctx, harbor.GetName(), controllers.Registry.String(), "basicauth"),
	}

	registryCtlURL := (&url.URL{
		Scheme: harbor.Spec.InternalTLS.GetScheme(),
		Host:   r.NormalizeName(ctx, harbor.GetName(), controllers.RegistryController.String()),
	}).String()
	registryURL := (&url.URL{
		Scheme: harbor.Spec.InternalTLS.GetScheme(),
		Host:   r.NormalizeName(ctx, harbor.GetName(), controllers.Registry.String()),
	}).String()
	portalURL := (&url.URL{
		Scheme: harbor.Spec.InternalTLS.GetScheme(),
		Host:   r.NormalizeName(ctx, harbor.GetName(), controllers.Portal.String()),
	}).String()

	var chartmuseum *goharborv1alpha2.CoreComponentsChartRepositorySpec

	if harbor.Spec.ChartMuseum != nil {
		chartmuseumURL := (&url.URL{
			Scheme: harbor.Spec.InternalTLS.GetScheme(),
			Host:   r.NormalizeName(ctx, harbor.GetName(), controllers.ChartMuseum.String()),
		}).String()
		chartmuseum = &goharborv1alpha2.CoreComponentsChartRepositorySpec{
			URL: chartmuseumURL,
		}
	}

	var trivy *goharborv1alpha2.CoreComponentsTrivySpec

	if harbor.Spec.Trivy != nil {
		trivyURL := (&url.URL{
			Scheme: harbor.Spec.InternalTLS.GetScheme(),
			Host:   r.NormalizeName(ctx, harbor.GetName(), controllers.Trivy.String()),
		}).String()
		trivy = &goharborv1alpha2.CoreComponentsTrivySpec{
			AdapterURL: trivyURL,
			URL:        trivyURL,
		}
	}

	adminPasswordRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "admin-password")
	coreSecretRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "secret")
	encryptionKeyRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "encryptionkey")
	csrfRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "csrf")
	jobserviceURL := (&url.URL{
		Scheme: harbor.Spec.InternalTLS.GetScheme(),
		Host:   r.NormalizeName(ctx, harbor.GetName(), controllers.JobService.String()),
	}).String()
	jobserviceSecretRef := r.NormalizeName(ctx, harbor.GetName(), controllers.JobService.String(), "secret")
	tokenServiceURL := (&url.URL{
		Scheme: harbor.Spec.InternalTLS.GetScheme(),
		Host:   r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String()),
		Path:   "/service/token",
	}).String()
	tokenCertificateRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "tokencert")

	registryRedis := harbor.Spec.RedisConnection(harbormetav1.RegistryRedis)

	coreRedis := harbor.Spec.RedisConnection(harbormetav1.CoreRedis)

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.CoreTLS))

	var publicCertificateRef string

	if harbor.Spec.Expose.Core.TLS.Enabled() {
		publicCertificateRef = harbor.Spec.Expose.Core.TLS.CertificateRef
	}

	storage, err := harbor.Spec.Database.GetPostgresqlConnection(harbormetav1.CoreComponent)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get database configuration")
	}

	return &goharborv1alpha2.Core{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.CoreSpec{
			ComponentSpec: harbor.Spec.Registry.ComponentSpec,
			Components: goharborv1alpha2.CoreComponentsSpec{
				Registry: goharborv1alpha2.CoreComponentsRegistrySpec{
					RegistryControllerConnectionSpec: goharborv1alpha2.RegistryControllerConnectionSpec{
						RegistryURL:   registryURL,
						ControllerURL: registryCtlURL,
						Credentials:   credentials,
					},
					Redis:               &registryRedis,
					StorageProviderName: harbor.Spec.ImageChartStorage.ProviderName(),
				},
				JobService: goharborv1alpha2.CoreComponentsJobServiceSpec{
					URL:       jobserviceURL,
					SecretRef: jobserviceSecretRef,
				},
				Portal: goharborv1alpha2.CoreComponentPortalSpec{
					URL: portalURL,
				},
				ChartRepository: chartmuseum,
				TokenService: goharborv1alpha2.CoreComponentsTokenServiceSpec{
					URL:            tokenServiceURL,
					CertificateRef: tokenCertificateRef,
				},
				Trivy: trivy,
				TLS:   tls,
			},
			CoreConfig: goharborv1alpha2.CoreConfig{
				AdminInitialPasswordRef: adminPasswordRef,
				SecretRef:               coreSecretRef,
				PublicCertificateRef:    publicCertificateRef,
			},
			CSRFKeyRef: csrfRef,
			Database: goharborv1alpha2.CoreDatabaseSpec{
				PostgresConnectionWithParameters: *storage,
				EncryptionKeyRef:                 encryptionKeyRef,
			},
			ExternalEndpoint: harbor.Spec.ExternalURL,
			Redis: goharborv1alpha2.CoreRedisSpec{
				RedisConnection: coreRedis,
			},
			Proxy: harbor.Spec.Proxy,
		},
	}, nil
}
