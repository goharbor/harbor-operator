package v1beta1

import "sigs.k8s.io/controller-runtime/pkg/conversion"

var _ conversion.Hub = &ChartMuseum{}

func (c *ChartMuseum) Hub() {}
