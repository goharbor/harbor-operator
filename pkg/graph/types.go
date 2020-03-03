package graph

import "context"

type Resource interface{}

type Manager interface {
	Run(context.Context, func(context.Context, Resource) error) error
	AddResource(Resource, []Resource) error
}
