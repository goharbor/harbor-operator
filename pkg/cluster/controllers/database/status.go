package database

import (
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	corev1 "k8s.io/api/core/v1"
)

func databaseReadyStatus(reason, message string, properties lcm.Properties) *lcm.CRStatus {
	return lcm.New(goharborv1.DatabaseReady).
		WithStatus(corev1.ConditionTrue).
		WithReason(reason).
		WithMessage(message).
		WithProperties(properties)
}

func databaseNotReadyStatus(reason, message string) *lcm.CRStatus {
	return lcm.New(goharborv1.DatabaseReady).
		WithStatus(corev1.ConditionFalse).
		WithReason(reason).
		WithMessage(message)
}

func databaseUnknownStatus() *lcm.CRStatus {
	return lcm.New(goharborv1.DatabaseReady).
		WithStatus(corev1.ConditionUnknown)
}
