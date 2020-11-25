package database

import (
	"context"

	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/k8s"
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

type Connect struct {
	Host     string
	Port     string
	Password string
	Username string
	Database string
}

func (p *PostgreSQLController) HealthChecker() lcm.HealthChecker {
	panic("implement me")
}

func (p *PostgreSQLController) Apply(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {

	p.Client.WithContext(ctx)
	p.DClient.WithContext(ctx)
	p.HarborCluster = harborcluster

	crdClient := p.DClient.WithResource(databaseGVR).WithNamespace(p.HarborCluster.Namespace)
	name := fmt.Sprintf("%s-%s", p.HarborCluster.Namespace, p.HarborCluster.Name)

	actualCR, err := crdClient.Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return p.Deploy()
	} else if err != nil {
		return databaseNotReadyStatus(GetDatabaseCrError, err.Error()), err
	}

	expectCR, err := p.GetPostgresCR()
	if err != nil {
		return databaseNotReadyStatus(GenerateDatabaseCrError, err.Error()), err
	}

	if err := controllerutil.SetControllerReference(p.HarborCluster, expectCR, p.Scheme); err != nil {
		return databaseNotReadyStatus(SetOwnerReferenceError, err.Error()), err
	}

	p.ActualCR = actualCR
	p.ExpectCR = expectCR

	if _, err := p.Update(); err != nil {
		return databaseNotReadyStatus(CheckDatabaseHealthError, err.Error()), err
	}

	return p.Readiness()
}

func (p *PostgreSQLController) Delete(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (p *PostgreSQLController) Upgrade(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
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
