package database

import (
	"context"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/ovh/configstore"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PostgreSQLController struct {
	Log         logr.Logger
	DClient     *k8s.DynamicClientWrapper
	Client      client.Client
	Scheme      *runtime.Scheme
	ConfigStore *configstore.Store
}

type Connect struct {
	Host     string
	Port     string
	Password string
	Username string
	Database string
}

func (p *PostgreSQLController) Apply(ctx context.Context, harborcluster *goharborv1.HarborCluster, _ ...lcm.Option) (*lcm.CRStatus, error) {
	crdClient := p.DClient.DynamicClient(ctx, k8s.WithResource(databaseGVR), k8s.WithNamespace(harborcluster.Namespace))

	actualUnstructured, err := crdClient.Get(p.resourceName(harborcluster.Namespace, harborcluster.Name), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return p.Deploy(ctx, harborcluster)
	} else if err != nil {
		return databaseNotReadyStatus(GetDatabaseCrError, err.Error()), err
	}

	if _, err := p.Update(ctx, harborcluster, actualUnstructured); err != nil {
		return databaseNotReadyStatus(CheckDatabaseHealthError, err.Error()), err
	}

	return p.Readiness(ctx, harborcluster, actualUnstructured)
}

func (p *PostgreSQLController) Delete(_ context.Context, _ *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (p *PostgreSQLController) Upgrade(_ context.Context, _ *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func NewDatabaseController(options ...k8s.Option) lcm.Controller {
	o := &k8s.CtrlOptions{}

	for _, option := range options {
		option(o)
	}

	return &PostgreSQLController{
		Log:         o.Log,
		DClient:     o.DClient,
		Client:      o.Client,
		Scheme:      o.Scheme,
		ConfigStore: o.ConfigStore,
	}
}
