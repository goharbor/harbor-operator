package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Hub = &Portal{}

func (*Portal) Hub() {
}
