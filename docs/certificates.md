# Certificates

Harbor components required certificates. They are generated thanks to [Certificate resources](https://cert-manager.io/docs/concepts/certificate/).  
To do so, you will need to configure issuer.

## Public certificate

When [deploying the sample](#deploy-the-sample), a [self-signed issuer](https://cert-manager.io/docs/configuration/selfsigned/) is created.

This issue is used to generate the *public certificate*. This one should be trusted by the client.

## WebHook

When [deploying the operator](#deploy-the-operator), a certificate-authority is generated thanks to the [CA-Injector](https://cert-manager.io/docs/concepts/ca-injector/). This is then used by Kubernetes to trust harbor-operator webhook.
