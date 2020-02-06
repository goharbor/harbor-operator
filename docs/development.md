# Development

## Dependencies

### Packages

- [Go 1.13+](https://golang.org/)
- [Helm](https://helm.sh/)
- [Docker](https://docker.com) & [Docker Compose](https://docs.docker.com/compose/install/)
- [OpenSSL](https://www.openssl.org/)
- [npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)

Install dev tools with:

```bash
make dev-tools
```

### Services

- Kubernetes cluster. You can order a Kubernetes cluster on [ovh.com](https://www.ovh.com/fr/public-cloud/kubernetes/).
- [CertManager](https://cert-manager.io/docs/installation/kubernetes/#steps) >= 0.11

  ```bash
  helm install cert-manager \
    --namespace cert-manager \
    --version v0.12.0 \
    jetstack/cert-manager
  ```

- Jaeger server.
  
  ```bash
  docker-compose up -d
  ```

### Kubernetes resources

```bash
make install-dependencies # This will install Helm charts in your current cluster
make install # Install Custom Resource Definition
```

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

Run tests

```bash
export USE_EXISTING_CLUSTER=true
make test
```
