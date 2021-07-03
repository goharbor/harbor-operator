package harbor

import (
	"context"
	"fmt"

	goharborv1beta1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type CoreLB graph.Resource

func (r *Reconciler) AddCoreLB(ctx context.Context, harbor *goharborv1beta1.Harbor, core Core, portal Portal) (CoreLB, error) {
	lbRes, err := r.Controller.AddServiceToManage(ctx, r.GetCoreLB(ctx, harbor), core, portal)

	return CoreLB(lbRes), errors.Wrap(err, "cannot add core lb service")
}

func (r *Reconciler) GetCoreLB(ctx context.Context, harbor *goharborv1beta1.Harbor) *corev1.Service {
	if harbor.Spec.Expose.Core.LoadBalancer == nil {
		return nil
	}

	if !harbor.Spec.Expose.Core.LoadBalancer.Enable {
		return nil
	}

	core := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String())

	name := fmt.Sprintf("%s-lb", core)
	namespace := harbor.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.CoreHTTPPortName,
				Port:       harbormetav1.HTTPPort,
				TargetPort: intstr.FromString(harbormetav1.CoreHTTPPortName),
				Protocol:   corev1.ProtocolTCP,
			}, {
				Name:       harbormetav1.CoreHTTPSPortName,
				Port:       harbormetav1.HTTPSPort,
				TargetPort: intstr.FromString(harbormetav1.CoreHTTPSPortName),
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: map[string]string{
				r.Label("name"):      core,
				r.Label("namespace"): namespace,
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}
}

type NotaryServerLB graph.Resource

func (r *Reconciler) AddNotaryServerLB(ctx context.Context, harbor *goharborv1beta1.Harbor, server NotaryServer) (NotaryServerLB, error) {
	lbRes, err := r.Controller.AddServiceToManage(ctx, r.GetNotaryServerLB(ctx, harbor), server)

	return NotaryServerLB(lbRes), errors.Wrap(err, "cannot add notary server lb service")
}

func (r *Reconciler) GetNotaryServerLB(ctx context.Context, harbor *goharborv1beta1.Harbor) *corev1.Service {
	if harbor.Spec.Expose.Notary == nil || harbor.Spec.Expose.Notary.LoadBalancer == nil {
		return nil
	}

	if !harbor.Spec.Expose.Notary.LoadBalancer.Enable {
		return nil
	}

	server := r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String())
	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.NotaryServerTLS))

	name := fmt.Sprintf("%s-lb", server)
	namespace := harbor.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.NotaryServerAPIPortName,
				Port:       tls.GetInternalPort(),
				TargetPort: intstr.FromString(harbormetav1.NotaryServerAPIPortName),
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}
}

type NotarySignerLB graph.Resource

func (r *Reconciler) AddNotarySignerLB(ctx context.Context, harbor *goharborv1beta1.Harbor, signer NotarySigner) (NotarySignerLB, error) {
	lbRes, err := r.Controller.AddServiceToManage(ctx, r.GetNotarySignerLB(ctx, harbor), signer)

	return NotarySignerLB(lbRes), errors.Wrap(err, "cannot add notary signer lb service")
}

func (r *Reconciler) GetNotarySignerLB(ctx context.Context, harbor *goharborv1beta1.Harbor) *corev1.Service {
	if harbor.Spec.Expose.Notary == nil || harbor.Spec.Expose.Notary.LoadBalancer == nil {
		return nil
	}

	if !harbor.Spec.Expose.Notary.LoadBalancer.Enable {
		return nil
	}

	server := r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String())
	name := fmt.Sprintf("%s-lb", server)
	namespace := harbor.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.NotarySignerAPIPortName,
				Port:       goharborv1beta1.NotarySignerAPIPort,
				TargetPort: intstr.FromString(harbormetav1.NotarySignerAPIPortName),
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}
}

type PortalLB graph.Resource

func (r *Reconciler) AddPortalLB(ctx context.Context, harbor *goharborv1beta1.Harbor, portal Portal) (NotarySignerLB, error) {
	lbRes, err := r.Controller.AddServiceToManage(ctx, r.GetPortalLB(ctx, harbor), portal)

	return PortalLB(lbRes), errors.Wrap(err, "cannot add portal lb service")
}

func (r *Reconciler) GetPortalLB(ctx context.Context, harbor *goharborv1beta1.Harbor) *corev1.Service {
	if harbor.Spec.Expose.Portal == nil || harbor.Spec.Expose.Portal.LoadBalancer == nil {
		return nil
	}

	if !harbor.Spec.Expose.Portal.LoadBalancer.Enable {
		return nil
	}

	server := r.NormalizeName(ctx, harbor.GetName(), controllers.Portal.String())
	name := fmt.Sprintf("%s-lb", server)
	namespace := harbor.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.PortalHTTPPortName,
				Port:       harbormetav1.HTTPPort,
				TargetPort: intstr.FromString(harbormetav1.PortalHTTPPortName),
				Protocol:   corev1.ProtocolTCP,
			}, {
				Name:       harbormetav1.PortalHTTPSPortName,
				Port:       harbormetav1.HTTPSPort,
				TargetPort: intstr.FromString(harbormetav1.PortalHTTPSPortName),
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}
}
