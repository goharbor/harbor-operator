package harbor

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/lcm"
)

type HarborReconciler struct {
}

// Reconciler implements the reconcile logic of services
func (harbor *HarborReconciler) Reconcile(harborCluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}
