package common

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/harbor/components"
)

func (r *Controller) ApplyMutationFunc(ctx context.Context, harbor *goharborv1alpha1.Harbor, resource components.Resource, result metav1.Object, mutate controllerutil.MutateFn) func() error {
	return func() error {
		// Immutable field
		resourceVersion := result.GetResourceVersion()

		defer func() { result.SetResourceVersion(resourceVersion) }()

		// Keep old annotation
		// Often used by other controllers
		annotations := result.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}

		defer func() {
			for key, value := range result.GetAnnotations() {
				annotations[key] = value
			}

			result.SetAnnotations(annotations)

			r.MutateAnnotations(ctx, result)
		}()

		// Keep old labels
		labels := result.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}

		defer func() {
			for key, value := range result.GetLabels() {
				labels[key] = value
			}

			result.SetLabels(labels)

			r.MutateLabels(ctx, result)
		}()

		err := mutate()
		if err != nil {
			return errors.Wrap(err, "cannot mutate resource")
		}

		// Set Harbor instance as the owner and controller of the resource
		err = controllerutil.SetControllerReference(harbor, result, r.Scheme)

		return errors.Wrapf(err, "cannot set controller reference for %s/%s", resource.GroupVersionKind().GroupKind(), resource.GetName())
	}
}

func (r *Controller) ApplyResource(ctx context.Context, harbor *goharborv1alpha1.Harbor, resource components.Resource, objectFactory components.ResourceFactory, objectMutation components.ResourceMutationGetter) (components.Resource, error) {
	kind, version := resource.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()

	span, ctx := opentracing.StartSpanFromContext(ctx, "deployResource", opentracing.Tags{
		"Resource.Kind":    kind,
		"Resource.Version": version,
	})
	defer span.Finish()

	// Get resource result from factory
	result := objectFactory()
	result.SetName(resource.GetName())
	result.SetNamespace(resource.GetNamespace())

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, result, r.ApplyMutationFunc(ctx, harbor, resource, result, objectMutation(resource, result)))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create/update %s/%s", resource.GroupVersionKind().GroupKind(), resource.GetName())
	}

	span.SetTag("Resource.Operation", op)

	return result, nil
}

func (r *Controller) ApplyResources(ctx context.Context, harbor *goharborv1alpha1.Harbor, resources []components.Resource, objectFactory func() components.Resource, objectMutation func(components.Resource, components.Resource) controllerutil.MutateFn) error {
	var g errgroup.Group

	for _, resource := range resources {
		resource := resource

		g.Go(func() error {
			_, err := r.ApplyResource(ctx, harbor, resource, objectFactory, objectMutation)
			return err
		})
	}

	return g.Wait()
}

func mutateSecret(secretResource, result components.Resource) controllerutil.MutateFn {
	secretResult, ok := result.(*corev1.Secret)
	secret := secretResource.(*corev1.Secret)

	return func() error {
		if !ok {
			return errors.Errorf("unexpected argument %+v", result)
		}

		// Most of password are generated
		// Do not override existing secrets
		// To update secrets value, we should rename the key or
		//  delete it before recreating it.
		for key := range secretResult.Data {
			_, okString := secret.StringData[key]
			_, okBytes := secret.Data[key]

			if !okString && !okBytes {
				delete(secretResult.Data, key)
			}

			delete(secret.Data, key)
			delete(secret.StringData, key)
		}

		if secretResult.Data == nil {
			secretResult.Data = map[string][]byte{}
		}

		for name, value := range secret.Data {
			secretResult.Data[name] = value
		}

		// StringData is write only, it overrides Data
		// so we can compare with remote secretResult
		secretResult.StringData = secret.StringData

		return nil
	}
}

func mutateCertificate(certificateResource, result components.Resource) controllerutil.MutateFn {
	certificateResult, ok := result.(*certv1.Certificate)
	certificate := certificateResource.(*certv1.Certificate)

	return func() error {
		if !ok {
			return errors.Errorf("unexpected argument %+v", result)
		}

		certificate.DeepCopyInto(certificateResult)

		return nil
	}
}

func mutateService(serviceResource, result components.Resource) controllerutil.MutateFn {
	serviceResult, ok := result.(*corev1.Service)
	service := serviceResource.(*corev1.Service)

	return func() error {
		if !ok {
			return errors.Errorf("unexpected argument %+v", result)
		}

		// Immutable field
		clusterIP := serviceResult.Spec.ClusterIP

		defer func() { serviceResult.Spec.ClusterIP = clusterIP }()

		for _, port := range serviceResult.Spec.Ports {
			port := port

			defer func() {
				ports := make([]corev1.ServicePort, len(serviceResult.Spec.Ports))

				for i, p := range serviceResult.Spec.Ports {
					if p.Name == port.Name {
						p.NodePort = port.NodePort
					}

					ports[i] = p
				}

				serviceResult.Spec.Ports = ports
			}()
		}

		service.DeepCopyInto(serviceResult)

		return nil
	}
}

