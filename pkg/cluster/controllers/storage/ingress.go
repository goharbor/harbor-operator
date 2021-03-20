package storage

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	minio "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/api/v1"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	netv1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (m *MinIOController) applyIngress(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	// expose minIO access endpoint by ingress if necessary.
	if !harborcluster.Spec.InClusterStorage.MinIOSpec.Redirect.Enable {
		m.Log.Info("Redirect of MinIO is not enabled")

		return m.cleanupIngress(ctx, harborcluster)
	}

	// Get current minIO ingress
	curIngress := &netv1.Ingress{}
	err := m.KubeClient.Get(ctx, m.getMinIONamespacedName(harborcluster), curIngress)

	if k8serror.IsNotFound(err) {
		m.Log.Info("Creating minIO ingress")

		return m.createIngress(ctx, harborcluster)
	} else if err != nil {
		return minioNotReadyStatus(GetMinIOIngressError, err.Error()), err
	}

	// Generate desired ingress object
	ingress, err := m.generateIngress(harborcluster)
	if err != nil {
		return minioNotReadyStatus(CreateMinIOIngressError, err.Error()), err
	}

	// Update if necessary
	if !equality.Semantic.DeepDerivative(ingress.DeepCopy().Spec, curIngress.DeepCopy().Spec) {
		m.Log.Info("Updating MinIO ingress")

		if err := m.KubeClient.Update(ctx, ingress); err != nil {
			return minioNotReadyStatus(UpdateIngressError, err.Error()), err
		}

		m.Log.Info("MinIO ingress is updated")
	}

	return minioUnknownStatus(), nil
}

func (m *MinIOController) createIngress(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	// Get the existing minIO CR first
	minioCR := &minio.Tenant{}
	if err := m.KubeClient.Get(ctx, m.getMinIONamespacedName(harborcluster), minioCR); err != nil {
		return minioNotReadyStatus(GetMinIOError, err.Error()), err
	}

	// Generate desired ingress object
	ingress, err := m.generateIngress(harborcluster)
	if err != nil {
		return minioNotReadyStatus(CreateMinIOIngressError, err.Error()), err
	}

	ingress.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(minioCR, HarborClusterMinIOGVK),
	}

	if err := m.KubeClient.Create(ctx, ingress); err != nil {
		return minioNotReadyStatus(CreateMinIOIngressError, err.Error()), err
	}

	m.Log.Info("MinIO ingress is created")

	return minioUnknownStatus(), nil
}

// cleanupIngress cleanups ingress of minio if exist.
func (m *MinIOController) cleanupIngress(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	ingress := &netv1.Ingress{}

	err := m.KubeClient.Get(ctx, m.getMinIONamespacedName(harborcluster), ingress)
	if err != nil {
		if k8serror.IsNotFound(err) {
			// no need cleanup
			return minioUnknownStatus(), nil
		}

		m.Log.Error(err, "Get minio ingress error")

		return minioUnknownStatus(), err
	}

	// clean ingress
	if err = m.KubeClient.Delete(ctx, ingress); err != nil {
		m.Log.Error(err, "Delete minio ingress error")

		return minioUnknownStatus(), err
	}

	return minioUnknownStatus(), nil
}

func (m *MinIOController) generateIngress(harborcluster *goharborv1.HarborCluster) (*netv1.Ingress, error) {
	var tls []netv1.IngressTLS

	if harborcluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose != nil &&
		harborcluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.TLS.Enabled() {
		tls = []netv1.IngressTLS{{
			SecretName: harborcluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.TLS.CertificateRef,
			Hosts:      []string{harborcluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.Ingress.Host},
		}}
	}

	annotations := make(map[string]string)
	annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = "0"

	if harborcluster.Spec.Expose.Core.Ingress.Controller == v1alpha1.IngressControllerNCP {
		annotations["ncp/use-regex"] = "true"
		annotations["ncp/http-redirect"] = "true"
	}

	ingressPath, err := common.GetIngressPath(harborcluster.Spec.Expose.Core.Ingress.Controller)

	return &netv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: netv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.getServiceName(harborcluster),
			Namespace:   harborcluster.Namespace,
			Labels:      m.getLabels(),
			Annotations: annotations,
		},
		Spec: netv1.IngressSpec{
			TLS: tls,
			Rules: []netv1.IngressRule{
				{
					Host: harborcluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.Ingress.Host,
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path: ingressPath,
									Backend: netv1.IngressBackend{
										ServiceName: m.getServiceName(harborcluster),
										ServicePort: intstr.FromInt((int)(m.getServicePort())),
									},
								},
							},
						},
					},
				},
			},
		},
	}, err
}

func (m *MinIOController) getServicePort() int32 {
	return DefaultServicePort
}
