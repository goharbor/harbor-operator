package database

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Deploy reconcile will deploy database cluster if that does not exist.
// It does:
// - check postgres.does exist
// - create any new postgresqls.acid.zalan.do CRs
// - create postgres connection secret
// It does not:
// - perform any postgresqls downscale (left for downscale phase)
// - perform any postgresqls upscale (left for upscale phase)
// - perform any pod upgrade (left for rolling upgrade phase).
func (p *PostgreSQLController) Deploy(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	var expectCR *unstructured.Unstructured

	crdClient := p.DClient.DynamicClient(ctx, k8s.WithResource(databaseGVR), k8s.WithNamespace(harborcluster.Namespace))

	expectCR, err := p.GetPostgresCR(ctx, harborcluster)
	if err != nil {
		return databaseNotReadyStatus(GenerateDatabaseCrError, err.Error()), err
	}

	if err := controllerutil.SetControllerReference(harborcluster, expectCR, p.Scheme); err != nil {
		return databaseNotReadyStatus(SetOwnerReferenceError, err.Error()), err
	}

	resName := p.resourceName(harborcluster.Namespace, harborcluster.Name)

	p.Log.Info("Creating Database.", "namespace", harborcluster.Namespace, "name", resName)

	_, err = crdClient.Create(expectCR, metav1.CreateOptions{})
	if err != nil {
		return databaseNotReadyStatus(CreateDatabaseCrError, err.Error()), err
	}

	p.Log.Info("Database create complete.", "namespace", harborcluster.Namespace, "name", resName)

	return databaseUnknownStatus(), nil
}
