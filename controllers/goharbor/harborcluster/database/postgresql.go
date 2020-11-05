package database

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/lcm"
)

type PostgreSQLReconciler struct {
}

func (p PostgreSQLReconciler) Apply(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (p PostgreSQLReconciler) Delete(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (p PostgreSQLReconciler) Upgrade(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}
