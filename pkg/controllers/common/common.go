package common

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	client.Client

	Name    string
	Version string

	Scheme *runtime.Scheme
}

func (r *Controller) GetVersion() string {
	return r.Version
}

func (r *Controller) GetName() string {
	return r.Name
}
