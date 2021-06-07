package database

import (
	"context"
	"fmt"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/database/api"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Update reconcile will update PostgreSQL CR.
func (p *PostgreSQLController) Update(ctx context.Context, harborcluster *goharborv1.HarborCluster, curUnstructured *unstructured.Unstructured) (*lcm.CRStatus, error) {
	expectUnstructuredCR, err := p.GetPostgresCR(ctx, harborcluster)
	if err != nil {
		return databaseNotReadyStatus(GenerateDatabaseCrError, err.Error()), err
	}

	name := fmt.Sprintf("%s-%s", harborcluster.Namespace, harborcluster.Name)
	crdClient := p.DClient.DynamicClient(ctx, k8s.WithResource(databaseGVR), k8s.WithNamespace(harborcluster.Namespace))

	var actualCR, expectCR api.Postgresql

	if err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(curUnstructured.UnstructuredContent(), &actualCR); err != nil {
		return databaseNotReadyStatus(DefaultUnstructuredConverterError, err.Error()), err
	}

	if err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(expectUnstructuredCR.UnstructuredContent(), &expectCR); err != nil {
		return databaseNotReadyStatus(DefaultUnstructuredConverterError, err.Error()), err
	}

	if !common.Equals(ctx, p.Scheme, harborcluster, &actualCR) {
		p.Log.Info(
			"Update Database resource",
			"namespace", harborcluster.Namespace, "name", name,
		)

		expectCR.ObjectMeta.SetResourceVersion(actualCR.ObjectMeta.GetResourceVersion())

		data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&expectCR)
		if err != nil {
			return databaseNotReadyStatus(DefaultUnstructuredConverterError, err.Error()), err
		}

		_, err = crdClient.Update(&unstructured.Unstructured{Object: data}, metav1.UpdateOptions{})
		if err != nil {
			return databaseNotReadyStatus(UpdateDatabaseCrError, err.Error()), err
		}
	}

	return databaseUnknownStatus(), nil
}
