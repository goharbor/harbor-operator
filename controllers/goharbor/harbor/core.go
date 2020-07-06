package harbor

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
)

type CoreCSRF graph.Resource

func (r *Reconciler) AddCoreCSRF(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreCSRF, error) {
	csrf, err := r.GetCSRFSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get secret")
	}

	csrfRes, err := r.AddSecretToManage(ctx, csrf)
	if err != nil {
		return nil, errors.Wrap(err, "cannot add secret")
	}

	return CoreCSRF(csrfRes), nil
}

type CoreSecret graph.Resource

func (r *Reconciler) AddCoreSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreSecret, error) {
	secret, err := r.GetCoreSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get secret")
	}

	secretRes, err := r.AddSecretToManage(ctx, secret)
	if err != nil {
		return nil, errors.Wrap(err, "cannot add secret")
	}

	return CoreSecret(secretRes), nil
}

type CoreTokenCertificate graph.Resource

func (r *Reconciler) AddCoreTokenCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreTokenCertificate, error) {
	certificate, err := r.GetCoreTokenCertificate(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get certificate")
	}

	certificateRes, err := r.AddCertificateToManage(ctx, certificate)
	if err != nil {
		return nil, errors.Wrap(err, "cannot add certificate")
	}

	return CoreTokenCertificate(certificateRes), nil
}

type CoreAdminPassword graph.Resource

func (r *Reconciler) AddCoreAdminPassword(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreAdminPassword, error) {
	adminPassword, err := r.GetCoreAdminPassword(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get secret")
	}

	adminPasswordRes, err := r.AddSecretToManage(ctx, adminPassword)
	if err != nil {
		return nil, errors.Wrap(err, "cannot add secret")
	}

	return CoreAdminPassword(adminPasswordRes), nil
}

type CoreEncryptionKey graph.Resource

func (r *Reconciler) AddCoreEncryptionKey(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreEncryptionKey, error) {
	encryptionKey, err := r.GetCoreEncryptionKey(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get encryption key")
	}

	encryptionKeyRes, err := r.AddSecretToManage(ctx, encryptionKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot add encryption key")
	}

	return CoreEncryptionKey(encryptionKeyRes), nil
}

func (r *Reconciler) AddCoreConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor) (CoreCSRF, CoreTokenCertificate, CoreSecret, CoreAdminPassword, CoreEncryptionKey, error) {
	csrf, err := r.AddCoreCSRF(ctx, harbor)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "csrf")
	}

	secret, err := r.AddCoreSecret(ctx, harbor)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "secret")
	}

	tokenCertificate, err := r.AddCoreTokenCertificate(ctx, harbor)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "secret")
	}

	adminPassword, err := r.AddCoreAdminPassword(ctx, harbor)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "admin password")
	}

	encryptionKey, err := r.AddCoreEncryptionKey(ctx, harbor)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "encryption key")
	}

	return csrf, tokenCertificate, secret, adminPassword, encryptionKey, nil
}

type Core graph.Resource

func (r *Reconciler) AddCore(ctx context.Context, harbor *goharborv1alpha2.Harbor, registryAuth RegistryAuthSecret, chartMuseumAuth ChartMuseumAuthSecret, csrf CoreCSRF, tokenCertificate CoreTokenCertificate, secret CoreSecret, adminPassword CoreAdminPassword, encryptionKey CoreEncryptionKey) (Core, error) {
	core, err := r.GetCore(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get core")
	}

	coreRes, err := r.AddBasicResource(ctx, core, registryAuth, chartMuseumAuth, csrf, tokenCertificate, secret, adminPassword, encryptionKey)

	return Core(coreRes), errors.Wrap(err, "cannot add basic resource")
}

const (
	CoreAdminPasswordLength      = 32
	CoreAdminPasswordNumDigits   = 5
	CoreAdminPasswordNumSpecials = 10
)

func (r *Reconciler) GetCoreAdminPassword(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), "core", "admin-password")
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
		Type:      goharborv1alpha2.SecretTypeSingle,
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: password,
		},
	}, nil
}

const (
	CoreSecretPasswordLength      = 128
	CoreSecretPasswordNumDigits   = 16
	CoreSecretPasswordNumSpecials = 48
)

func (r *Reconciler) GetCoreSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), "core", "secret")
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
		Type:      goharborv1alpha2.SecretTypeSingle,
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: secret,
		},
	}, nil
}

const (
	CoreTokenServiceDefaultKeySize             = 4096
	CoreTokenServiceDefaultCertificateDuration = 3 * 30 * 24 * time.Hour
)

