package harbor

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (r *Reconciler) AddTrivyConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (TrivyInternalCertificate, TrivyUpdateSecret, error) {
	if harbor.Spec.Trivy == nil {
		return nil, nil, nil
	}

	certificate, err := r.AddTrivyInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "certificate")
	}

	secret, err := r.AddTrivyUpdateSecret(ctx, harbor)
	if err != nil {
		return nil, nil, errors.Wrap(err, "update secret")
	}

	return certificate, secret, nil
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

type TrivyUpdateSecret graph.Resource

func (r *Reconciler) AddTrivyUpdateSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (TrivyUpdateSecret, error) {
	authSecret, err := r.GetTrivyUpdateSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	authSecretRes, err := r.AddSecretToManage(ctx, authSecret)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return TrivyUpdateSecret(authSecretRes), nil
}

const TrivyGithubCredentialsConfigKey = "trivy-github-credentials"

func (r *Reconciler) GetTrivyUpdateSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.Trivy.String(), "github")
	namespace := harbor.GetNamespace()

	github, err := r.GetGithubCredentials(TrivyGithubCredentialsConfigKey)
	if err != nil {
		return nil, serrors.UnrecoverrableError(err, serrors.OperatorReason, "cannot get default github credentials")
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

type Trivy graph.Resource

func (r *Reconciler) AddTrivy(ctx context.Context, harbor *goharborv1alpha2.Harbor, certificate TrivyInternalCertificate, seretUpdate TrivyUpdateSecret) (Trivy, error) {
	if harbor.Spec.Trivy == nil {
		return nil, nil
	}

	trivy, err := r.GetTrivy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	trivyRes, err := r.AddBasicResource(ctx, trivy, certificate, seretUpdate)

	return Trivy(trivyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetTrivy(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.Trivy, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	redis := harbor.Spec.RedisConnection(harbormetav1.TrivyRedis)

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.TrivyTLS))

	return &goharborv1alpha2.Trivy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.TrivySpec{
			ComponentSpec: harbor.Spec.Trivy.ComponentSpec,
			Redis: goharborv1alpha2.TrivyRedisSpec{
				RedisConnection: redis,
			},
			Storage: goharborv1alpha2.TrivyStorageSpec{
				Reports: r.TrivyReportsStorage(ctx, harbor),
				Cache:   r.TrivyCacheStorage(ctx, harbor),
			},
			Server: goharborv1alpha2.TrivyServerSpec{
				TLS:                                  tls,
				TokenServiceCertificateAuthorityRefs: []string{harbor.Spec.Expose.Core.TLS.CertificateRef},
			},
			Update: goharborv1alpha2.TrivyUpdateSpec{
				Skip:           harbor.Spec.Trivy.SkipUpdate,
				GithubTokenRef: r.NormalizeName(ctx, harbor.GetName(), controllers.Trivy.String(), "github"),
			},
		},
	}, nil
}
