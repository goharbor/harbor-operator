package storage

import "github.com/goharbor/harbor-operator/pkg/cluster/lcm"

func (m *MinIOController) Scale() (*lcm.CRStatus, error) {
	minioCR := m.CurrentMinIOCR
	minioCR.Spec.Zones[0].Servers = m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Replicas

	err := m.KubeClient.Update(minioCR)

	return minioUnknownStatus(), err
}
