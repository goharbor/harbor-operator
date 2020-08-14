package harbor

import (
	"context"
	"time"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	v1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (r *Reconciler) AddNotarySignerConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor) (NotarySignerCertificateIssuer, NotarySignerCertificate, NotarySignerEncryptionKey, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil, nil, nil
	}

	caIssuer, err := r.AddNotarySignerCertificateAuthorityIssuer(ctx, harbor)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "ca-issuer")
	}

	ca, err := r.AddNotarySignerCertificateAuthority(ctx, harbor, caIssuer)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "ca-issuer")
	}

	issuer, err := r.AddNotarySignerCertificateIssuer(ctx, harbor, ca)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "issuer")
	}

	certificate, err := r.AddNotarySignerCertificate(ctx, harbor, issuer)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "certificate")
	}

	encryptionKey, err := r.AddNotarySignerEncryptionKey(ctx, harbor)

	return issuer, certificate, encryptionKey, errors.Wrap(err, "encryption key")
}

const (
	NotarySignerCertificateAuthorityDurationConfigKey     = "notary-signer-certificate-authority-duration"
	NotarySignerCertificateAuthorityDurationDefaultConfig = 365 * 24 * time.Hour
)

func (r *Reconciler) GetNotarySignerCertificateAuthority(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*certv1.Certificate, error) {
	duration := NotarySignerCertificateAuthorityDurationDefaultConfig

	durationValue, err := configstore.GetItemValue(NotarySignerCertificateAuthorityDurationConfigKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, err
		}
	} else {
		duration, err = time.ParseDuration(durationValue)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid config %s", NotarySignerCertificateAuthorityDurationConfigKey)
		}
	}

	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication", "authority"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: certv1.CertificateSpec{
			SecretName: r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication", "authority"),
			IssuerRef: v1.ObjectReference{
				Name: r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication", "authority"),
			},
			Duration: &metav1.Duration{
				Duration: duration,
			},
			CommonName: r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String()),
			IsCA:       true,
			Usages: []certv1.KeyUsage{
				certv1.UsageDigitalSignature,
				certv1.UsageKeyEncipherment,
				// certv1.UsageKeyAgreement,
				certv1.UsageClientAuth,
				// certv1.UsageCertSign,
				// certv1.UsageCRLSign,
				// certv1.UsageOCSPSigning,
			},
		},
	}, nil
}

type NotarySignerEncryptionKey graph.Resource

func (r *Reconciler) AddNotarySignerEncryptionKey(ctx context.Context, harbor *goharborv1alpha2.Harbor) (NotarySignerEncryptionKey, error) {
	secret, err := r.GetNotarySignerEncryptionKey(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	secretRes, err := r.Controller.AddImmutableSecretToManage(ctx, secret)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return NotarySignerEncryptionKey(secretRes), nil
}

const (
	NotarySignerEncryptionKeyLength      = 128
	NotarySignerEncryptionKeyNumDigits   = 16
	NotarySignerEncryptionKeyNumSpecials = 48
)

func (r *Reconciler) GetNotarySignerEncryptionKey(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "encryption-key")
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
		Type: harbormetav1.SecretTypeNotarySignerAliases,
		StringData: map[string]string{
			harbormetav1.DefaultAliasSecretKey: "defaultalias",
			"defaultalias":                     secret,
		},
	}, nil
}

type NotarySignerCertificateAuthorityIssuer graph.Resource

func (r *Reconciler) AddNotarySignerCertificateAuthorityIssuer(ctx context.Context, harbor *goharborv1alpha2.Harbor) (NotarySignerCertificateAuthorityIssuer, error) {
	issuer, err := r.GetNotarySignerCertificateAuthorityIssuer(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	issuerRes, err := r.Controller.AddIssuerToManage(ctx, issuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return NotarySignerCertificateAuthorityIssuer(issuerRes), nil
}

func (r *Reconciler) GetNotarySignerCertificateAuthorityIssuer(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*certv1.Issuer, error) {
	return &certv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication", "authority"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				SelfSigned: &certv1.SelfSignedIssuer{},
			},
		},
	}, nil
}

type NotarySignerCertificateAuthority graph.Resource

