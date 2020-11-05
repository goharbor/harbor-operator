package database

import (
	"fmt"

	"github.com/goharbor/harbor-cluster-operator/controllers/database/api"
	pg "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	databaseGVR  = pg.SchemeGroupVersion.WithResource(pg.PostgresCRDResourcePlural)
	databaseKind = "postgresql"
	databaseApi  = "acid.zalan.do/v1"
)

// GetPostgresCR returns PostgreSqls CRs
func (p *PostgreSQLReconciler) GetPostgresCR() (*unstructured.Unstructured, error) {
	resource := p.GetPostgreResource()
	replica := p.GetPostgreReplica()
	storageSize := p.GetPostgreStorageSize()
	version := p.GetPostgreVersion()
	databases := p.GetDatabases()
	name := fmt.Sprintf("%s-%s", p.HarborCluster.Namespace, p.HarborCluster.Name)

	conf := &api.Postgresql{
		TypeMeta: metav1.TypeMeta{
			Kind:       databaseKind,
			APIVersion: databaseApi,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: p.HarborCluster.Namespace,
			Labels:    p.Labels,
		},
		Spec: api.PostgresSpec{
			Volume: api.Volume{
				Size: storageSize,
			},
			TeamID:            p.HarborCluster.Namespace,
			NumberOfInstances: replica,
			Users:             GetUsers(),
			Patroni:           GetPatron(),
			Databases:         databases,
			PostgresqlParam: api.PostgresqlParam{
				PgVersion: version,
			},
			Resources: resource,
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
		InitDB: GetInitDb(),
		PgHba:  GetPgHba(),
	}
}

func GetPgHba() []string {
	return []string{
		"hostssl all all 0.0.0.0/0 md5",
		"host    all all 0.0.0.0/0 md5",
	}
}

func GetInitDb() map[string]string {
	return map[string]string{
		"encoding":       "UTF8",
		"locale":         "en_US.UTF-8",
		"data-checksums": "true",
	}
}

func GetUsers() map[string]api.UserFlags {
	return map[string]api.UserFlags{
		"zalando": {
			"superuser",
			"createdb",
		},
		"foo_user": {},
	}
}

//GetDatabaseSecret returns database connection secret
func (p *PostgreSQLReconciler) GetDatabaseSecret(conn *Connect, secretName, propertyName string) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: p.HarborCluster.Namespace,
			Labels:    p.Labels,
		},
		StringData: map[string]string{
			"host":     conn.Host,
			"port":     conn.Port,
			"database": conn.Database,
			"username": conn.Username,
			"password": conn.Password,
		},
	}

	if propertyName == HarborClair {
		secret.StringData["database"] = ClairDatabase
		secret.StringData["ssl"] = "disable"
	}

	if propertyName == HarborNotaryServer {
		secret.StringData["database"] = NotaryServerDatabase
		secret.StringData["ssl"] = "disable"
	}
	if propertyName == HarborNotarySigner {
		secret.StringData["database"] = NotarySignerDatabase
		secret.StringData["ssl"] = "disable"
	}

	return secret
}