func mutateIngress(ingressResource, result components.Resource) controllerutil.MutateFn {
	ingressResult, ok := result.(*netv1.Ingress)
	ingress := ingressResource.(*netv1.Ingress)

	return func() error {
		if !ok {
			return errors.Errorf("unexpected argument %+v", result)
		}

		ingress.DeepCopyInto(ingressResult)

		return nil
	}
}

func mutateDeployment(deploymentResource, result components.Resource) controllerutil.MutateFn {
	deploymentResult, ok := result.(*appsv1.Deployment)
	deployment := deploymentResource.(*appsv1.Deployment)

	return func() error {
		if !ok {
			return errors.Errorf("unexpected argument %+v", result)
		}

		deployment.DeepCopyInto(deploymentResult)

		return nil
	}
}

func mutateConfigMap(configResource, result components.Resource) controllerutil.MutateFn {
	configResult, ok := result.(*corev1.ConfigMap)
	config := configResource.(*corev1.ConfigMap)

	return func() error {
		if !ok {
			return errors.Errorf("unexpected argument %+v", result)
		}

		config.DeepCopyInto(configResult)

		return nil
	}
}

// +kubebuilder:rbac:groups="",resources="configmaps",verbs=get;list;watch;update;patch;create
// +kubebuilder:rbac:groups="",resources="secrets",verbs=get;list;watch;update;patch;create
// +kubebuilder:rbac:groups="cert-manager.io",resources="certificates",verbs=get;list;watch;update;patch;create
// +kubebuilder:rbac:groups="",resources="services",verbs=get;list;watch;update;patch;create
// +kubebuilder:rbac:groups="networking.k8s.io",resources="ingresses",verbs=get;list;watch;update;patch;create
// +kubebuilder:rbac:groups="apps",resources="deployments",verbs=get;list;watch;update;patch;create

func (r *Controller) ApplyComponent(ctx context.Context, harbor *goharborv1alpha1.Harbor, component *components.ComponentRunner) error {
	service := func(ctx context.Context, harbor *goharborv1alpha1.Harbor, resources []components.Resource) error {
		return r.ApplyResources(ctx, harbor, resources, func() components.Resource { return &corev1.Service{} }, mutateService)
	}
	configMap := func(ctx context.Context, harbor *goharborv1alpha1.Harbor, resources []components.Resource) error {
		return r.ApplyResources(ctx, harbor, resources, func() components.Resource { return &corev1.ConfigMap{} }, mutateConfigMap)
	}
	ingress := func(ctx context.Context, harbor *goharborv1alpha1.Harbor, resources []components.Resource) error {
		return r.ApplyResources(ctx, harbor, resources, func() components.Resource { return &netv1.Ingress{} }, mutateIngress)
	}
	secret := func(ctx context.Context, harbor *goharborv1alpha1.Harbor, resources []components.Resource) error {
		return r.ApplyResources(ctx, harbor, resources, func() components.Resource { return &corev1.Secret{} }, mutateSecret)
	}
	certificate := func(ctx context.Context, harbor *goharborv1alpha1.Harbor, resources []components.Resource) error {
		return r.ApplyResources(ctx, harbor, resources, func() components.Resource { return &certv1.Certificate{} }, mutateCertificate)
	}
	deployment := func(ctx context.Context, harbor *goharborv1alpha1.Harbor, resources []components.Resource) error {
		return r.ApplyResources(ctx, harbor, resources, func() components.Resource { return &appsv1.Deployment{} }, mutateDeployment)
	}

	return component.ParallelRun(ctx, harbor, service, configMap, ingress, secret, certificate, deployment, true)
}

func (r *Controller) Apply(ctx context.Context, harbor *goharborv1alpha1.Harbor) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "apply")
	defer span.Finish()

	harborResource, err := components.GetComponents(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot get resources to manage")
	}

	var g errgroup.Group

	if harbor.Spec.Components.Clair == nil {
		g.Go(func() error {
			err := r.DeleteComponent(ctx, harbor, goharborv1alpha1.ClairName)
			return errors.Wrap(err, "cannot delete clair")
		})
	}

	if harbor.Spec.Components.Notary == nil {
		g.Go(func() error {
			err := r.DeleteComponent(ctx, harbor, goharborv1alpha1.NotaryName)
			return errors.Wrap(err, "cannot delete notary")
		})
	}

	g.Go(func() error {
		err = harborResource.ParallelRun(ctx, harbor, r.ApplyComponent)
		return errors.Wrap(err, "cannot deploy component")
	})

	return g.Wait()
}
