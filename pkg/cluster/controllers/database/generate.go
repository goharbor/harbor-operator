package database

import (
	"context"
	"fmt"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/database/api"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
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
func (p *PostgreSQLController) GetPostgresCR(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*unstructured.Unstructured, error) {
	resource := p.GetPostgreResource(harborcluster)
	replica := p.GetPostgreReplica(harborcluster)
	storageSize := p.GetPostgreStorageSize(harborcluster)
	databases := p.GetDatabases(harborcluster)
	storageClass := p.GetStorageClass(harborcluster)

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
			Name:      p.resourceName(harborcluster.Namespace, harborcluster.Name),
			Namespace: harborcluster.Namespace,
		},
		Spec: api.PostgresSpec{
			Volume: api.Volume{
				StorageClass: storageClass,
				Size:         storageSize,
			},
			TeamID:            p.teamID(harborcluster.Namespace),
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

	dependencies := checksum.New(p.Scheme)
	dependencies.Add(ctx, harborcluster, true)
	dependencies.AddAnnotations(conf)

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
		"host all all 0.0.0.0/0 md5",
		"local all all trust",
		"local replication postgres trust",
		"hostssl replication postgres all md5",
		"local   replication standby trust",
		"hostssl replication standby all md5",
		"hostssl all +zalandos all pam",
		"hostssl all all all md5",
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
func (p *PostgreSQLController) GetDatabaseSecret(conn *Connect, ns, secretName string) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.PostgresqlPasswordKey: conn.Password,
		},
	}

	return secret
}

// resourceName return the formatted name of the managed psql resource.
func (p *PostgreSQLController) resourceName(ns, name string) string {
	return fmt.Sprintf("%s-%s", p.teamID(ns), name)
}

// teamID return the team ID of the managed psql service.
func (p *PostgreSQLController) teamID(ns string) string {
	return fmt.Sprintf("%s-%s", databasePrefix, ns)
}