func (r *Reconciler) GetCoreTokenCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*certv1.Certificate, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), "core", "tokencert")
	namespace := harbor.GetNamespace()

	publicDNS, err := url.Parse(harbor.Spec.ExternalURL)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse external url")
	}

	secretName := r.NormalizeName(ctx, harbor.GetName(), "core", "tokencert")

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
	name := r.NormalizeName(ctx, harbor.GetName(), "core", "encryptionkey")
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
		Type:      goharborv1alpha2.SecretTypeSingle,
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: key,
		},
	}, nil
}

const (
	CSRFSecretPasswordLength      = 32
	CSRFSecretPasswordNumDigits   = 5
	CSRFSecretPasswordNumSpecials = 0
)

func (r *Reconciler) GetCSRFSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), "core", "csrf")
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
		Type:      goharborv1alpha2.SecretTypeSingle,
		StringData: map[string]string{
			goharborv1alpha2.CSRFSecretKey: csrf,
		},
	}, nil
}

func (r *Reconciler) GetCore(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.Core, error) { // nolint:funlen
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	credentials := goharborv1alpha2.CoreComponentsRegistryCredentialsSpec{
		Username:    RegistryAuthenticationUsername,
		PasswordRef: r.NormalizeName(ctx, harbor.GetName(), "registry", "basicauth"),
	}

	registryCtlURL := fmt.Sprintf("http://%s", r.NormalizeName(ctx, harbor.GetName(), "registryctl"))
	registryURL := fmt.Sprintf("http://%s", r.NormalizeName(ctx, harbor.GetName(), "registry"))
	portalURL := fmt.Sprintf("http://%s", r.NormalizeName(ctx, harbor.GetName(), "portal"))
	adminPasswordRef := r.NormalizeName(ctx, harbor.GetName(), "core", "admin-password")
	coreSecretRef := r.NormalizeName(ctx, harbor.GetName(), "core", "secret")
	encryptionKeyRef := r.NormalizeName(ctx, harbor.GetName(), "core", "encryptionkey")
	csrfRef := r.NormalizeName(ctx, harbor.GetName(), "core", "csrf")
	jobserviceURL := fmt.Sprintf("http://%s", r.NormalizeName(ctx, harbor.GetName(), "jobservice"))
	jobserviceSecretRef := r.NormalizeName(ctx, harbor.GetName(), "jobservice", "secret")
	tokenServiceURL := fmt.Sprintf("http://%s", fmt.Sprintf("%s/service/token", r.NormalizeName(ctx, harbor.GetName(), "core")))
	tokenCertificateRef := r.NormalizeName(ctx, harbor.GetName(), "core", "tokencert")

	dbDSN := harbor.Spec.DatabaseDSN(goharborv1alpha2.CoreDatabase)

	dbURL, err := dbDSN.GetDSN("")
	if err != nil {
		return nil, errors.Wrap(err, "cannot get database DSN")
	}

	port, err := strconv.ParseInt(dbURL.Port(), 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "unsupported port value")
	}

	postgresqlSpec := goharborv1alpha2.CorePostgresqlSpec{
		ExternalDatabaseSpec: goharborv1alpha2.ExternalDatabaseSpec{
			Host:        dbURL.Hostname(),
			PasswordRef: dbDSN.PasswordRef,
			Port:        int32(port),
			SSLMode:     dbURL.Query().Get("sslmode"),
			Username:    dbURL.User.Username(),
		},
		Name: strings.TrimLeft(dbURL.Path, "/"),
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
					URL:                 registryURL,
					ControllerURL:       registryCtlURL,
					Credentials:         credentials,
					Redis:               harbor.Spec.RedisDSN(goharborv1alpha2.RegistryRedis),
					StorageProviderName: harbor.Spec.Persistence.ImageChartStorage.Name(),
				},
				JobService: goharborv1alpha2.CoreComponentsJobServiceSpec{
					URL:       jobserviceURL,
					SecretRef: jobserviceSecretRef,
				},
				Portal: goharborv1alpha2.CoreComponentPortalSpec{
					URL: portalURL,
				},
				TokenService: goharborv1alpha2.CoreComponentsTokenServiceSpec{
					URL: tokenServiceURL,
				},
			},
			CoreConfig: goharborv1alpha2.CoreConfig{
				AdminInitialPasswordRef: adminPasswordRef,
				SecretRef:               coreSecretRef,
			},
			CSRFKeyRef: csrfRef,
			Database: goharborv1alpha2.CoreDatabaseSpec{
				EncryptionKeyRef:   encryptionKeyRef,
				CorePostgresqlSpec: postgresqlSpec,
			},
			ExternalEndpoint: harbor.Spec.ExternalURL,
			Redis: goharborv1alpha2.CoreRedisSpec{
				OpacifiedDSN: harbor.Spec.RedisDSN(goharborv1alpha2.CoreRedis),
			},
			Proxy: harbor.Spec.Proxy,
			ServiceToken: goharborv1alpha2.CoreServiceTokenSpec{
				CertificateRef: tokenCertificateRef,
			},
		},
	}, nil
}
