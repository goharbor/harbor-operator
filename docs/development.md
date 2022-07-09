# Development

## Get the sources

```bash
git clone https://github.com/goharbor/harbor-operator.git
cd harbor-operator
```

You developped a new cool feature? Fixed an annoying bug? We would be happy to hear from you!

Have a look in [CONTRIBUTING.md](https://github.com/goharbor/harbor-operator/blob/master/CONTRIBUTING.md)

## Dependencies

### Packages

- [Go 1.18+](https://golang.org/)
- [Helm](https://helm.sh/)
- [Docker](https://docker.com) & [Docker Compose](https://docs.docker.com/compose/install/)
- [OpenSSL](https://www.openssl.org/)
- [npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)

Install dev tools with:

```bash
make dev-tools
```

### Services

- [Kubernetes cluster](https://kubernetes.io). You can order a Kubernetes cluster on [ovh.com](https://www.ovh.com/fr/public-cloud/kubernetes/). Then configure your environment to have the right [`Kubeconfig`](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/).
- Jaeger server.
  
  ```bash
  docker-compose up -d
  ```

### Kubernetes resources

```bash
make install-dependencies # This will install Helm charts in your current cluster
make install # Install Custom Resource Definition
```

### Configuration

Generate resources using `make generate`

## Run it

```bash
export CONFIGURATION_FROM="file:$(pwd)/config-dev.yml"
make run
```

## Deploy a harbor instance

```bash
export LBAAS_DOMAIN=$(kubectl get svc nginx-nginx-ingress-controller -o jsonpath={.status.loadBalancer.ingress[0].hostname})
export CORE_DATABASE_SECRET=$(kubectl get secret core-database-postgresql -o jsonpath={.data.postgresql-password} | base64 --decode)
export CLAIR_DATABASE_SECRET=$(kubectl get secret clair-database-postgresql -o jsonpath={.data.postgresql-password} | base64 --decode)
export NOTARY_SERVER_DATABASE_SECRET=$(kubectl get secret notary-server-database-postgresql -o jsonpath={.data.postgresql-password} | base64 --decode)
export NOTARY_SIGNER_DATABASE_SECRET=$(kubectl get secret notary-signer-database-postgresql -o jsonpath={.data.postgresql-password} | base64 --decode)
kubectl kustomize config/samples | gomplate | kubectl apply -f -

cat <<EOF

Admin password: $(kubectl get secret admin-password-secret -o jsonpath={.data.password} | base64 --decode)
EOF
```

## Linters

```bash
make lint
```

## Tests

__Warning__: Some resource may not be deleted if a test fails

 1. Export `KUBECONFIG` variable:

    ```bash
    export KUBECONFIG=/path/to/kubeconfig
    export USE_EXISTING_CLUSTER=true
    ```

 2. ```bash
    make test
    ```
