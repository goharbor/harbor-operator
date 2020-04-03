# Database installation

Many application of the Harbor stack require a [Postgresql](https://www.postgresql.org) database.

## With helm

Recommended: Install the [postgresql helm chart](https://github.com/bitnami/charts/tree/master/bitnami/postgresql) from bitnami repository.

### Requirements

Checks [chart requirements](https://github.com/bitnami/charts/tree/master/bitnami/postgresql#prerequisites).

### Step by step

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami

export COMPONENT="clair"
```

Please repeat following steps for each components requiring a database: `clair`, `core`, `notary-server`, `notary-signer`, ...

1. Install the helm chart

   ```bash
   # Many parameters can be set to configure the database
   # @see https://github.com/bitnami/charts/tree/master/bitnami/postgresql#parameters
   helm install "$COMPONENT-database" bitnami/postgresql
   ```

2. Get credentials

   ```bash
   export PG_PASSWORD="$(kubectl get secret "$COMPONENT-database-postgresql" -o jsonpath='{.data.postgresql-password}' | base64 --decode)"
   ```

3. Create the computed secret with correct keys (see [`api/v1alpha1/harbor_secret_format.go`](../../api/v1alpha1/harbor_secret_format.go))

   ```bash
   kubectl create secret "$COMPONENT-database" \
      --from-literal host="$COMPONENT-database-postgresql" \
      --from-literal port='5432' \
      --from-literal database='postgres' \
      --from-literal username='postgres' \
      --from-literal password="$PG_PASSWORD"
   ```

The secret is now ready to use in the *Harbor spec*. Please do previous steps for every components requiring a database: `clair`, `core`, `notary-server`, `notary-signer`, ...

```yaml
apiVersion: goharbor.io/v1alpha1
kind: Harbor
metadata:
  ...
spec:
  ...
  components:
    ...
    core:
      databaseSecret: core-database
      ...
    clair:
      databaseSecret: clair-database
      ...
    notary:
      ...
      server:
        databaseSecret: notary-server-database
        ...
      signer:
        databaseSecret: notary-signer-database
        ...
```
