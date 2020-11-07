package storage

import "github.com/goharbor/harbor-operator/pkg/lcm"

func (m *MinIOReconciler) Delete() (*lcm.CRStatus, error) {
	minioCR := m.generateMinIOCR()
	err := m.KubeClient.Delete(minioCR)
	if err != nil {
		return minioUnknownStatus(), err
	}
	return nil, nil
}
