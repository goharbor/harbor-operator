package storage

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/lcm"
)

type MinIOController struct {
}

func (m *MinIOController) Apply(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (m *MinIOController) Delete(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (m *MinIOController) Upgrade(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}
