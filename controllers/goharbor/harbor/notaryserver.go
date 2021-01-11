package harbor

import (
	"context"
	"net/url"
	"time"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	v1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) AddNotaryServerConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer, notaryIssuer NotarySignerCertificateIssuer) (NotaryServerCertificate, NotaryServerInternalCertificate, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil, nil
	}

	clientCert, err := r.AddNotaryServerClientCertificate(ctx, harbor, notaryIssuer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "client certificate")
	}

	certificate, err := r.AddNotaryServerInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "certificate")
	}

	return clientCert, certificate, nil
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
	durationValue, err := r.ConfigStore.GetItemValue(NotaryServerCertificateDurationConfigKey)
	if err != nil {
		if config.IsNotFound(err, NotaryServerCertificateDurationConfigKey) {
			return NotarySignerCertificateDurationDefaultConfig, nil
		}

		return NotaryServerCertificateDurationDefaultConfig, err
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
		if config.IsNotFound(err, NotaryServerCertificateAlgorithmConfigKey) {
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

type NotaryServer graph.Resource

func (r *Reconciler) AddNotaryServer(ctx context.Context, harbor *goharborv1alpha2.Harbor, certificate NotaryServerInternalCertificate, authCert NotaryServerCertificate) (NotaryServer, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil
	}

	notaryServer, err := r.GetNotaryServer(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	notaryServerRes, err := r.AddBasicResource(ctx, notaryServer, certificate, authCert)

	return NotaryServer(notaryServerRes), errors.Wrap(err, "add")
}

const (
	TokenServiceIssuer                = "harbor-token-issuer"
	NotaryServerAuthenticationService = "harbor-notary"
)

func (r *Reconciler) GetNotaryServer(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.NotaryServer, error) { // nolint:funlen
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

	migrationEnabled := harbor.Spec.Notary.IsMigrationEnabled()

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
			MigrationEnabled: &migrationEnabled,
		},
	}, nil
}
