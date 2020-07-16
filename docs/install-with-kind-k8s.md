# Installation in KIND k8s

- OS ubuntu 18.04, 8G mem 4CPU

- install docker

  ```bash
  apt install docker.io
  ```

- install kind

  ```bash
  curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-$(uname)-amd64
  ```

- install kubectl

  ```bash
  curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
  ```

- create kind cluster

  ```bash
  cat <<EOF | kind create cluster --name mine --config=-
    kind: Cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    nodes:
    - role: control-plane
      kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
            authorization-mode: "AlwaysAllow"
      extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
    - role: worker
    - role: worker
    - role: worker
  EOF
  ```

- install make golang-go npm helm gofmt  golangci-lint  kube-apiserver  kubebuilder  kubectl  kustomize  pkger etcd

  ```bash
  sudo apt install make npm -y
  
  curl https://dl.google.com/go/go1.14.1.linux-amd64.tar.gz | tar xzv
  export PATH=~/go/bin:$PATH
  
  curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
  
  cd harbor-operator
  make dev-tools
  ```

- install cert-manager

  ```bash
  helm repo add jetstack https://charts.jetstack.io
  helm repo add bitnami https://charts.bitnami.com/bitnami

  helm repo update

  kubectl create namespace cert-manager

  helm install cert-manager jetstack/cert-manager --namespace cert-manager --version v0.13.1
  ```

- install harbor-operator-system

  ```bash
  kubectl create namespace harbor-operator-system
  make deploy
  ```

- install nginx-ingess with nodeport

  ```bash
  helm install nginx stable/nginx-ingress --set-string controller.config.proxy-body-size=0 --set controller.service.type=NodePort
  
  kubectl patch deployments nginx-nginx-ingress-controller -p '{"spec":{"template":{"spec":{"containers":[{"name":"nginx-ingress-controller","ports":[{"containerPort":80,"hostPort":80},{"containerPort":443,"hostPort":443}]}],"nodeSelector":{"ingress-ready":"true"},"tolerations":[{"key":"node-role.kubernetes.io/master","operator":"Equal","effect":"NoSchedule"}]}}}}'
  ```

  below command not work yet

  ```bash
  helm install nginx stable/nginx-ingress \
    --set-string 'controller.config.proxy-body-size'=0 \
    --set-string 'controller.nodeSelector.ingress-ready'=true \
    --set 'controller.service.type'=NodePort \
    --set 'controller.tolerations[0].key'=node-role.kubernetes.io/master \
    --set 'controller.tolerations[0].operator'=Equal \
    --set 'controller.tolerations[0].effect'=NoSchedule
  ```

- install redis database

  ```bash
  make install-dependencies
  ```

- install harbor

  ```bash
  IP="$(hostname -I|awk '{print $1}')"
  export LBAAS_DOMAIN="harbor.$IP.xip.io" \
    NOTARY_DOMAIN="harbor.$IP.xip.io" \
    CORE_DATABASE_SECRET="$(kubectl get secret core-database-postgresql -o jsonpath='{.data.postgresql-password}' | base64 --decode)" \
    CLAIR_DATABASE_SECRET="$(kubectl get secret clair-database-postgresql -o jsonpath='{.data.postgresql-password}' | base64 --decode)" \
    NOTARY_SERVER_DATABASE_SECRET="$(kubectl get secret notary-server-database-postgresql -o jsonpath='{.data.postgresql-password}' | base64 --decode)" \
    NOTARY_SIGNER_DATABASE_SECRET="$(kubectl get secret notary-signer-database-postgresql -o jsonpath='{.data.postgresql-password}' | base64 --decode)" ; \
  kubectl kustomize config/samples | gomplate | kubectl apply -f -
  ```

- export self-sign cert

  ```bash
  sudo mkdir -p "/etc/docker/certs.d/$LBAAS_DOMAIN"

  kubectl get secret "$(kubectl get h harbor-sample -o jsonpath='{.spec.tlsSecretName}')" -o jsonpath='{.data.ca\.crt}' \
     | base64 --decode \
     | sudo tee "/etc/docker/certs.d/$LBAAS_DOMAIN/ca.crt"
  ```

- push image

  ```bash
  docker login "$LBAAS_DOMAIN" -u admin -p $(whoami)

  docker tag busybox "$LBAAS_DOMAIN/library/testbusybox"

  docker push "$LBAAS_DOMAIN/library/testbusybox"
  ```

- clean

  ```bash
  kind delete cluster --name mine
  ```
