#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

DEFAULT_DEV_HABOR_OPERATOR_IMAGE=harbor-operator:dev_test
DEFAULT_KIND_VERSION=v0.12.0
DEFAULT_CERT_MANGER_VERSION=1.3.3
DEFAULT_INGRESS_VERSION=1.0.5
DEFAULT_CLUSTER_NAME=harbor
DEFAULT_NODE_IMAGE=kindest/node:v1.23.0
DEFAULT_KUBECTL_VERSION=v1.23.0
PROJECT_ROOT=$(dirname $(dirname $0))
DEFAULT_CONFIG=$PROJECT_ROOT/.github/kind.yaml
DEFAULT_RUNNER_TOOL_CACHE=/usr/local/harbor-operator
DEFAULT_OPERATOR_NAMESPACE=harbor-operator-ns
ETCD_DATA_DIR=/tmp/lib/etcd
LOCAL_SAMPLE_DIR=~/harbor_operator
IP=$(hostname -I | awk '{print $1}')

# check if the docker command is available
check_docker_installed() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "Docker is not installed. Please install docker first."
    exit 1
  fi
}

show_help() {
    cat <<EOF
Usage: $(basename "$0") start [options]
    -h, --help                              Display help
    -v, --version                           The kind version to use (default: $DEFAULT_KIND_VERSION)"
    -c, --config                            The path to the kind config file"
    -i, --node-image                        The Docker image for the cluster nodes"
    -w, --wait                              The duration to wait for the control plane to become ready (default: 60s)"
    -l, --log-level                         The log level for kind [panic, fatal, error, warning, info, debug, trace] (default: warning)
    -k, --kubectl-version                   The kubectl version to use (default: $DEFAULT_KUBECTL_VERSION)"
EOF
}

set_dev_harbor_operator_image() {
    DEFAULT_DEV_HABOR_OPERATOR_IMAGE="${DEFAULT_DEV_HABOR_OPERATOR_IMAGE}_$(date +%s)"
    echo "The harbor operator image is set to $DEFAULT_DEV_HABOR_OPERATOR_IMAGE"
}

cacheDir() {
    local arch=$(uname -m)
    echo "$RUNNER_TOOL_CACHE/tools/$version/$arch"
}
start() {

    shift
    parse_command_line "$@"

    if [[ ! -d "$RUNNER_TOOL_CACHE" ]]; then
        mkdir -p $RUNNER_TOOL_CACHE
    fi

    cache_dir=$(cacheDir)
    local kind_dir="$cache_dir/kind/bin"
    if [[ ! -x "$kind_dir/kind" ]]; then
        install_kind
    fi

    local kubectl_dir="$cache_dir/kubectl/bin"
    if [[ ! -x "$kubectl_dir/kubectl" ]]; then
        install_kubectl
    fi

    "$kind_dir/kind" version
    "$kubectl_dir/kubectl" version --client=true

    mount_memory_etcd

    create_kind_cluster

    install_cert_manager

    install_ingress

    build_load_harbor_operator_image

    install_harbor_operator

    install_harbor

    echo "kubectl location: $kubectl_dir/kubectl"
    echo "kind location: $kind_dir/kind"

    echo "Access the harbor with: https://core.$IP.nip.io, enjoy!"

}

parse_command_line() {
    while :; do
        case "${1:-}" in
        -h | --help)
            show_help
            exit
            ;;
        -v | --version)
            if [[ -n "${2:-}" ]]; then
                version="$2"
                shift
            else
                echo "ERROR: '-v|--version' cannot be empty." >&2
                show_help
                exit 1
            fi
            ;;
        -c | --config)
            if [[ -n "${2:-}" ]]; then
                config="$2"
                shift
            else
                echo "ERROR: '--config' cannot be empty." >&2
                show_help
                exit 1
            fi
            ;;
        -i | --node-image)
            if [[ -n "${2:-}" ]]; then
                node_image="$2"
                shift
            else
                echo "ERROR: '-i|--node-image' cannot be empty." >&2
                show_help
                exit 1
            fi
            ;;
        -w | --wait)
            if [[ -n "${2:-}" ]]; then
                wait="$2"
                shift
            else
                echo "ERROR: '--wait' cannot be empty." >&2
                show_help
                exit 1
            fi
            ;;
        -l | --log-level)
            if [[ -n "${2:-}" ]]; then
                log_level="$2"
                shift
            else
                echo "ERROR: '--log-level' cannot be empty." >&2
                show_help
                exit 1
            fi
            ;;
        -k | --kubectl-version)
            if [[ -n "${2:-}" ]]; then
                kubectl_version="$2"
                shift
            else
                echo "ERROR: '-k|--kubectl-version' cannot be empty." >&2
                show_help
                exit 1
            fi
            ;;
        *)
            break
            ;;
        esac

        shift
    done
}

