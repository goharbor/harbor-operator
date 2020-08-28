package harbor

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RegistryController graph.Resource

func (r *Reconciler) AddRegistryController(ctx context.Context, harbor *goharborv1alpha2.Harbor, registry Registry, tlsIssuer InternalTLSIssuer) (RegistryControllerInternalCertificate, RegistryController, error) {
	certificate, err := r.AddRegistryControllerInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "certificate")
	}

	registryCtl, err := r.GetRegistryCtl(ctx, harbor)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot get registryCtl")
	}

	registryCtlRes, err := r.AddBasicResource(ctx, registryCtl, registry, certificate)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot add registryCtl")
	}

	return certificate, RegistryController(registryCtlRes), nil
}

type RegistryControllerInternalCertificate graph.Resource

func (r *Reconciler) AddRegistryControllerInternalCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (RegistryControllerInternalCertificate, error) {
	cert, err := r.GetInternalTLSCertificate(ctx, harbor, harbormetav1.RegistryControllerTLS)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return RegistryControllerInternalCertificate(certRes), nil
}

func (r *Reconciler) GetRegistryCtl(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.RegistryController, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	registryName := r.NormalizeName(ctx, harbor.GetName())

	coreSecretRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "secret")
	jobserviceSecretRef := r.NormalizeName(ctx, harbor.GetName(), controllers.JobService.String(), "secret")

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.RegistryControllerTLS))

	return &goharborv1alpha2.RegistryController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.RegistryControllerSpec{
			ComponentSpec: harbor.Spec.Registry.ComponentSpec,
			RegistryRef:   registryName,
			Log: goharborv1alpha2.RegistryControllerLogSpec{
				Level: harbor.Spec.LogLevel.RegistryCtl(),
			},
			Authentication: goharborv1alpha2.RegistryControllerAuthenticationSpec{
				CoreSecretRef:       coreSecretRef,
				JobServiceSecretRef: jobserviceSecretRef,
			},
			TLS: tls,
		},
	}, nil
}
