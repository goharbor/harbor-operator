package storage

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/lcm"
)

func (m *MinIOController) Upgrade(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}
