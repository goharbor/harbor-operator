package registryctl

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkPolicy graph.Resource

func (r *Reconciler) AddNetworkPolicies(ctx context.Context, registryctl *goharborv1.RegistryController) error {
	areNetworkPoliciesEnabled, err := r.AreNetworkPoliciesEnabled(ctx, registryctl)
	if err != nil {
		return errors.Wrapf(err, "cannot get status")
	}

	if !areNetworkPoliciesEnabled {
		return nil
	}

	_, err = r.AddIngressNetworkPolicy(ctx, registryctl)

	return errors.Wrap(err, "ingress")
}

func (r *Reconciler) AddIngressNetworkPolicy(ctx context.Context, registryctl *goharborv1.RegistryController) (NetworkPolicy, error) {
	networkPolicy, err := r.GetIngressNetworkPolicy(ctx, registryctl)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetIngressNetworkPolicy(ctx context.Context, registryctl *goharborv1.RegistryController) (*netv1.NetworkPolicy, error) {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, registryctl.GetName(), "ingress"),
			Namespace: registryctl.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"): r.NormalizeName(ctx, registryctl.GetName()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}
