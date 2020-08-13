package harbor

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (r *Reconciler) AddTrivyConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (TrivyInternalCertificate, error) {
	if harbor.Spec.Trivy == nil {
		return nil, nil
	}

	certificate, err := r.AddTrivyInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "certificate")
	}

	return certificate, nil
}

type TrivyInternalCertificate graph.Resource

func (r *Reconciler) AddTrivyInternalCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (TrivyInternalCertificate, error) {
	cert, err := r.GetInternalTLSCertificate(ctx, harbor, harbormetav1.TrivyTLS)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return TrivyInternalCertificate(certRes), nil
}

type Trivy graph.Resource

func (r *Reconciler) AddTrivy(ctx context.Context, harbor *goharborv1alpha2.Harbor, certificate TrivyInternalCertificate) (Trivy, error) {
	if harbor.Spec.Trivy == nil {
		return nil, nil
	}

	trivy, err := r.GetTrivy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	trivyRes, err := r.AddBasicResource(ctx, trivy, certificate)

	return Trivy(trivyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetTrivy(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.Trivy, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	redis := harbor.Spec.RedisConnection(harbormetav1.TrivyRedis)

	skipUpdate := harbor.Spec.Trivy.SkipUpdate
	githubToken := harbor.Spec.Trivy.GithubToken

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.TrivyTLS))

	return &goharborv1alpha2.Trivy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.TrivySpec{
			ComponentSpec: harbor.Spec.Trivy.ComponentSpec,
			Cache: goharborv1alpha2.TrivyCacheSpec{
				Redis: redis,
			},
			Storage: goharborv1alpha2.TrivyStorageSpec{
				Reports: r.TrivyReportsStorage(ctx, harbor),
				Cache:   r.TrivyCacheStorage(ctx, harbor),
			},
			Server: goharborv1alpha2.TrivyServerSpec{
				TLS:         tls,
				SkipUpdate:  skipUpdate,
				GithubToken: githubToken,
			},
		},
	}, nil
}
