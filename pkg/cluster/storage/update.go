package storage

import (
	"github.com/goharbor/harbor-operator/pkg/lcm"
)

func (m *MinIOReconciler) Update() (*lcm.CRStatus, error) {
	m.CurrentMinIOCR.Spec = m.DesiredMinIOCR.Spec
	err := m.KubeClient.Update(m.CurrentMinIOCR)
	if err != nil {
		return minioNotReadyStatus(UpdateMinIOError, err.Error()), err
	}

	return minioUnknownStatus(), nil
}
