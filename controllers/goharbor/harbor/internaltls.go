package harbor

import (
	"context"
	"time"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/graph"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	v1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) AddInternalTLSConfiguration(ctx context.Context, harbor *goharborv1.Harbor) (InternalTLSCertificateAuthorityIssuer, InternalTLSCertificateAuthority, InternalTLSIssuer, error) {
	caIssuer, err := r.AddInternalTLSCAIssuer(ctx, harbor)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "CA issuer")
	}

	ca, err := r.AddInternalTLSCACertificate(ctx, harbor, caIssuer)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "CA")
	}

	issuer, err := r.AddInternalTLSIssuer(ctx, harbor, ca)

	return caIssuer, ca, issuer, errors.Wrap(err, "TLS issuer")
}

type InternalTLSCertificateAuthorityIssuer graph.Resource

func (r *Reconciler) AddInternalTLSCAIssuer(ctx context.Context, harbor *goharborv1.Harbor) (InternalTLSCertificateAuthorityIssuer, error) {
	caIssuer, err := r.GetInternalTLSCertificateAuthorityIssuer(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	caIssuerRes, err := r.Controller.AddIssuerToManage(ctx, caIssuer)

	return InternalTLSCertificateAuthorityIssuer(caIssuerRes), errors.Wrap(err, "add")
}

type InternalTLSCertificateAuthority graph.Resource

func (r *Reconciler) AddInternalTLSCACertificate(ctx context.Context, harbor *goharborv1.Harbor, caIssuer InternalTLSCertificateAuthorityIssuer) (InternalTLSCertificateAuthority, error) {
	ca, err := r.GetInternalTLSCertificateAuthority(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	caRes, err := r.Controller.AddCertificateToManage(ctx, ca, caIssuer)

	return InternalTLSCertificateAuthority(caRes), errors.Wrap(err, "add")
}

type InternalTLSIssuer graph.Resource

func (r *Reconciler) AddInternalTLSIssuer(ctx context.Context, harbor *goharborv1.Harbor, ca InternalTLSCertificate) (InternalTLSIssuer, error) {
	issuer, err := r.GetInternalTLSIssuer(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	issuerRes, err := r.Controller.AddIssuerToManage(ctx, issuer, ca)

	return InternalTLSIssuer(issuerRes), errors.Wrap(err, "add")
}

const (
	InternalTLSCertificateAuthorityDurationConfigKey     = "internal-tls-certificate-authority-duration"
	InternalTLSCertificateAuthorityDurationDefaultConfig = 365 * 24 * time.Hour
)

func (r *Reconciler) GetInternalTLSCertificateAuthority(ctx context.Context, harbor *goharborv1.Harbor) (*certv1.Certificate, error) {
	if !harbor.Spec.InternalTLS.IsEnabled() {
		return nil, nil
	}

	duration := InternalTLSCertificateAuthorityDurationDefaultConfig

	durationValue, err := configstore.GetItemValue(InternalTLSCertificateAuthorityDurationConfigKey)
	if err != nil {
		if !config.IsNotFound(err, InternalTLSCertificateAuthorityDurationConfigKey) {
			return nil, err
		}
	} else {
		duration, err = time.ParseDuration(durationValue)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid config %s", InternalTLSCertificateAuthorityDurationConfigKey)
		}
	}

	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), "internal", "authority"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: certv1.CertificateSpec{
			SecretName: r.NormalizeName(ctx, harbor.GetName(), "internal-tls", "authority"),
			IssuerRef: v1.ObjectReference{
				Name: r.NormalizeName(ctx, harbor.GetName(), "internal", "authority"),
			},
			Duration: &metav1.Duration{
				Duration: duration,
			},
			CommonName: r.NormalizeName(ctx, harbor.GetName()),
			IsCA:       true,
		},
	}, nil
}

func (r *Reconciler) GetInternalTLSCertificateAuthorityIssuer(ctx context.Context, harbor *goharborv1.Harbor) (*certv1.Issuer, error) {
	if !harbor.Spec.InternalTLS.IsEnabled() {
		return nil, nil
	}

	return &certv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), "internal", "authority"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				SelfSigned: &certv1.SelfSignedIssuer{},
			},
		},
	}, nil
}

func (r *Reconciler) GetInternalTLSIssuer(ctx context.Context, harbor *goharborv1.Harbor) (*certv1.Issuer, error) {
	if !harbor.Spec.InternalTLS.IsEnabled() {
		return nil, nil
	}

	return &certv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), "internal"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				CA: &certv1.CAIssuer{
					SecretName: r.NormalizeName(ctx, harbor.GetName(), "internal-tls", "authority"),
				},
			},
		},
	}, nil
}

type InternalTLSCertificate graph.Resource

func (r *Reconciler) GetInternalTLSCertificateName(ctx context.Context, harbor *goharborv1.Harbor, component harbormetav1.ComponentWithTLS) string {
	return r.NormalizeName(ctx, harbor.GetName(), "internal", component.GetName())
}

func (r *Reconciler) GetInternalTLSCertificateSecretName(ctx context.Context, harbor *goharborv1.Harbor, component harbormetav1.ComponentWithTLS) string {
	return r.NormalizeName(ctx, harbor.GetName(), "internal-tls", component.GetName())
}

const (
	InternalTLSDurationConfigKey     = "internal-tls-duration"
	InternalTLSDurationDefaultConfig = 90 * 24 * time.Hour
)

func (r *Reconciler) GetInternalTLSCertificate(ctx context.Context, harbor *goharborv1.Harbor, component harbormetav1.ComponentWithTLS) (*certv1.Certificate, error) {
	if !harbor.Spec.InternalTLS.IsEnabled() {
		return nil, nil
	}

	duration := InternalTLSDurationDefaultConfig

	durationValue, err := configstore.GetItemValue(InternalTLSDurationConfigKey)
	if err != nil {
		if !config.IsNotFound(err, InternalTLSDurationConfigKey) {
			return nil, err
		}
	} else {
		duration, err = time.ParseDuration(durationValue)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid config %s", InternalTLSDurationConfigKey)
		}
	}

	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.GetInternalTLSCertificateName(ctx, harbor, component),
			Namespace: harbor.GetNamespace(),
		},
		Spec: certv1.CertificateSpec{
			SecretName: r.GetInternalTLSCertificateSecretName(ctx, harbor, component),
			IssuerRef: v1.ObjectReference{
				Name: r.NormalizeName(ctx, harbor.GetName(), "internal"),
			},
			Duration: &metav1.Duration{
				Duration: duration,
			},
			DNSNames: []string{r.NormalizeName(ctx, harbor.GetName(), component.GetName())},
		},
	}, nil
}
