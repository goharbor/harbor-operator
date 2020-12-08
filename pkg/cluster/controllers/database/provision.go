package database

import (
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
func (p *PostgreSQLController) Deploy() (*lcm.CRStatus, error) {
	var expectCR *unstructured.Unstructured

	crdClient := p.DClient.WithResource(databaseGVR).WithNamespace(p.HarborCluster.Namespace)

	expectCR, err := p.GetPostgresCR()
	if err != nil {
		return databaseNotReadyStatus(GenerateDatabaseCrError, err.Error()), err
	}

	if err := controllerutil.SetControllerReference(p.HarborCluster, expectCR, p.Scheme); err != nil {
		return databaseNotReadyStatus(SetOwnerReferenceError, err.Error()), err
	}

	p.Log.Info("Creating Database.", "namespace", p.HarborCluster.Namespace, "name", p.resourceName())

	_, err = crdClient.Create(expectCR, metav1.CreateOptions{})
	if err != nil {
		return databaseNotReadyStatus(CreateDatabaseCrError, err.Error()), err
	}

	p.Log.Info("Database create complete.", "namespace", p.HarborCluster.Namespace, "name", p.resourceName())

	return databaseUnknownStatus(), nil
}
