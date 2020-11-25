package storage

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
)

func (m *MinIOController) Delete(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	minioCR := m.generateMinIOCR()
	err := m.KubeClient.Delete(minioCR)
	if err != nil {
		return minioUnknownStatus(), err
	}
	return nil, nil
}
