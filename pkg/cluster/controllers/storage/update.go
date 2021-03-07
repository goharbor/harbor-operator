package storage

import (
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	netv1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
)

func (m *MinIOController) Update() (*lcm.CRStatus, error) {
	m.CurrentMinIOCR.Spec = m.DesiredMinIOCR.Spec

	err := m.KubeClient.Update(m.CurrentMinIOCR)
	if err != nil {
		return minioNotReadyStatus(UpdateMinIOError, err.Error()), err
	}

	return minioUnknownStatus(), nil
}

func (m *MinIOController) checkRedirectUpdate() (bool, error) {
	currntIngress := &netv1.Ingress{}

	err := m.KubeClient.Get(m.getMinIONamespacedName(), currntIngress)
	if k8serror.IsNotFound(err) {
		m.Log.Info("minio ingress not exists.")

		return false, nil
	} else if err != nil {
		return false, err
	}

	desiredingress, err := m.generateIngress()

	if currntIngress.Spec.Rules[0].Host != desiredingress.Spec.Rules[0].Host {
		return true, err
	}

	if currntIngress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0] != desiredingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0] {
		return true, err
	}

	if currntIngress.Spec.TLS[0].Hosts[0] != desiredingress.Spec.TLS[0].Hosts[0] {
		return true, err
	}

	if !equality.Semantic.DeepEqual(currntIngress.DeepCopy().Spec, desiredingress.DeepCopy().Spec) {
		return true, err
	}

	return false, err
}

func (m *MinIOController) updateMinioIngress() error {
	desiredingress, err := m.generateIngress()
	if err != nil {
		return err
	}

	err = m.KubeClient.Update(desiredingress)
	if err != nil {
		return err
	}

	return nil
}
