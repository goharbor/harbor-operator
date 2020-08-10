package harbor

import (
	"context"
	"net/url"
	"time"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	v1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (r *Reconciler) AddNotaryServerConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer, notaryIssuer NotarySignerCertificateIssuer) (NotaryServerCertificate, NotaryServerInternalCertificate, NotaryServerMigrationSecret, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil, nil, nil
	}

	clientCert, err := r.AddNotaryServerClientCertificate(ctx, harbor, notaryIssuer)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "client certificate")
	}

	migrationSecret, err := r.AddNotaryServerMigrationSecret(ctx, harbor)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "migration")
	}

	certificate, err := r.AddNotaryServerInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "certificate")
	}

	return clientCert, certificate, migrationSecret, nil
}

type NotaryServerCertificate graph.Resource

func (r *Reconciler) AddNotaryServerClientCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, issuer NotarySignerCertificateIssuer) (NotaryServerCertificate, error) {
	cert, err := r.GetNotaryServerCertificate(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, issuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return NotaryServerCertificate(certRes), nil
}

const (
	NotaryServerCertificateDurationConfigKey     = "notaryserver-certificate-duration"
	NotaryServerCertificateDurationDefaultConfig = 90 * 24 * time.Hour
)

func (r *Reconciler) getNotaryServerCertificateDuration() (time.Duration, error) {
	durationValue, err := r.ConfigStore.GetItemValue(NotarySignerCertificateDurationConfigKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			return NotarySignerCertificateDurationDefaultConfig, nil
		}

		return NotarySignerCertificateDurationDefaultConfig, err
	}

	return time.ParseDuration(durationValue)
}

const (
	NotaryServerCertificateAlgorithmConfigKey     = "notaryserver-certificate-algorithm"
	NotaryServerCertificateAlgorithmDefaultConfig = certv1.ECDSAKeyAlgorithm
)

func (r *Reconciler) getNotaryServerCertificateAlgorithm() (certv1.KeyAlgorithm, error) {
	algorithm, err := r.ConfigStore.GetItemValue(NotaryServerCertificateAlgorithmConfigKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			return NotaryServerCertificateAlgorithmDefaultConfig, nil
		}

		return NotaryServerCertificateAlgorithmDefaultConfig, err
	}

	return certv1.KeyAlgorithm(algorithm), nil
}

func (r *Reconciler) GetNotaryServerCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*certv1.Certificate, error) {
	duration, err := r.getNotaryServerCertificateDuration()
	if err != nil {
		return nil, errors.Wrap(err, "duration configuration")
	}

	algorithm, err := r.getNotaryServerCertificateAlgorithm()
	if err != nil {
		return nil, errors.Wrap(err, "algorithm configuration")
	}

	notarySignerIssuer := r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication")
	secretName := r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String(), "authentication")

	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String(), "authentication"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: certv1.CertificateSpec{
			SecretName: secretName,
			IssuerRef: v1.ObjectReference{
				Name: notarySignerIssuer,
			},
			KeyAlgorithm: algorithm,
			Duration:     &metav1.Duration{Duration: duration},
			CommonName:   r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String()),
			DNSNames:     []string{r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String())},
			Usages: []certv1.KeyUsage{
				certv1.UsageDigitalSignature,
				certv1.UsageKeyEncipherment,
				certv1.UsageClientAuth,
			},
			IsCA: false,
		},
	}, nil
}

type NotaryServerInternalCertificate graph.Resource

func (r *Reconciler) AddNotaryServerInternalCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (NotaryServerInternalCertificate, error) {
	cert, err := r.GetInternalTLSCertificate(ctx, harbor, harbormetav1.NotaryServerTLS)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return NotaryServerInternalCertificate(certRes), nil
}

func (r *Reconciler) GetDefaultNotaryServerMigrationSource(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.NotaryMigrationGithubSpec, error) {
	source, err := r.GetDefaultNotaryMigrationSource()
	if err != nil {
		return nil, err
	}

	return &goharborv1alpha2.NotaryMigrationGithubSpec{
		CredentialsRef: r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String(), "migration"),
		Owner:          source.Owner,
		Path:           source.Path,
		Reference:      source.Reference,
		RepositoryName: source.Repository,
	}, nil
}

