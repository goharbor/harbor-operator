package setup

import (
	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

var webhooks = map[ControllerUID]WebHook{
	Harbor:       &goharborv1alpha2.Harbor{},
	JobService:   &goharborv1alpha2.JobService{},
	NotaryServer: &goharborv1alpha2.NotaryServer{},
	NotarySigner: &goharborv1alpha2.NotarySigner{},
	Registry:     &goharborv1alpha2.Registry{},
	Trivy:        &goharborv1alpha2.Trivy{},
}
