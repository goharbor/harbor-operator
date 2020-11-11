package storage

import (
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/lcm"
)

func (m *MinIOController) Upgrade(harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}
