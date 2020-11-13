package database

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/k8s"
	"github.com/goharbor/harbor-operator/pkg/lcm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type PostgreSQLController struct {
	HarborCluster *goharborv1alpha2.HarborCluster
	Ctx           context.Context
	Client        k8s.Client
	Log           logr.Logger
	DClient       k8s.DClient
	Scheme        *runtime.Scheme
	ExpectCR      *unstructured.Unstructured
	ActualCR      *unstructured.Unstructured
}

func (p *PostgreSQLController) Apply(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (p *PostgreSQLController) Delete(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (p *PostgreSQLController) Upgrade(harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func NewDatabaseController(ctx context.Context, options ...k8s.Option) lcm.Controller {

	o := &k8s.CtrlOptions{}

	for _, option := range options {
		option(o)
	}

	return &PostgreSQLController{
		Ctx:     ctx,
		Client:  o.Client,
		Log:     o.Log,
		DClient: o.DClient,
		Scheme:  o.Scheme,
	}
}
