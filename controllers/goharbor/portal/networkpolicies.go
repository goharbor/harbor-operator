package portal

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

func (r *Reconciler) AddNetworkPolicies(ctx context.Context, portal *goharborv1.Portal) error {
	areNetworkPoliciesEnabled, err := r.AreNetworkPoliciesEnabled(ctx, portal)
	if err != nil {
		return errors.Wrapf(err, "cannot get status")
	}

	if !areNetworkPoliciesEnabled {
		return nil
	}

	_, err = r.AddIngressNetworkPolicy(ctx, portal)
	if err != nil {
		return errors.Wrapf(err, "ingress")
	}

	_, err = r.AddEgressNetworkPolicy(ctx, portal)
	if err != nil {
		return errors.Wrapf(err, "egress")
	}

	return nil
}

func (r *Reconciler) AddIngressNetworkPolicy(ctx context.Context, portal *goharborv1.Portal) (NetworkPolicy, error) {
	networkPolicy, err := r.GetIngressNetworkPolicy(ctx, portal)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetIngressNetworkPolicy(ctx context.Context, portal *goharborv1.Portal) (*netv1.NetworkPolicy, error) {
	var port intstr.IntOrString

	if portal.Spec.TLS != nil {
		port = intstr.FromString(harbormetav1.PortalHTTPSPortName)
	} else {
		port = intstr.FromString(harbormetav1.PortalHTTPPortName)
	}

	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, portal.GetName(), "ingress"),
			Namespace: portal.GetNamespace(),
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
					r.Label("name"): r.NormalizeName(ctx, portal.GetName()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}

func (r *Reconciler) AddEgressNetworkPolicy(ctx context.Context, portal *goharborv1.Portal) (NetworkPolicy, error) {
	networkPolicy, err := r.GetEgressNetworkPolicy(ctx, portal)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetEgressNetworkPolicy(ctx context.Context, portal *goharborv1.Portal) (*netv1.NetworkPolicy, error) {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, portal.GetName(), "egress"),
			Namespace: portal.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Egress: []netv1.NetworkPolicyEgressRule{},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"): r.NormalizeName(ctx, portal.GetName()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeEgress,
			},
		},
	}, nil
}
