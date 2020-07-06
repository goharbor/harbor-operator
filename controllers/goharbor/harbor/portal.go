package harbor

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
)

func (r *Reconciler) AddPortal(ctx context.Context, harbor *goharborv1alpha2.Harbor) (graph.Resource, error) {
	portal, err := r.GetPortal(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get portal")
	}

	portalRes, err := r.AddBasicResource(ctx, portal)

	return portalRes, errors.Wrap(err, "cannot add basic resource")
}

func (r *Reconciler) GetPortal(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.Portal, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	return &goharborv1alpha2.Portal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.PortalSpec{
			ComponentSpec: harbor.Spec.Portal,
		},
	}, nil
}
