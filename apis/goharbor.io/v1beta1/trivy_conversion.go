package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Hub = &Trivy{}

func (t *Trivy) Hub() {
}
