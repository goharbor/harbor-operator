package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	//HarborAdminUserKey     = corev1.BasicAuthUsernameKey
	HarborAdminPasswordKey = corev1.BasicAuthPasswordKey
)

const (
	// ipaddress:port[,weight,password,database_index]
	HarborRegistryURLKey       = "url"
	HarborRegistryNamespaceKey = "namespace"
)

const (
	HarborCoreDatabaseHostKey     = "host"
	HarborCoreDatabasePortKey     = "port"
	HarborCoreDatabaseNameKey     = "database"
	HarborCoreDatabaseUserKey     = "username"
	HarborCoreDatabasePasswordKey = "password"
)

const (
	// ipaddress:port[,weight,password,database_index]
	HarborJobServiceBrokerURLKey       = "url"
	HarborJobServiceBrokerNamespaceKey = "namespace"
)

const (
	HarborClairDatabaseHostKey     = "host"
	HarborClairDatabasePortKey     = "port"
	HarborClairDatabaseNameKey     = "database"
	HarborClairDatabaseUserKey     = "username"
	HarborClairDatabasePasswordKey = "password"
	HarborClairDatabaseSSLKey      = "ssl"
)

const (
	HarborClairAdapterBrokerURLKey       = "url"
	HarborClairAdapterBrokerNamespaceKey = "namespace"
)
