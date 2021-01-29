package core

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

func (r *Reconciler) AddIngressNetworkPolicy(ctx context.Context, core *goharborv1alpha2.Core) (NetworkPolicy, error) {
	networkPolicy, err := r.GetIngressNetworkPolicy(ctx, core)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetIngressNetworkPolicy(ctx context.Context, core *goharborv1alpha2.Core) (*netv1.NetworkPolicy, error) {
	httpPort := intstr.FromString(harbormetav1.CoreHTTPPortName)
	httpsPort := intstr.FromString(harbormetav1.CoreHTTPSPortName)

	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, core.GetName(), "ingress"),
			Namespace: core.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					Ports: []netv1.NetworkPolicyPort{
						{
							Port: &httpPort,
						},
						{
							Port: &httpsPort,
						},
					},
				},
			},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"): r.NormalizeName(ctx, core.GetName()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}
