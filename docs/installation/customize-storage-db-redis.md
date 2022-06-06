# Customize the existing storage, database and cache services

Harbor depends on the storage, database and cache services to support its related functionalities. If you have pre-installed dependent services or subscribed cloud services in hand, you can customize the related specs of the deploying Harbor cluster to use those services. The guide documented here shows you how to do the customization of the storage, database and cache services.

## Customize storage

Harbor operator supports configuring `filesystem`(PV), `s3`(as well as `s3` compatible object storage service such as `Minio`) and `swift` as the backend storage of the deploying Harbor cluster.

For customizing the storage spec, you can directly follow the [CRD spec](../CRD/custom-resource-definition.md#storage-related-fields) guideline that has very detailed description.

## Customize database

Harbor uses PostgreSQL as its database to store the related metadata. You can create a database instance from your cloud provider or pre-install a PostgreSQL on your resources. e.g.:

```shell
helm upgrade --install harbor-database bitnami/postgresql --version 10.14.3 --set-string initdbScriptsConfigMap=harbor-init-db --set-string auth.postgresPassword=the-psql-password --set-string image.registry=ghcr.io --set-string image.repository=goharbor/postgresql
```

Here the `initdbScriptsConfigMap` is pointing to a `configMap` used to initialize the databases. e.g.:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    sample: "true"
  name: harbor-init-db
data:
  init-db.sql: |
    CREATE DATABASE core WITH OWNER postgres;
    CREATE DATABASE notaryserver WITH OWNER postgres;
    CREATE DATABASE notarysigner WITH OWNER postgres;
```

>NOTES: `make postgresql` can also help install a PostgreSQL instance quickly.

Get the password of your PostgreSQL instance and wrap it into a Kubernetes secret.

```yaml
apiVersion: v1
kind: Secret
metadata:
  labels:
    sample: "true"
  name: harbor-database-password
data:
  postgresql-password: dGhlLXBzcWwtcGFzc3dvcmQ=
type: goharbor.io/postgresql
```

Then put the related PostgreSQL info into the `database` spec. e.g.:

```yaml
spec:
  database:
    # Configure existing pre-deployed or cloud database service.
    kind: PostgreSQL
    # Database spec
    spec:
      # PostgreSQL configuration spec.
      postgresql:
        # PostgreSQL user name to connect as.
        # Defaults to be the same as the operating system name of the user running the application.
        username: postgres # Required
        # Secret containing the password to be used if the server demands password authentication.
        passwordRef: harbor-database-password # Optional
        # PostgreSQL hosts.
        # At least 1.
        hosts:
        # Name of host to connect to.
        # If a host name begins with a slash, it specifies Unix-domain communication rather than
        # TCP/IP communication; the value is the name of the directory in which the socket file is stored.
        - host: my.psql.com # Required
        # Port number to connect to at the server host,
        # or socket file name extension for Unix-domain connections.
        # Zero, specifies the default port number established when PostgreSQL was built.
        port: 5432 # Optional
        # PostgreSQL has native support for using SSL connections to encrypt client/server communications for increased security.
        # Supports values ["disable","allow","prefer","require","verify-ca","verify-full"].
        sslMode: prefer # Optional, default=prefer
        prefix: prefix # Optional
```

The thing to note here is the names of the databases `core`, `notaryserver` (only needed when enabling notary) and `notarysigner` (only needed when enabling notary) are relatively unchangeable. You can only append some prefixes to the database names by setting the optional field `prefix` in the `database` spec. For example, if the `spec.database.prefix` is "prefix", the database names will be "prefix-core", "prefix-notaryserver" and "prefix-notarysigner".

>NOTES: You need to make sure the related databases have been created before configuring them to the deploying Harbor cluster.

## Customize cache

Harbor uses Redis as its cache service to cache data. You can create a cache instance from your cloud provider or pre-install a Redis on your resources. e.g.:

```shell
helm upgrade --install harbor-redis bitnami/redis --version 15.7.0 --set-string password=the-redis-password --set usePassword=true --set-string image.registry=ghcr.io --set-string image.repository=goharbor/redis
```

>NOTES: `make redis` can also help install a Redis instance quickly.

Get the Redis password and wrap it into a Kubernetes secret. If your Redis instance does not require a password to access, you can skip this creation.

```yaml
apiVersion: v1
data:
  redis-password: dGhlLXJlZGlzLXBhc3N3b3JkCg==
kind: Secret
metadata:
  name: harbor-redis
  labels:
    sample: "true"
type: goharbor.io/redis
```

Then put the related Redis info into the `redis` spec. e.g.:

```yaml
spec:
  kind: Redis
  spec:
    # Redis configuration.
    redis:
      # Server host.
      host: myredis.com # Required
      # Server port.
      port: 6347 # Required
      # For setting sentinel masterSet.
      sentinelMasterSet: sentinel # Optional
      # Secret containing the password to use when connecting to the server.
      passwordRef: harbor-redis # Optional
      # Secret containing the client certificate to authenticate with.
      certificateRef: cert # Optional
```
