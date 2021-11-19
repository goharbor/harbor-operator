package graph

import (
	"context"
)

type (
	Resource interface{}
	RunFunc  func(context.Context, Resource) error
)

type Manager interface {
	Run(context.Context) error
	AddResource(context.Context, Resource, []Resource, RunFunc) error
	GetAllResources(context.Context) []Resource
}
