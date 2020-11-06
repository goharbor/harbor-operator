package harborcluster

import (
	"github.com/goharbor/harbor-operator/pkg/lcm"
)

type ServiceGetter interface {
	// For Redis
	Cache() lcm.Controller

	// For database
	Database() lcm.Controller

	// For storage
	Storage() lcm.Controller

	// For harbor itself
	Harbor() lcm.Controller
}
