package database

import (
	"fmt"
	"strconv"

	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/database/api"
	"github.com/ovh/configstore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	ConfigMaxConnectionsKey       = "max-connections"
	DefaultDatabaseReplica        = 3
	DefaultDatabaseMemory         = "1Gi"
	DefaultDatabaseVersion        = "12"
	DefaultDatabaseMaxConnections = "1024"
)

func (p *PostgreSQLController) GetDatabases() map[string]string {
	databases := map[string]string{
		CoreDatabase: DefaultDatabaseUser,
	}

	if p.HarborCluster.Spec.Notary != nil {
		databases[NotaryServerDatabase] = DefaultDatabaseUser
		databases[NotarySignerDatabase] = DefaultDatabaseUser
	}

	return databases
}

// GetDatabaseConn is getting database connection.
func (p *PostgreSQLController) GetDatabaseConn(secretName string) (*Connect, error) {
	secret, err := p.GetSecret(secretName)
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
func (p *PostgreSQLController) GetStorageClass() string {
	if p.HarborCluster.Spec.InClusterDatabase != nil && p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec != nil {
		return p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec.StorageClassName
	}

	return ""
}

// GetSecret returns the database connection Secret.
func (p *PostgreSQLController) GetSecret(secretName string) (map[string][]byte, error) {
	secret := &corev1.Secret{}

	err := p.Client.Get(types.NamespacedName{Name: secretName, Namespace: p.HarborCluster.Namespace}, secret)
	if err != nil {
		return nil, err
	}

	data := secret.Data

	return data, nil
}

// GetPostgreResource returns postgres resource.
func (p *PostgreSQLController) GetPostgreResource() api.Resources {
	resources := api.Resources{}

	if p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec == nil {
		resources.ResourceRequests = api.ResourceDescription{
			CPU:    "1",
			Memory: "1Gi",
		}
		resources.ResourceRequests = api.ResourceDescription{
			CPU:    "2",
			Memory: "2Gi",
		}

		return resources
	}

	cpu := p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec.Resources.Requests.Cpu()
	mem := p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec.Resources.Requests.Memory()

	request := api.ResourceDescription{}
	if cpu != nil {
		request.CPU = cpu.String()
	}

	if mem != nil {
		request.Memory = mem.String()
	}

	resources.ResourceRequests = request
	resources.ResourceLimits = request

	return resources
}

// GetRedisServerReplica returns postgres replicas.
func (p *PostgreSQLController) GetPostgreReplica() int32 {
	if p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec == nil {
		return DefaultDatabaseReplica
	}

	if p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec.Replicas == 0 {
		return DefaultDatabaseReplica
	}

	return int32(p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec.Replicas)
}

// GetPostgreStorageSize returns Postgre storage size.
func (p *PostgreSQLController) GetPostgreStorageSize() string {
	if p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec == nil {
		return DefaultDatabaseMemory
	}

	if p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec.Storage == "" {
		return DefaultDatabaseMemory
	}

	return p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec.Storage
}

func (p *PostgreSQLController) GetPostgreVersion() string {
	if p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec == nil {
		return DefaultDatabaseVersion
	}

	if p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec.Version == "" {
		return DefaultDatabaseVersion
	}

	return p.HarborCluster.Spec.InClusterDatabase.PostgresSQLSpec.Version
}

func (p *PostgreSQLController) GetPostgreParameters() map[string]string {
	return map[string]string{
		"max_connections": p.GetPosgresMaxConnections(),
	}
}

func (p *PostgreSQLController) GetPosgresMaxConnections() string {
	maxConnections, err := p.ConfigStore.GetItemValue(ConfigMaxConnectionsKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			// Just logged
			p.Log.V(5).Error(err, "failed to get database max connections")
		}

		maxConnections = DefaultDatabaseMaxConnections
	}

	if _, err := strconv.ParseInt(maxConnections, 10, 64); err != nil {
		p.Log.V(5).Error(err, "%s is not a valid number for postgres max connections", maxConnections)

		maxConnections = DefaultDatabaseMaxConnections
	}

	return maxConnections
}

// GenDatabaseURL returns database connection url.
func (c *Connect) GenDatabaseURL() string {
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.Username, c.Password, c.Host, c.Port, c.Database)

	return databaseURL
}