type NotaryServerMigrationSecret graph.Resource

func (r *Reconciler) AddNotaryServerMigrationSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (NotaryServerMigrationSecret, error) {
	authSecret, err := r.GetNotaryServerMigrationSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	authSecretRes, err := r.AddSecretToManage(ctx, authSecret)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return NotaryServerMigrationSecret(authSecretRes), nil
}

func (r *Reconciler) GetNotaryServerMigrationSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String(), "migration")
	namespace := harbor.GetNamespace()

	github, err := r.GetDefaultNotaryMigrationCredentials()
	if err != nil {
		return nil, serrors.UnrecoverrableError(err, serrors.OperatorReason, "cannot get default migration source")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varFalse,
		Type:      harbormetav1.SecretTypeGithubToken,
		StringData: map[string]string{
			harbormetav1.GithubTokenUserKey:     github.User,
			harbormetav1.GithubTokenPasswordKey: github.Token,
		},
	}, nil
}

type NotaryServer graph.Resource

func (r *Reconciler) AddNotaryServer(ctx context.Context, harbor *goharborv1alpha2.Harbor, certificate NotaryServerInternalCertificate, authCert NotaryServerCertificate, migrationSecret NotaryServerMigrationSecret) (NotaryServer, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil
	}

	notaryServer, err := r.GetNotaryServer(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	notaryServerRes, err := r.AddBasicResource(ctx, notaryServer, certificate, authCert, migrationSecret)

	return NotaryServer(notaryServerRes), errors.Wrap(err, "add")
}

const (
	TokenServiceIssuer                = "harbor-token-issuer"
	NotaryServerAuthenticationService = "harbor-notary"
)

func (r *Reconciler) GetNotaryServer(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.NotaryServer, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	tokenServiceCertificateRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "tokencert")
	trustServiceHost := r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String())
	authCertificateRef := r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String(), "authentication")

	serviceTokenURL, err := url.Parse(harbor.Spec.ExternalURL)
	if err != nil {
		return nil, serrors.UnrecoverrableError(errors.Wrap(err, "cannot parse externalURL"), serrors.InvalidSpecReason, "unable to configure service token")
	}

	serviceTokenURL.Path += "/service/token"

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.NotaryServerTLS))

	storage, err := harbor.Spec.Database.GetPostgresqlConnection(harbormetav1.NotaryServerComponent)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get storage configuration")
	}

	algorithm, err := r.getNotarySignerCertificateAlgorithm()
	if err != nil {
		return nil, errors.Wrap(err, "algorithm configuration")
	}

	var migration *goharborv1alpha2.NotaryMigrationSpec

	if harbor.Spec.Notary.IsMigrationEnabled() {
		migration = &goharborv1alpha2.NotaryMigrationSpec{}

		github, err := r.GetDefaultNotaryServerMigrationSource(ctx, harbor)
		if err != nil {
			return nil, serrors.UnrecoverrableError(err, serrors.OperatorReason, "cannot get notary migration source")
		}

		migration.Github = github
	}

	return &goharborv1alpha2.NotaryServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.NotaryServerSpec{
			ComponentSpec: harbor.Spec.Notary.Server,
			TLS:           tls,
			Authentication: &goharborv1alpha2.NotaryServerAuthSpec{
				Token: goharborv1alpha2.NotaryServerAuthTokenSpec{
					CertificateRef: tokenServiceCertificateRef,
					Issuer:         TokenServiceIssuer,
					Realm:          serviceTokenURL.String(),
					Service:        NotaryServerAuthenticationService,
				},
			},
			Logging: goharborv1alpha2.NotaryLoggingSpec{
				Level: harbor.Spec.LogLevel.Notary(),
			},
			Storage: goharborv1alpha2.NotaryStorageSpec{
				Postgres: *storage,
			},
			TrustService: goharborv1alpha2.NotaryServerTrustServiceSpec{
				Remote: &goharborv1alpha2.NotaryServerTrustServiceRemoteSpec{
					Host:           trustServiceHost,
					CertificateRef: authCertificateRef,
					KeyAlgorithm:   algorithm,
					Port:           goharborv1alpha2.NotarySignerAPIPort,
				},
			},
			Migration: migration,
		},
	}, nil
}
