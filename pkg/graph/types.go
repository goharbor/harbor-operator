package graph

import "context"

type Resource interface{}

type Manager interface {
	Run(context.Context, func(context.Context, Resource) error) error
	AddResource(context.Context, Resource, []Resource) error
}