func (r *Reconciler) AddNotarySignerCertificateAuthority(ctx context.Context, harbor *goharborv1alpha2.Harbor, issuer NotarySignerCertificateAuthorityIssuer) (NotarySignerCertificateAuthority, error) {
	cert, err := r.GetNotarySignerCertificateAuthority(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, issuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return NotarySignerCertificateAuthority(certRes), nil
}

type NotarySignerCertificateIssuer graph.Resource

func (r *Reconciler) AddNotarySignerCertificateIssuer(ctx context.Context, harbor *goharborv1alpha2.Harbor, ca NotarySignerCertificateAuthority) (NotarySignerCertificateIssuer, error) {
	issuer, err := r.GetNotarySignerCertificateIssuer(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	issuerRes, err := r.Controller.AddIssuerToManage(ctx, issuer, ca)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return NotarySignerCertificateIssuer(issuerRes), nil
}

func (r *Reconciler) GetNotarySignerCertificateIssuer(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*certv1.Issuer, error) {
	return &certv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				CA: &certv1.CAIssuer{
					SecretName: r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication", "authority"),
				},
			},
		},
	}, nil
}

type NotarySignerCertificate graph.Resource

func (r *Reconciler) AddNotarySignerCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, issuer NotarySignerCertificateIssuer) (NotarySignerCertificate, error) {
	cert, err := r.GetNotarySignerCertificate(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, issuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return NotarySignerCertificate(certRes), nil
}

const (
	NotarySignerCertificateDurationConfigKey     = "notarysigner-certificate-duration"
	NotarySignerCertificateDurationDefaultConfig = 90 * 24 * time.Hour
)

func (r *Reconciler) getNotarySignerCertificateDuration() (time.Duration, error) {
	durationValue, err := configstore.GetItemValue(NotarySignerCertificateDurationConfigKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			return NotarySignerCertificateDurationDefaultConfig, nil
		}

		return NotarySignerCertificateDurationDefaultConfig, err
	}

	return time.ParseDuration(durationValue)
}

const (
	NotarySignerCertificateAlgorithmConfigKey     = "notarysigner-certificate-algorithm"
	NotarySignerCertificateAlgorithmDefaultConfig = certv1.ECDSAKeyAlgorithm
)

func (r *Reconciler) getNotarySignerCertificateAlgorithm() (certv1.KeyAlgorithm, error) {
	algorithm, err := configstore.GetItemValue(NotarySignerCertificateAlgorithmConfigKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			return NotarySignerCertificateAlgorithmDefaultConfig, nil
		}

		return NotarySignerCertificateAlgorithmDefaultConfig, err
	}

	return certv1.KeyAlgorithm(algorithm), nil
}

func (r *Reconciler) GetNotarySignerCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*certv1.Certificate, error) {
	duration, err := r.getNotarySignerCertificateDuration()
	if err != nil {
		return nil, errors.Wrap(err, "duration configuration")
	}

	algorithm, err := r.getNotarySignerCertificateAlgorithm()
	if err != nil {
		return nil, errors.Wrap(err, "algorithm configuration")
	}

	secretName := r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication")

	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: certv1.CertificateSpec{
			SecretName: secretName,
			IssuerRef: v1.ObjectReference{
				Name: r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication"),
			},
			KeyAlgorithm: algorithm,
			Duration:     &metav1.Duration{Duration: duration},
			CommonName:   r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String()),
			DNSNames:     []string{r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String())},
			Usages: []certv1.KeyUsage{
				certv1.UsageDigitalSignature,
				certv1.UsageKeyEncipherment,
				// certv1.UsageKeyAgreement,
				certv1.UsageClientAuth,
				// certv1.UsageServerAuth,
			},
			IsCA: false,
		},
	}, nil
}

type NotarySigner graph.Resource

func (r *Reconciler) AddNotarySigner(ctx context.Context, harbor *goharborv1alpha2.Harbor, certificate NotarySignerCertificate, encryptionKey NotarySignerEncryptionKey) (NotarySigner, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil
	}

	notaryServer, err := r.GetNotarySigner(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	notaryServerRes, err := r.AddBasicResource(ctx, notaryServer, certificate, encryptionKey)

	return NotarySigner(notaryServerRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetNotarySigner(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.NotarySigner, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	encryptionKeyAliasesRef := r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "encryption-key")

	certificateRef := r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "authentication")

	storage, err := harbor.Spec.Database.GetPostgresqlConnection(harbormetav1.NotarySignerComponent)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get storage configuration")
	}

	return &goharborv1alpha2.NotarySigner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.NotarySignerSpec{
			ComponentSpec: harbor.Spec.Notary.ComponentSpec,
			Authentication: goharborv1alpha2.NotarySignerAuthenticationSpec{
				CertificateRef: certificateRef,
			},
			Logging: goharborv1alpha2.NotaryLoggingSpec{
				Level: harbor.Spec.LogLevel.Notary(),
			},
			Storage: goharborv1alpha2.NotarySignerStorageSpec{
				NotaryStorageSpec: goharborv1alpha2.NotaryStorageSpec{
					Postgres: *storage,
				},
				AliasesRef: encryptionKeyAliasesRef,
			},
		},
	}, nil
}