install_kind() {
    echo 'Installing kind...'

    mkdir -p "$kind_dir"

    curl -sSLo "$kind_dir/kind" "https://github.com/kubernetes-sigs/kind/releases/download/$version/kind-linux-amd64"
    chmod +x "$kind_dir/kind"
}

install_kubectl() {
    echo 'Installing kubectl...'

    mkdir -p "$kubectl_dir"

    curl -sSLo "$kubectl_dir/kubectl" "https://storage.googleapis.com/kubernetes-release/release/$kubectl_version/bin/linux/amd64/kubectl"
    chmod +x "$kubectl_dir/kubectl"
}

create_kind_cluster() {
    echo "Creating kind cluster..."
    if ! "$kind_dir/kind" get clusters | grep $DEFAULT_CLUSTER_NAME >/dev/null; then
        echo 'Creating kind cluster...'
        local args=(create cluster "--name=$DEFAULT_CLUSTER_NAME" "--wait=$wait")

        if [[ -n "$node_image" ]]; then
            args+=("--image=$node_image")
        fi

        if [[ -n "$config" ]]; then
            args+=("--config=$config")
        fi

        if [[ -n "$log_level" ]]; then
            args+=("--loglevel=$log_level")
        fi

        "$kind_dir/kind" "${args[@]}"
    else
        echo "kind cluster $DEFAULT_CLUSTER_NAME already exists"
    fi
}

mount_memory_etcd() {
    echo 'Mounting memory etcd...'
    mkdir -p $ETCD_DATA_DIR
    if ! df -h | grep "^tmpfs.*/tmp/lib/etcd" &>/dev/null; then
        mount -t tmpfs tmpfs $ETCD_DATA_DIR
    fi
}

install_cert_manager() {
    echo 'Installing cert-manager...'
    $kubectl_dir/kubectl apply -f "https://github.com/jetstack/cert-manager/releases/download/v$DEFAULT_CERT_MANGER_VERSION/cert-manager.yaml"
    sleep 5
    time $kubectl_dir/kubectl -n cert-manager wait --for=condition=Available deployment --all --timeout 300s
}

install_ingress() {
    echo 'Installing ingress...'
    $kubectl_dir/kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v$DEFAULT_INGRESS_VERSION/deploy/static/provider/kind/deploy.yaml
    time $kubectl_dir/kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=300s
}

build_load_harbor_operator_image() {
    echo 'Building and load harbor-operator image...'
    cd $PROJECT_ROOT
    make manifests docker-build IMG=$DEFAULT_DEV_HABOR_OPERATOR_IMAGE
    cd - &>/dev/null
    $kind_dir/kind load docker-image $DEFAULT_DEV_HABOR_OPERATOR_IMAGE --name $DEFAULT_CLUSTER_NAME
}

install_harbor_operator() {
    echo 'Installing harbor-operator...'
    cd $PROJECT_ROOT
    make helm-install NAMESPACE=$DEFAULT_OPERATOR_NAMESPACE IMG=$DEFAULT_DEV_HABOR_OPERATOR_IMAGE
    $kubectl_dir/kubectl -n $DEFAULT_OPERATOR_NAMESPACE wait --for=condition=Available deployment --all --timeout 300s

    if ! time $kubectl_dir/kubectl -n $DEFAULT_OPERATOR_NAMESPACE wait --for=condition=Available deployment --all --timeout 300s; then
        $kubectl_dir/kubectl get all -n $DEFAULT_OPERATOR_NAMESPACE
        exit 1
    fi
    cd - &>/dev/null
}

