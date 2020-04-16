# Registry Storage configuration

The [Registry](https://docs.docker.com/registry) application required a backend to store docker layers and manifests.

Once you choosed the [right storage](https://docs.docker.com/registry/configuration/#storage), add the file to a [secret](https://kubernetes.io/docs/concepts/configuration/secret/):

## Create the secret

To create the secret required by the operator, you must have the desired configuration for the right a *[storage type](https://docs.docker.com/registry/configuration/#storage)* (`azure`, `filesystem`, `gcs`, `s3`, `swift`, `oss`).

Then create the secret in the *same namespace than the Harbor resource*.

```bash
export STORAGE_TYPE='s3'

# Example with s3
# https://docs.docker.com/registry/storage-drivers/s3/
cat <<EOF > /tmp/registry-backend.yaml
accesskey: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
secretkey: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
bucket: harbor-registry
region: GRA
regionendpoint: https://storage.gra.cloud.ovh.net
encrypt: false
keyid: ""
secure: true
skipverify: false
v4auth: true
chunksize: 5242880
multipartcopychunksize: 33554432
multipartcopymaxconcurrency: 100
multipartcopythresholdsize: 33554432
rootdirectory: /files
storageclass: STANDARD
objectacl: private
sessiontoken: ""
EOF

# Add the secret to kubernetes
# You can specify the namespace using -n
kubectl create secret generic 'registry-backend' \
  --from-file "${STORAGE_TYPE}=/tmp/registry-backend.yaml"
```

## Deploy the Harbor resource

In **the same namespace** than the secret, create or update the Harbor resource:

```yaml
apiVersion: goharbor.io/v1alpha1
kind: Harbor
metadata:
  name: my-harbor
spec:
  components:
    registry:
      storageSecret: registry-backend
...
```

Ensure that the pod registry is rescheduled.
