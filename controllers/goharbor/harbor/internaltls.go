package harbor

import (
	"context"
	"time"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	v1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

type (
	InternalTLSCertificateAuthorityIssuer graph.Resource
	InternalTLSCertificateAuthority       graph.Resource
	InternalTLSIssuer                     graph.Resource
)

func (r *Reconciler) AddInternalTLSIssuer(ctx context.Context, harbor *goharborv1alpha2.Harbor) (InternalTLSCertificateAuthorityIssuer, InternalTLSCertificateAuthority, InternalTLSIssuer, error) {
	caIssuer, err := r.GetInternalTLSCertificateAuthorityIssuer(ctx, harbor)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "cannot get TLS CA issuer")
	}

	caIssuerRes, err := r.Controller.AddIssuerToManage(ctx, caIssuer)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "cannot add TLS CA issuer")
	}

	ca, err := r.GetInternalTLSCertificateAuthority(ctx, harbor)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "cannot get TLS CA")
	}

	caRes, err := r.Controller.AddCertificateToManage(ctx, ca, caIssuerRes)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "cannot add TLS CA")
	}

	issuer, err := r.GetInternalTLSIssuer(ctx, harbor)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "cannot get TLS issuer")
	}

	issuerRes, err := r.Controller.AddIssuerToManage(ctx, issuer, caRes)

	return InternalTLSCertificateAuthorityIssuer(caIssuerRes), InternalTLSCertificateAuthority(caRes), InternalTLSIssuer(issuerRes), errors.Wrap(err, "cannot add TLS issuer")
}

const (
	InternalTLSCertificateAuthorityDurationConfigKey     = "internal-tls-certificate-authority-duration"
	InternalTLSCertificateAuthorityDurationDefaultConfig = 365 * 24 * time.Hour
)

func (r *Reconciler) GetInternalTLSCertificateAuthority(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*certv1.Certificate, error) {
	if !harbor.Spec.InternalTLS.IsEnabled() {
		return nil, nil
	}

	duration := InternalTLSCertificateAuthorityDurationDefaultConfig

	durationValue, err := configstore.GetItemValue(InternalTLSCertificateAuthorityDurationConfigKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
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

func (r *Reconciler) GetInternalTLSCertificateAuthorityIssuer(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*certv1.Issuer, error) {
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

func (r *Reconciler) GetInternalTLSIssuer(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*certv1.Issuer, error) {
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

func (r *Reconciler) GetInternalTLSCertificateName(ctx context.Context, harbor *goharborv1alpha2.Harbor, component goharborv1alpha2.ComponentWithTLS) string {
	return r.NormalizeName(ctx, harbor.GetName(), "internal", component.GetName())
}

func (r *Reconciler) GetInternalTLSCertificateSecretName(ctx context.Context, harbor *goharborv1alpha2.Harbor, component goharborv1alpha2.ComponentWithTLS) string {
	return r.NormalizeName(ctx, harbor.GetName(), "internal-tls", component.GetName())
}

const (
	InternalTLSDurationConfigKey     = "internal-tls-duration"
	InternalTLSDurationDefaultConfig = 90 * 24 * time.Hour
)

func (r *Reconciler) GetInternalTLSCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, component goharborv1alpha2.ComponentWithTLS) (*certv1.Certificate, error) {
	if !harbor.Spec.InternalTLS.IsEnabled() {
		return nil, nil
	}

	duration := InternalTLSDurationDefaultConfig

	durationValue, err := configstore.GetItemValue(InternalTLSDurationConfigKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
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
