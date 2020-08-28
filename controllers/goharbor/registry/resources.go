package registry

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1alpha2.Registry{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	registry, ok := resource.(*goharborv1alpha2.Registry)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "addResources")
	defer span.Finish()

	service, err := r.GetService(ctx, registry)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	_, err = r.Controller.AddServiceToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot add service %s", service.GetName())
	}

	deploymentDependencies, err := r.GetSecrets(ctx, registry)
	if err != nil {
		return errors.Wrap(err, "cannot get secrets")
	}

	configMap, err := r.GetConfigMap(ctx, registry)
	if err != nil {
		return errors.Wrap(err, "cannot get configMap")
	}

	configMapResource, err := r.Controller.AddConfigMapToManage(ctx, configMap)
	if err != nil {
		return errors.Wrapf(err, "cannot add configMap %s", configMap.GetName())
	}

	deploymentDependencies = append(deploymentDependencies, configMapResource)

	if registry.Spec.HTTP.SecretRef != "" {
		httpSecret, err := r.AddExternalTypedSecret(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registry.Spec.HTTP.SecretRef,
				Namespace: registry.GetNamespace(),
			},
		}, harbormetav1.SecretTypeRegistry)
		if err != nil {
			return errors.Wrap(err, "cannot add http secret")
		}

		deploymentDependencies = append(deploymentDependencies, httpSecret)
	}

	deployment, err := r.GetDeployment(ctx, registry)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	_, err = r.Controller.AddDeploymentToManage(ctx, deployment, deploymentDependencies...)
	if err != nil {
		return errors.Wrapf(err, "cannot add deployment %s", deployment.GetName())
	}

	return nil
}

func (r *Reconciler) GetSecrets(ctx context.Context, registry *goharborv1alpha2.Registry) ([]graph.Resource, error) {
	secrets, err := r.GetStorageSecrets(ctx, registry)
	if err != nil {
		return nil, errors.Wrap(err, "storage")
	}

	proxy, err := r.GetProxySecrets(ctx, registry)
	if err != nil {
		return nil, errors.Wrap(err, "proxy")
	}

	secrets = append(secrets, proxy...)

	http, err := r.GetHTTPSecrets(ctx, registry)
	if err != nil {
		return nil, errors.Wrap(err, "http")
	}

	secrets = append(secrets, http...)

	auth, err := r.GetAuthenticationSecrets(ctx, registry)
	if err != nil {
		return nil, errors.Wrap(err, "authentication")
	}

	secrets = append(secrets, auth...)

	return secrets, nil
}

func (r *Reconciler) GetProxySecrets(ctx context.Context, registry *goharborv1alpha2.Registry) ([]graph.Resource, error) {
	if registry.Spec.Proxy == nil {
		return nil, nil
	}

	if registry.Spec.Proxy.BasicAuthRef != "" {
		secret, err := r.Controller.AddExternalTypedSecret(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registry.Spec.Proxy.BasicAuthRef,
				Namespace: registry.GetNamespace(),
			},
		}, harbormetav1.SecretTypeSingle)

		return []graph.Resource{secret}, errors.Wrapf(err, "cannot add external typed secret %s", registry.Spec.Proxy.BasicAuthRef)
	}

	return nil, nil
}

func (r *Reconciler) GetHTTPSecrets(ctx context.Context, registry *goharborv1alpha2.Registry) ([]graph.Resource, error) {
	if registry.Spec.HTTP.SecretRef != "" {
		secret, err := r.Controller.AddExternalTypedSecret(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registry.Spec.HTTP.SecretRef,
				Namespace: registry.GetNamespace(),
			},
		}, harbormetav1.SecretTypeSingle)

		return []graph.Resource{secret}, errors.Wrapf(err, "cannot add external typed secret %s", registry.Spec.HTTP.SecretRef)
	}

	return nil, nil
}

func (r *Reconciler) GetAuthenticationSecrets(ctx context.Context, registry *goharborv1alpha2.Registry) ([]graph.Resource, error) {
	if registry.Spec.Authentication.Token != nil {
		secret, err := r.Controller.AddExternalTypedSecret(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registry.Spec.Authentication.Token.CertificateRef,
				Namespace: registry.GetNamespace(),
			},
		}, harbormetav1.SecretTypeSingle)

		return []graph.Resource{secret}, errors.Wrapf(err, "cannot add external typed secret %s", registry.Spec.Authentication.Token.CertificateRef)
	}

	if registry.Spec.Authentication.HTPasswd != nil {
		secret, err := r.Controller.AddExternalTypedSecret(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registry.Spec.Authentication.HTPasswd.SecretRef,
				Namespace: registry.GetNamespace(),
			},
		}, harbormetav1.SecretTypeSingle)

		return []graph.Resource{secret}, errors.Wrapf(err, "cannot add external typed secret %s", registry.Spec.Authentication.HTPasswd.SecretRef)
	}

	return nil, nil
}

func (r *Reconciler) GetStorageSecrets(ctx context.Context, registry *goharborv1alpha2.Registry) ([]graph.Resource, error) {
	if registry.Spec.Storage.Driver.S3 != nil {
		return r.GetS3StorageSecrets(ctx, registry)
	}

	if registry.Spec.Storage.Driver.Swift != nil {
		return r.GetSwiftStorageSecrets(ctx, registry)
	}

	return nil, nil
}

func (r *Reconciler) GetS3StorageSecrets(ctx context.Context, registry *goharborv1alpha2.Registry) ([]graph.Resource, error) {
	if registry.Spec.Storage.Driver.S3.SecretKeyRef != "" {
		secret, err := r.Controller.AddExternalTypedSecret(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registry.Spec.Storage.Driver.S3.SecretKeyRef,
				Namespace: registry.GetNamespace(),
			},
		}, harbormetav1.SecretTypeSingle)

		return []graph.Resource{secret}, errors.Wrapf(err, "cannot add external typed secret %s", registry.Spec.Storage.Driver.S3.SecretKeyRef)
	}

	return nil, nil
}

func (r *Reconciler) GetSwiftStorageSecrets(ctx context.Context, registry *goharborv1alpha2.Registry) ([]graph.Resource, error) {
	secret, err := r.Controller.AddExternalTypedSecret(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registry.Spec.Storage.Driver.Swift.PasswordRef,
			Namespace: registry.GetNamespace(),
		},
	}, harbormetav1.SecretTypeSingle)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot add external typed secret %s", registry.Spec.Storage.Driver.Swift.PasswordRef)
	}

	resources := []graph.Resource{secret}

	if registry.Spec.Storage.Driver.Swift.SecretKeyRef != "" {
		secret, err := r.Controller.AddExternalTypedSecret(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registry.Spec.Storage.Driver.Swift.SecretKeyRef,
				Namespace: registry.GetNamespace(),
			},
		}, harbormetav1.SecretTypeSingle)

		return append(resources, secret), errors.Wrapf(err, "cannot add external typed secret %s", registry.Spec.Storage.Driver.Swift.SecretKeyRef)
	}

	return resources, nil
}
