package database

import (
	"context"
	"fmt"
	"strconv"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/database/api"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	ConfigMaxConnectionsKey       = "postgresql-max-connections"
	DefaultDatabaseReplica        = 3
	DefaultDatabaseMemory         = "1Gi"
	DefaultDatabaseMaxConnections = "1024"
	baseInt10                     = 10
	basebBitSize                  = 64
)

var postgresqlVersions = map[string]string{
	"*": "12", // TODO: change to postgres 9.6
}

func (p *PostgreSQLController) GetDatabases(harborcluster *goharborv1.HarborCluster) map[string]string {
	databases := map[string]string{
		CoreDatabase: DefaultDatabaseUser,
	}

	if harborcluster.Spec.Notary != nil {
		databases[NotaryServerDatabase] = DefaultDatabaseUser
		databases[NotarySignerDatabase] = DefaultDatabaseUser
	}

	return databases
}

// GetDatabaseConn is getting database connection.
func (p *PostgreSQLController) GetDatabaseConn(ctx context.Context, ns, secretName string) (*Connect, error) {
	secret, err := p.GetSecret(ctx, ns, secretName)
	if err != nil {
		return nil, err
	}

	conn := &Connect{
		Host:     string(secret["host"]),
		Port:     string(secret["port"]),
		Password: string(secret["password"]),
		Username: string(secret["username"]),
		Database: string(secret["database"]),
	}

	return conn, nil
}

// GetStorageClass returns the storage class name.
func (p *PostgreSQLController) GetStorageClass(harborcluster *goharborv1.HarborCluster) string {
	if harborcluster.Spec.Database.Kind == goharborv1.KindDatabaseZlandoPostgreSQL && harborcluster.Spec.Database.Spec.ZlandoPostgreSQL != nil {
		return harborcluster.Spec.Database.Spec.ZlandoPostgreSQL.StorageClassName
	}

	return ""
}

// GetSecret returns the database connection Secret.
func (p *PostgreSQLController) GetSecret(ctx context.Context, ns, secretName string) (map[string][]byte, error) {
	secret := &corev1.Secret{}

	err := p.Client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: ns}, secret)
	if err != nil {
		return nil, err
	}

	data := secret.Data

	return data, nil
}

// GetPostgreResource returns postgres resource.
func (p *PostgreSQLController) GetPostgreResource(harborcluster *goharborv1.HarborCluster) api.Resources {
	resources := api.Resources{}

	if harborcluster.Spec.Database.Spec.ZlandoPostgreSQL == nil {
		return resources
	}

	spec := harborcluster.Spec.Database.Spec.ZlandoPostgreSQL

	resources.ResourceRequests = getResourceDescription(spec.Resources.Requests)
	resources.ResourceLimits = getResourceDescription(spec.Resources.Limits)

	return resources
}

// GetPostgreReplica returns postgres replicas.
func (p *PostgreSQLController) GetPostgreReplica(harborcluster *goharborv1.HarborCluster) int32 {
	if harborcluster.Spec.Database.Spec.ZlandoPostgreSQL == nil {
		return DefaultDatabaseReplica
	}

	if harborcluster.Spec.Database.Spec.ZlandoPostgreSQL.Replicas == 0 {
		return DefaultDatabaseReplica
	}

	return int32(harborcluster.Spec.Database.Spec.ZlandoPostgreSQL.Replicas)
}

// GetPostgreStorageSize returns Postgre storage size.
func (p *PostgreSQLController) GetPostgreStorageSize(harborcluster *goharborv1.HarborCluster) string {
	if harborcluster.Spec.Database.Spec.ZlandoPostgreSQL == nil {
		return DefaultDatabaseMemory
	}

	if harborcluster.Spec.Database.Spec.ZlandoPostgreSQL.Storage == "" {
		return DefaultDatabaseMemory
	}

	return harborcluster.Spec.Database.Spec.ZlandoPostgreSQL.Storage
}

func (p *PostgreSQLController) GetPostgreVersion(harborcluster *goharborv1.HarborCluster) (string, error) {
	for _, harborVersion := range []string{harborcluster.Spec.Version, "*"} {
		if version, ok := postgresqlVersions[harborVersion]; ok {
			return version, nil
		}
	}

	return "", errors.Errorf("postgresql version not found for harbor %s", harborcluster.Spec.Version)
}

func (p *PostgreSQLController) GetPostgreParameters() map[string]string {
	return map[string]string{
		"max_connections": p.GetPosgresMaxConnections(),
	}
}

func (p *PostgreSQLController) GetPosgresMaxConnections() string {
	maxConnections, err := p.ConfigStore.GetItemValue(ConfigMaxConnectionsKey)
	if err != nil {
		if !config.IsNotFound(err, ConfigMaxConnectionsKey) {
			// Just logged
			p.Log.Error(err, "failed to get database max connections")
		}

		maxConnections = DefaultDatabaseMaxConnections
	}

	if _, err := strconv.ParseInt(maxConnections, baseInt10, basebBitSize); err != nil {
		p.Log.Error(err, "%s is not a valid number for postgres max connections", maxConnections)

		maxConnections = DefaultDatabaseMaxConnections
	}

	return maxConnections
}

// GenDatabaseURL returns database connection url.
func (c *Connect) GenDatabaseURL() string {
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.Username, c.Password, c.Host, c.Port, c.Database)

	return databaseURL
}

func getResourceDescription(res corev1.ResourceList) api.ResourceDescription {
	var rd api.ResourceDescription

	if cpu := res.Cpu(); cpu != nil {
		rd.CPU = cpu.String()
	}

	if mem := res.Memory(); mem != nil {
		rd.Memory = mem.String()
	}

	return rd
}
