package trivy

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1alpha2.Trivy{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	trivy, ok := resource.(*goharborv1alpha2.Trivy)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	err := r.AddService(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "service")
	}

	var github graph.Resource

	if trivy.Spec.Update.GithubTokenRef != "" {
		github, err = r.AddExternalTypedSecret(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
			Name:      trivy.Spec.Update.GithubTokenRef,
			Namespace: trivy.GetNamespace(),
		}}, harbormetav1.SecretTypeGithubToken)
		if err != nil {
			return errors.Wrap(err, "github")
		}
	}

	cm, err := r.AddConfigMap(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "configmap")
	}

	secret, err := r.AddSecret(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "secret")
	}

	err = r.AddDeployment(ctx, trivy, cm, secret, github)
	if err != nil {
		return errors.Wrap(err, "deployment")
	}

	return nil
}
