package storage

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	miniov2 "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/apis/minio.min.io/v2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NCPIngressValueTrue     = "true"
	ContourIngressValueTrue = "true"
)

func (m *MinIOController) applyIngress(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	// expose minIO access endpoint by ingress if necessary.
	redirect := harborcluster.Spec.Storage.Spec.Redirect
	if redirect == nil && harborcluster.Spec.Storage.Spec.MinIO != nil {
		redirect = harborcluster.Spec.Storage.Spec.MinIO.Redirect
	}

	if redirect == nil || !redirect.Enable {
		m.Log.Info("Redirect of MinIO is not enabled")

		return m.cleanupIngress(ctx, harborcluster)
	} else if redirect.Expose == nil || redirect.Expose.Ingress == nil {
		err := errors.New("Expose.Ingress should be defined when redirect enabled")

		return minioNotReadyStatus(UpdateIngressError, err.Error()), err
	}

	// Get current minIO ingress
	curIngress := &netv1.Ingress{}
	err := m.KubeClient.Get(ctx, m.getMinIONamespacedName(harborcluster), curIngress)

	if k8serror.IsNotFound(err) {
		m.Log.Info("Creating minIO ingress")

		return m.createIngress(ctx, harborcluster, redirect)
	} else if err != nil {
		return minioNotReadyStatus(GetMinIOIngressError, err.Error()), err
	}

	// Generate desired ingress object
	ingress := m.generateIngress(ctx, harborcluster, redirect)

	// Update if necessary
	if !common.Equals(ctx, m.Scheme, harborcluster, curIngress) {
		m.Log.Info("Updating MinIO ingress")

		if err := m.KubeClient.Update(ctx, ingress); err != nil {
			return minioNotReadyStatus(UpdateIngressError, err.Error()), err
		}

		m.Log.Info("MinIO ingress is updated")
	}

	return minioUnknownStatus(), nil
}

func (m *MinIOController) createIngress(ctx context.Context, harborcluster *goharborv1.HarborCluster, redirect *goharborv1.StorageRedirectSpec) (*lcm.CRStatus, error) {
	// Get the existing minIO CR first
	minioCR := &miniov2.Tenant{}
	if err := m.KubeClient.Get(ctx, m.getMinIONamespacedName(harborcluster), minioCR); err != nil {
		return minioNotReadyStatus(GetMinIOError, err.Error()), err
	}

	// Generate desired ingress object
	ingress := m.generateIngress(ctx, harborcluster, redirect)

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

func (m *MinIOController) getMinioIngressAnnotations(redirect *goharborv1.StorageRedirectSpec) map[string]string {
	isEnableExpose := false
	if redirect.Expose != nil {
		isEnableExpose = true
	}

	istls := false
	if isEnableExpose && redirect.Expose.TLS.Enabled() {
		istls = true
	}

	annotations := map[string]string{
		// resolve 413(Too Large Entity) error when push large image. It only works for NGINX ingress.
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	}

	if isEnableExpose && redirect.Expose.Ingress.Controller == harbormetav1.IngressControllerNCP {
		annotations["ncp/use-regex"] = NCPIngressValueTrue
		if istls {
			annotations["ncp/http-redirect"] = NCPIngressValueTrue
		}
	} else if redirect.Expose.Ingress.Controller == harbormetav1.IngressControllerContour {
		if istls {
			annotations["ingress.kubernetes.io/force-ssl-redirect"] = ContourIngressValueTrue
		}
	}

	if isEnableExpose {
		for key, value := range redirect.Expose.Ingress.Annotations {
			annotations[key] = value
		}
	}

	return annotations
}

func (m *MinIOController) generateIngress(ctx context.Context, harborcluster *goharborv1.HarborCluster, redirect *goharborv1.StorageRedirectSpec) *netv1.Ingress {
	var tls []netv1.IngressTLS

	if redirect.Expose != nil && redirect.Expose.TLS.Enabled() {
		tls = []netv1.IngressTLS{{
			SecretName: redirect.Expose.TLS.CertificateRef,
			Hosts:      []string{redirect.Expose.Ingress.Host},
		}}
	}

	annotations := m.getMinioIngressAnnotations(redirect)

	pathTypePrefix := netv1.PathTypePrefix

	ingress := &netv1.Ingress{
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
			TLS:              tls,
			IngressClassName: redirect.Expose.Ingress.IngressClassName,
			Rules: []netv1.IngressRule{
				{
					Host: redirect.Expose.Ingress.Host,
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: m.getTenantsServiceName(harborcluster),
											Port: netv1.ServiceBackendPort{
												Number: m.getServicePort(),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	dependencies := checksum.New(m.Scheme)
	dependencies.Add(ctx, harborcluster, true)
	dependencies.AddAnnotations(ingress)

	return ingress
}

func (m *MinIOController) getServicePort() int32 {
	return DefaultServicePort
}
