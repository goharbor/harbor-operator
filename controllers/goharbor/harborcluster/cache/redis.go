package cache

import (
	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/lcm"
)

type RedisReconciler struct {
}

func (r RedisReconciler) Reconcile(harborCluster *goharborv1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}