wait_for_condition() {
  time kubectl -n cluster-sample-ns wait --for=condition=$1 harborcluster harborcluster-sample --timeout $2
}

progress(){
    local b=""
    for((i=0;i<=100;i++)); do
        printf "progress: [%-100s] %d%%\r" $b $i
        sleep 2.2
        b+="#"
    done
}

install_harbor() {
    echo 'Installing harbor...'
    cd $PROJECT_ROOT
    CORE_HOST=core.$IP.nip.io
    NOTARY_HOST=notary.$IP.nip.io
    MINIO_HOST=minio.$IP.nip.io

    mkdir -p $LOCAL_SAMPLE_DIR
    cp -f manifests/samples/full_stack.yaml $LOCAL_SAMPLE_DIR
    sed -i s/core.harbor.domain/core.$IP.nip.io/g $LOCAL_SAMPLE_DIR/full_stack.yaml
    sed -i s/notary.harbor.domain/notary.$IP.nip.io/g $LOCAL_SAMPLE_DIR/full_stack.yaml
    sed -i s/minio.harbor.domain/minio.$IP.nip.io/g $LOCAL_SAMPLE_DIR/full_stack.yaml
    sed -i 's/controller: default/controller: contour/g' $LOCAL_SAMPLE_DIR/full_stack.yaml
    sed -i 's/logLevel: info/logLevel: debug/g' $LOCAL_SAMPLE_DIR/full_stack.yaml

    kubectl apply -f $LOCAL_SAMPLE_DIR/full_stack.yaml

    progress
    $kubectl_dir/kubectl get all -n cluster-sample-ns


    if ! wait_for_condition InProgress=False 600s && ! wait_for_condition Failed=False 60s; then
      echo install harbor failed
      kubectl describe harborcluster -n cluster-sample-ns
      kubectl describe harbor -n cluster-sample-ns
      kubectl get all -n cluster-sample-ns

      for n in $(kubectl -n cluster-sample-ns get po |grep -v Running|grep -v NAME|awk '{print $1}');do
        echo describe $n
        kubectl -n cluster-sample-ns describe pod $n
        echo show log $n
        kubectl -n cluster-sample-ns logs --tail 100 $n || true
      done
      kubectl logs -l control-plane=harbor-operator -n $DEFAULT_OPERATOR_NAMESPACE --tail 100
      free -h
      exit 1
    else
      kubectl -n cluster-sample-ns get all -o wide
      kubectl get harbor -n cluster-sample-ns -o wide
      kubectl get harborcluster -n cluster-sample-ns -o wide
      kubectl get all -n $DEFAULT_OPERATOR_NAMESPACE
    fi
    free -h
    df -h
    cd - &>/dev/null
}

check_args_number(){
    if [ $# -lt 1 ]; then
        echo "Usage: $0 start|clean"
        exit 1
    fi
}

clean(){
    echo "Cleaning up..."

    echo "delete sample yaml"
    rm -fr $LOCAL_SAMPLE_DIR/*
    cache_dir=$(cacheDir)

    echo "delete kind cluster"
    local kind_dir="$cache_dir/kind/bin"
    if "$kind_dir/kind" get clusters | grep $DEFAULT_CLUSTER_NAME >/dev/null; then
        $cache_dir/kind/bin/kind delete cluster --name $DEFAULT_CLUSTER_NAME
    fi

    echo "delete etcd data"
    rm -fr $ETCD_DATA_DIR/*


}

main(){
    check_docker_installed
    check_args_number $@
    set_dev_harbor_operator_image

    local RUNNER_TOOL_CACHE="$DEFAULT_RUNNER_TOOL_CACHE"
    local version="$DEFAULT_KIND_VERSION"
    local cluster_name="$DEFAULT_CLUSTER_NAME"
    local kubectl_version="$DEFAULT_KUBECTL_VERSION"
    local config="$DEFAULT_CONFIG"
    local node_image="$DEFAULT_NODE_IMAGE"
    local wait=60s
    local log_level=

    case $1 in
        start)
            start $@
            ;;
        clean)
            clean
            ;;
        *)
            echo "Usage: $0 start [options]|clean"
            exit 1
            ;;
    esac
}

# main entry
main $@
