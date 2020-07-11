package harbor

import (
	"context"
	"net/url"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
)

func (r *Reconciler) AddNotaryServerConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor) (graph.Resource, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil
	}

	// TODO

	return nil, nil
}

type NotaryServer graph.Resource

func (r *Reconciler) AddNotaryServer(ctx context.Context, harbor *goharborv1alpha2.Harbor) (NotaryServer, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil
	}

	notaryServer, err := r.GetNotaryServer(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get notaryserver")
	}

	notaryServerRes, err := r.AddBasicResource(ctx, notaryServer)

	return NotaryServer(notaryServerRes), errors.Wrap(err, "cannot add basic resource")
}

const (
	TokenServiceIssuer                = "harbor-token-issuer"
	NotaryServerAuthenticationService = "harbor-notary"
)

func (r *Reconciler) GetNotaryServer(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.NotaryServer, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	tokenServiceCertificateRef := r.NormalizeName(ctx, harbor.GetName(), "core", "tokencert")
	trustServiceHost := r.NormalizeName(ctx, harbor.GetName(), "notarysigner")
	httpsCertificateRef := r.NormalizeName(ctx, harbor.GetName(), "notarysigner", "tlscert")

	serviceTokenURL, err := url.Parse(harbor.Spec.ExternalURL)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parseexternalURL")
	}

	serviceTokenURL.Path += "/service/token"

	dbDSN, err := harbor.Spec.DatabaseDSN(goharborv1alpha2.NotaryServerDatabase)
	if err != nil {
		return nil, errors.Wrap(err, "database")
	}

	return &goharborv1alpha2.NotaryServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.NotaryServerSpec{
			ComponentSpec: harbor.Spec.Notary.ComponentSpec,
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
				OpacifiedDSN: *dbDSN,
				Type:         "postgres",
			},
			TrustService: goharborv1alpha2.NotaryServerTrustServiceSpec{
				Host:           trustServiceHost,
				KeyAlgorithm:   string(certv1.ECDSAKeyAlgorithm),
				Type:           "remote",
				CertificateRef: httpsCertificateRef,
			},
		},
	}, nil
}
