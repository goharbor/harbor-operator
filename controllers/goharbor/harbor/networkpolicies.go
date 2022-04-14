package harbor

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type NetworkPolicy graph.Resource

func (r *Reconciler) AddNetworkPolicies(ctx context.Context, harbor *goharborv1.Harbor) error {
	areNetworkPoliciesEnabled, err := r.AreNetworkPoliciesEnabled(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "cannot get status")
	}

	if !areNetworkPoliciesEnabled {
		return nil
	}

	_, err = r.AddChartMuseumIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "chartmuseum ingress")
	}

	_, err = r.AddCoreIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "core ingress")
	}

	_, err = r.AddJobServiceIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "jobservice ingress")
	}

	_, err = r.AddNotaryServerIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "notaryserver ingress")
	}

	_, err = r.AddNotarySignerIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "notary signer ingress")
	}

	_, err = r.AddPortalIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "portal ingress")
	}

	_, err = r.AddPortalEgressNetworkPolicy(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "portal egress")
	}

	_, err = r.AddRegistryIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "registry ingress")
	}

	_, err = r.AddRegistryControllerIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "registryctl ingress")
	}

	_, err = r.AddTrivyIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "trivy ingress")
	}

	return nil
}

func (r *Reconciler) AddChartMuseumIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (NetworkPolicy, error) {
	networkPolicy, err := r.GetChartMuseumIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetChartMuseumIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (*netv1.NetworkPolicy, error) {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.ChartMuseum.String(), "ingress"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					controllers.ChartMuseum.Label("name"): r.NormalizeName(ctx, harbor.GetName(), controllers.ChartMuseum.String()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}

func (r *Reconciler) AddCoreIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (NetworkPolicy, error) {
	networkPolicy, err := r.GetCoreIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetCoreIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (*netv1.NetworkPolicy, error) {
	var port intstr.IntOrString

	if harbor.Spec.Expose.Core.TLS != nil {
		port = intstr.FromString(harbormetav1.CoreHTTPSPortName)
	} else {
		port = intstr.FromString(harbormetav1.CoreHTTPPortName)
	}

	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "ingress"),
			Namespace: harbor.GetNamespace(),
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
					r.Label("name"): r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}

func (r *Reconciler) AddJobServiceIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (NetworkPolicy, error) {
	networkPolicy, err := r.GetJobServiceIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetJobServiceIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (*netv1.NetworkPolicy, error) {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.JobService.String(), "ingress"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"): r.NormalizeName(ctx, harbor.GetName(), controllers.JobService.String()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}

func (r *Reconciler) AddNotaryServerIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (NetworkPolicy, error) {
	networkPolicy, err := r.GetNotaryServerIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetNotaryServerIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (*netv1.NetworkPolicy, error) {
	port := intstr.FromString(harbormetav1.NotaryServerAPIPortName)

	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String(), "ingress"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					Ports: []netv1.NetworkPolicyPort{
						{
							Port: &port,
						},
					},
				},
			},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"): r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}

func (r *Reconciler) AddNotarySignerIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (NetworkPolicy, error) {
	networkPolicy, err := r.GetNotarySignerIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetNotarySignerIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (*netv1.NetworkPolicy, error) {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String(), "ingress"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					controllers.NotarySigner.Label("name"): r.NormalizeName(ctx, harbor.GetName(), controllers.NotarySigner.String()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}

func (r *Reconciler) AddPortalIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (NetworkPolicy, error) {
	networkPolicy, err := r.GetPortalIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetPortalIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (*netv1.NetworkPolicy, error) {
	httpPort := intstr.FromString(harbormetav1.PortalHTTPPortName)
	httpsPort := intstr.FromString(harbormetav1.PortalHTTPSPortName)

	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.Portal.String(), "ingress"),
			Namespace: harbor.GetNamespace(),
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
					r.Label("name"): r.NormalizeName(ctx, harbor.GetName(), controllers.Portal.String()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}

func (r *Reconciler) AddPortalEgressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (NetworkPolicy, error) {
	networkPolicy, err := r.GetPortalEgressNetworkPolicy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetPortalEgressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (*netv1.NetworkPolicy, error) {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.Portal.String(), "egress"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Egress: []netv1.NetworkPolicyEgressRule{},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"): r.NormalizeName(ctx, harbor.GetName(), controllers.Portal.String()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeEgress,
			},
		},
	}, nil
}

func (r *Reconciler) AddRegistryIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (NetworkPolicy, error) {
	networkPolicy, err := r.GetRegistryIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetRegistryIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (*netv1.NetworkPolicy, error) {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.Registry.String(), "ingress"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					controllers.Registry.Label("name"): r.NormalizeName(ctx, harbor.GetName(), controllers.Registry.String()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}

func (r *Reconciler) AddRegistryControllerIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (NetworkPolicy, error) {
	networkPolicy, err := r.GetRegistryControllerIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetRegistryControllerIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (*netv1.NetworkPolicy, error) {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.RegistryController.String(), "ingress"),
			Namespace: harbor.GetNamespace(),
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"): r.NormalizeName(ctx, harbor.GetName(), controllers.RegistryController.String()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}

func (r *Reconciler) AddTrivyIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (NetworkPolicy, error) {
	networkPolicy, err := r.GetTrivyIngressNetworkPolicy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	networkPolicyRes, err := r.Controller.AddNetworkPolicyToManage(ctx, networkPolicy)

	return NetworkPolicy(networkPolicyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetTrivyIngressNetworkPolicy(ctx context.Context, harbor *goharborv1.Harbor) (*netv1.NetworkPolicy, error) {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NormalizeName(ctx, harbor.GetName(), controllers.Trivy.String(), "ingress"),
			Namespace: harbor.GetNamespace(),
		},

		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					controllers.Trivy.Label("name"): r.NormalizeName(ctx, harbor.GetName(), controllers.Trivy.String()),
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	}, nil
}
