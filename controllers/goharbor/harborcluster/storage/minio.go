package storage

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/lcm"
)

type MinIOReconciler struct {
}

func (m MinIOReconciler) Apply(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (m MinIOReconciler) Delete(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (m MinIOReconciler) Upgrade(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}
