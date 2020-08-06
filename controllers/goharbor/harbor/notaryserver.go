package harbor

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (r *Reconciler) AddNotaryServerConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (NotaryServerInternalCertificate, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil
	}

	certificate, err := r.AddNotaryServerInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "certificate")
	}

	return certificate, nil
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

func (r *Reconciler) AddNotaryServer(ctx context.Context, harbor *goharborv1alpha2.Harbor, certificate NotaryServerInternalCertificate) (NotaryServer, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil
	}

	notaryServer, err := r.GetNotaryServer(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	notaryServerRes, err := r.AddBasicResource(ctx, notaryServer, certificate)

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
	notarySignerCertificateRef := r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication")

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

	return &goharborv1alpha2.NotaryServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.NotaryServerSpec{
			ComponentSpec: harbor.Spec.Notary.Server,
			TLS:           tls,
			Auth: &goharborv1alpha2.NotaryServerAuthSpec{
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
				Host:           trustServiceHost,
				Type:           goharborv1alpha2.NotaryServerTrustRemoteType,
				CertificateRef: notarySignerCertificateRef,
				KeyAlgorithm:   algorithm,
				Port:           goharborv1alpha2.NotarySignerAPIPort,
			},
		},
	}, nil
}
