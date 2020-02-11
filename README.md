# Harbor Operator

![githubbanner](https://user-images.githubusercontent.com/3379410/27423240-3f944bc4-5731-11e7-87bb-3ff603aff8a7.png)

Harbor Operator is a [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) which manage Harbor applications. Deploy harbor easily with a single Custom Resource.

[![Docker Hub](https://d36jcksde1wxzq.cloudfront.net/saas-mega/blueFingerprint.png)](https://hub.docker.com/r/ovhcom/harbor-operator)

## Installation

Currently working on a Helm Chart, meanwhile you can use the `make` command.

```bash
git clone https://github.com/ovh/harbor-operator.git
cd harbor-operator
make deploy
```

## Compatibility

This Operator currently only support Harbor version > 1.10.

## Howto's

## Configuration

Generate resources using `make generate`

## Hacking

Follow the [Development guide](https://github.com/ovh/harbor-operator/blob/master/docs/development.md) to start on the project.

## Get the sources

```bash
git clone https://github.com/ovh/harbor-operator.git
cd harbor-operator
```

You've developed a new cool feature ? Fixed an annoying bug ? We'd be happy
to hear from you!

Have a look in [CONTRIBUTING.md](https://github.com/ovh/harbor-operator/blob/master/CONTRIBUTING.md)

## Run the tests

 1. Get a working kubernetes cluster with a valid config file.
    You can get one for free here: <https://www.ovh.com/fr/public-cloud/kubernetes/>

 2. Export `KUBECONFIG` variable:

    ```bash
    export KUBECONFIG=/path/to/kubeconfig
    ```

 3. ```bash
    go test ./...
    ```

## Additional documentation

 1. [Learn how reconciliation works](https://github.com/ovh/harbor-operator/blob/master/docs/reconciler.md)
 2. [Custom Resource Definition](https://github.com/ovh/harbor-operator/blob/master/docs/custom-resource-definition.md)

## Related links

* Contribute: <https://github.com/ovh/harbor-operator/blob/master/CONTRIBUTING.md>
* Report bugs: <https://github.com/ovh/harbor-operator/issues>
* Get latest version: <https://hub.docker.com/r/ovhcom/harbor-operator>

## License

See <https://github.com/ovh/harbor-operator/blob/master/LICENSE>
