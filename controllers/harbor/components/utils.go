package components

import (
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ResourceFactory func() Resource
type ResourceMutationGetter func(Resource, Resource) controllerutil.MutateFn
