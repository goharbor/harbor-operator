package database

import (
	goharborv2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/lcm"
	corev1 "k8s.io/api/core/v1"
)

func databaseNotReadyStatus(reason, message string) *lcm.CRStatus {
	return lcm.New(goharborv2.DatabaseReady).
		WithStatus(corev1.ConditionFalse).
		WithReason(reason).
		WithMessage(message)
}

func databaseUnknownStatus() *lcm.CRStatus {
	return lcm.New(goharborv2.DatabaseReady).
		WithStatus(corev1.ConditionUnknown)
}
