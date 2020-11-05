package database

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/k8s"
	"github.com/goharbor/harbor-operator/pkg/lcm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

type PostgreSQLReconciler struct {
	HarborCluster *goharborv1alpha2.HarborCluster
	Ctx           context.Context
	Client        k8s.Client
	Recorder      record.EventRecorder
	Log           logr.Logger
	DClient       k8s.DClient
	Scheme        *runtime.Scheme
	ExpectCR      *unstructured.Unstructured
	ActualCR      *unstructured.Unstructured
	Labels        map[string]string
}

type Connect struct {
	Host     string
	Port     string
	Password string
	Username string
	Database string
}

func (p PostgreSQLReconciler) Reconcile(harborCluster *goharborv1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	if p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec == nil {
		return databaseUnknownStatus(), nil
	}

	p.Client.WithContext(p.Ctx)
	p.DClient.WithContext(p.Ctx)

	crStatus, err := p.Apply()
	if err != nil {
		return databaseNotReadyStatus(ApplyDatabaseHealthError, err.Error()), err
	}

	crStatus, err = p.Readiness()
	if err != nil {
		return databaseNotReadyStatus(CheckDatabaseHealthError, err.Error()), err
	}

	return crStatus, nil
}

func (p *PostgreSQLReconciler) Apply() (*lcm.CRStatus, error) {

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

	return p.Update()
}

func (p *PostgreSQLReconciler) Delete() (*lcm.CRStatus, error) {
	panic("implement me")
}

func (p *PostgreSQLReconciler) Scale() (*lcm.CRStatus, error) {
	panic("implement me")
}

func (p *PostgreSQLReconciler) ScaleUp(newReplicas uint64) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (p *PostgreSQLReconciler) ScaleDown(newReplicas uint64) (*lcm.CRStatus, error) {
	panic("implement me")
}
