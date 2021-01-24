package database

import (
	"context"
	"fmt"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/database/api"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: APIVersion}
	databaseGVR        = SchemeGroupVersion.WithResource(PostgresCRDResourcePlural)
	databaseKind       = "postgresql"
	databaseAPI        = "acid.zalan.do/v1"
	databasePrefix     = "postgresql"
)

// GetPostgresCR returns PostgreSqls CRs.
func (p *PostgreSQLController) GetPostgresCR(ctx context.Context, harborcluster *goharborv1alpha2.HarborCluster) (*unstructured.Unstructured, error) {
	resource := p.GetPostgreResource()
	replica := p.GetPostgreReplica()
	storageSize := p.GetPostgreStorageSize()
	databases := p.GetDatabases()
	storageClass := p.GetStorageClass()

	image, err := p.GetImage(ctx, harborcluster)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	version, err := p.GetPostgreVersion(harborcluster)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get postgresql version")
	}

	conf := &api.Postgresql{
		TypeMeta: metav1.TypeMeta{
			Kind:       databaseKind,
			APIVersion: databaseAPI,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.resourceName(),
			Namespace: p.HarborCluster.Namespace,
		},
		Spec: api.PostgresSpec{
			Volume: api.Volume{
				StorageClass: storageClass,
				Size:         storageSize,
			},
			TeamID:            p.teamID(),
			NumberOfInstances: replica,
			Users:             GetUsers(),
			Patroni:           GetPatron(),
			Databases:         databases,
			PostgresqlParam: api.PostgresqlParam{
				PgVersion:  version,
				Parameters: p.GetPostgreParameters(),
			},
			Resources:   resource,
			DockerImage: image,
		},
	}

	mapResult, err := runtime.DefaultUnstructuredConverter.ToUnstructured(conf)
	if err != nil {
		return nil, err
	}

	data := unstructured.Unstructured{Object: mapResult}

	return &data, nil
}

func GetPatron() api.Patroni {
	return api.Patroni{
		InitDB: GetInitDB(),
		PgHba:  GetPgHba(),
	}
}

func GetPgHba() []string {
	return []string{
		"hostssl all all 0.0.0.0/0 md5",
		"host    all all 0.0.0.0/0 md5",
	}
}

func GetInitDB() map[string]string {
	return map[string]string{
		"encoding":       "UTF8",
		"locale":         "en_US.UTF-8",
		"data-checksums": "true",
	}
}

func GetUsers() map[string]api.UserFlags {
	return map[string]api.UserFlags{
		DefaultDatabaseUser: {
			"superuser",
			"createdb",
		},
	}
}

// GetDatabaseSecret returns database connection secret.
func (p *PostgreSQLController) GetDatabaseSecret(conn *Connect, secretName string) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: p.HarborCluster.Namespace,
		},
		StringData: map[string]string{
			harbormetav1.PostgresqlPasswordKey: conn.Password,
		},
	}

	return secret
}

// resourceName return the formatted name of the managed psql resource.
func (p *PostgreSQLController) resourceName() string {
	return fmt.Sprintf("%s-%s", p.teamID(), p.HarborCluster.Name)
}

// teamID return the team ID of the managed psql service.
func (p *PostgreSQLController) teamID() string {
	return fmt.Sprintf("%s-%s", databasePrefix, p.HarborCluster.Namespace)
}
