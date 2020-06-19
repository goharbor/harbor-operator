package notaryserver

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

const (
	ConfigMigrationDatabaseSourceKey = "migration-source"
	MigrationSourceSecretKey         = "migration-source"
)

func (r *Reconciler) GetSecret(ctx context.Context, notaryserver *goharborv1alpha2.NotaryServer) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, notaryserver.GetName())
	namespace := notaryserver.GetNamespace()

	migrationDatabaseSource, err := r.ConfigStore.GetItemValue(ConfigMigrationDatabaseSourceKey)
	if err != nil {
		return nil, err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			MigrationSourceSecretKey: migrationDatabaseSource,
		},
	}, nil
}
