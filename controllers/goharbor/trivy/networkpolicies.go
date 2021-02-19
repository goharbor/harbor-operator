package trivy

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type NetworkPolicy graph.Resource

func (r *Reconciler) AddNetworkPolicies(ctx context.Context, trivy *goharborv1alpha2.Trivy) error {
	areNetworkPoliciesEnabled, err := r.AreNetworkPoliciesEnabled(ctx, trivy)
	if err != nil {
		return errors.Wrapf(err, "cannot get status")
	}

	if !areNetworkPoliciesEnabled {
		return nil
	}

	_, err = r.AddIngressNetworkPolicy(ctx, trivy)

	return errors.Wrap(err, "ingress")
}

func (r *Reconciler) AddIngressNetworkPolicy(ctx context.Context, trivy *goharborv1alpha2.Trivy) (NetworkPolicy, error) {
	networkPolicy, err := r.GetIngressNetworkPolicy(ctx, trivy)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetIngressNetworkPolicy(ctx context.Context, trivy *goharborv1alpha2.Trivy) (*netv1.NetworkPolicy, error) {
	var port intstr.IntOrString

	if trivy.Spec.Server.TLS != nil {
		port = intstr.FromString(harbormetav1.TrivyHTTPSPortName)
	} else {
		port = intstr.FromString(harbormetav1.TrivyHTTPPortName)
	}

	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, trivy.GetName(), "ingress"),
			Namespace: trivy.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					Ports: []netv1.NetworkPolicyPort{{
						Port: &port,
					}},
				},
			},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"): r.NormalizeName(ctx, trivy.GetName()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}
