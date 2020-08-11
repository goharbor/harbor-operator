package harbor

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

type Trivy graph.Resource

func (r *Reconciler) AddTrivy(ctx context.Context, harbor *goharborv1alpha2.Harbor) (Trivy, error) {
	if harbor.Spec.Trivy == nil {
		return nil, nil
	}

	trivy, err := r.GetTrivy(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	trivyRes, err := r.AddBasicResource(ctx, trivy)

	return Trivy(trivyRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetTrivy(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.Trivy, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	redisDSN := harbor.Spec.RedisDSN(harbormetav1.TrivyRedis)

	return &goharborv1alpha2.Trivy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.TrivySpec{
			ComponentSpec: harbor.Spec.Trivy.ComponentSpec,
			Cache: goharborv1alpha2.TrivyCacheSpec{
				Redis: redisDSN,
			},
		},
	}, nil
}
