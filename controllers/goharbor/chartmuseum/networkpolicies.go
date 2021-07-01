package chartmuseum

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type NetworkPolicy graph.Resource

func (r *Reconciler) AddNetworkPolicies(ctx context.Context, chartMuseum *goharborv1.ChartMuseum) error {
	areNetworkPoliciesEnabled, err := r.AreNetworkPoliciesEnabled(ctx, chartMuseum)
	if err != nil {
		return errors.Wrapf(err, "cannot get status")
	}

	if !areNetworkPoliciesEnabled {
		return nil
	}

	_, err = r.AddIngressNetworkPolicy(ctx, chartMuseum)
	if err != nil {
		return errors.Wrapf(err, "ingress")
	}

	return nil
}

func (r *Reconciler) AddIngressNetworkPolicy(ctx context.Context, chartmuseum *goharborv1.ChartMuseum) (NetworkPolicy, error) {
	networkPolicy, err := r.GetIngressNetworkPolicy(ctx, chartmuseum)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetIngressNetworkPolicy(ctx context.Context, chartmuseum *goharborv1.ChartMuseum) (*netv1.NetworkPolicy, error) {
	var port intstr.IntOrString

	if chartmuseum.Spec.Server.TLS != nil {
		port = intstr.FromString(harbormetav1.ChartMuseumHTTPSPortName)
	} else {
		port = intstr.FromString(harbormetav1.ChartMuseumHTTPPortName)
	}

	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, chartmuseum.GetName(), "ingress"),
			Namespace: chartmuseum.GetNamespace(),
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
					r.Label("name"): r.NormalizeName(ctx, chartmuseum.GetName()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}
