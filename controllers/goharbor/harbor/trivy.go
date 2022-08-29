package harbor

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) AddTrivyConfigurations(ctx context.Context, harbor *goharborv1.Harbor, tlsIssuer InternalTLSIssuer) (TrivyInternalCertificate, TrivyUpdateSecret, error) {
	if harbor.Spec.Trivy == nil {
		return nil, nil, nil
	}

	certificate, err := r.AddTrivyInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "certificate")
	}

	var secret TrivyUpdateSecret

	if harbor.Spec.Trivy.GithubTokenRef == "" {
		secret, err = r.AddTrivyUpdateSecret(ctx, harbor)
		if err != nil {
			return nil, nil, errors.Wrap(err, "update secret")
		}
	}

	return certificate, secret, nil
}

type TrivyInternalCertificate graph.Resource

func (r *Reconciler) AddTrivyInternalCertificate(ctx context.Context, harbor *goharborv1.Harbor, tlsIssuer InternalTLSIssuer) (TrivyInternalCertificate, error) {
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

func (r *Reconciler) AddTrivyUpdateSecret(ctx context.Context, harbor *goharborv1.Harbor) (TrivyUpdateSecret, error) {
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

const TrivyGithubCredentialsConfigKey = "trivy-github-credentials" //nolint:gosec

func (r *Reconciler) GetTrivyUpdateSecret(ctx context.Context, harbor *goharborv1.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.Trivy.String(), "github")
	namespace := harbor.GetNamespace()

	token, err := r.GetGithubToken(TrivyGithubCredentialsConfigKey)
	if err != nil {
		if config.IsNotFound(err, TrivyGithubCredentialsConfigKey) {
			return nil, nil
		}

		if config.IsNotFound(err, GithubCredentialsConfigKey) {
			return nil, nil
		}

		return nil, serrors.UnrecoverrableError(err, serrors.OperatorReason, "cannot get github credentials")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varFalse,
		Type:      harbormetav1.SecretTypeGithubToken,
		StringData: map[string]string{
			harbormetav1.GithubTokenKey: token,
		},
	}, nil
}

type Trivy graph.Resource

func (r *Reconciler) AddTrivy(ctx context.Context, harbor *goharborv1.Harbor, certificate TrivyInternalCertificate, secretUpdate TrivyUpdateSecret) (Trivy, error) {
	if harbor.Spec.Trivy == nil {
		return nil, nil
	}

	trivy, err := r.GetTrivy(ctx, harbor, secretUpdate != nil)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	trivyRes, err := r.AddBasicResource(ctx, trivy, certificate, secretUpdate)

	return Trivy(trivyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetTrivy(ctx context.Context, harbor *goharborv1.Harbor, hasUpdateSecret bool) (*goharborv1.Trivy, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	redis := harbor.Spec.RedisConnection(harbormetav1.TrivyRedis)

	githubTokenRef := harbor.Spec.Trivy.GithubTokenRef
	if githubTokenRef == "" && hasUpdateSecret {
		githubTokenRef = r.NormalizeName(ctx, harbor.GetName(), controllers.Trivy.String(), "github")
	}

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.TrivyTLS))

	var tokenServiceCertificateAuthorityRefs []string
	if harbor.Spec.Expose.Core.TLS != nil {
		tokenServiceCertificateAuthorityRefs = append(tokenServiceCertificateAuthorityRefs, harbor.Spec.Expose.Core.TLS.CertificateRef)
	}

	return &goharborv1.Trivy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: version.SetVersion(map[string]string{
				harbormetav1.NetworkPoliciesAnnotationName: harbormetav1.NetworkPoliciesAnnotationDisabled,
			}, harbor.Spec.Version),
		},
		Spec: goharborv1.TrivySpec{
			ComponentSpec: harbor.GetComponentSpec(ctx, harbormetav1.TrivyComponent),
			Redis: goharborv1.TrivyRedisSpec{
				RedisConnection: redis,
			},
			Storage: goharborv1.TrivyStorageSpec{
				Reports: r.TrivyReportsStorage(ctx, harbor),
				Cache:   r.TrivyCacheStorage(ctx, harbor),
			},
			Server: goharborv1.TrivyServerSpec{
				TLS:                                  tls,
				TokenServiceCertificateAuthorityRefs: tokenServiceCertificateAuthorityRefs,
			},
			Update: goharborv1.TrivyUpdateSpec{
				Skip:           harbor.Spec.Trivy.SkipUpdate,
				GithubTokenRef: githubTokenRef,
			},
			CertificateInjection: harbor.Spec.Trivy.CertificateInjection,
			Proxy:                harbor.GetComponentProxySpec(harbormetav1.TrivyComponent),
			Network:              harbor.Spec.Network,
			OfflineScan:          harbor.Spec.Trivy.OfflineScan,
		},
	}, nil
}
