#!/bin/bash
kubectl -n ${NAMESPACE:-default} patch deploy $CORE_DEPLOYMENT -p '{"spec":{"template":{"spec":{"containers":[{"name":"core","env":[{"name":"GC_TIME_WINDOW_HOURS","value":"0"}]}]}}}}'
sleep 10
kubectl -n ${NAMESPACE:-default} wait --for=condition=Ready -l job-type!=minio-init pod --all --timeout 600s


git clone https://github.com/goharbor/harbor -b release-2.6.0

# run 'df -h' before test
sed -i '15i\    ${dfout}=  Run  df -h\n    Log To Console  ${dfout}' harbor/tests/resources/APITest-Util.robot

# increase the timeout of the docker client because the performance of pushing images to harbor with minio storage is very poor
sed -i 's/timeout=30/timeout=300/g' harbor/tests/apitests/python/library/docker_api.py


EXCLUDES="--exclude metrics --exclude singularity --exclude proxy_cache --exclude push_cnab --exclude scan_data_export --exclude log_forward"
ROBOT_FILES="/drone/tests/robot-cases/Group1-Nightly/Setup.robot /drone/tests/robot-cases/Group0-BAT/API_DB.robot"
CMD="robot -v DOCKER_USER:$DOCKER_USER -v DOCKER_PWD:$DOCKER_PWD -v ip:$CORE_HOST -v ip1: -v HARBOR_PASSWORD:Harbor12345 -v http_get_ca:true $EXCLUDES $ROBOT_FILES"

E2E_IMAGE="goharbor/harbor-e2e-engine:4.3.0-api"

# mount dir in the host to the /var/lib/docker in the container to improve the performance of the docker deamon
DOCKER_DATA_DIR=`mktemp -d -t docker-XXXXXX`

mkdir -p /var/log/harbor/

docker run -i --rm --privileged -v `pwd`/harbor:/drone -v /var/log/harbor/:/var/log/harbor/ -v $DOCKER_DATA_DIR:/var/lib/docker -w /drone $E2E_IMAGE make swagger_client
docker run -i --rm --privileged -v `pwd`/harbor:/drone -v /var/log/harbor/:/var/log/harbor/ -v $DOCKER_DATA_DIR:/var/lib/docker -w /drone $E2E_IMAGE $CMD

rc=$?

free -m
df -h

rm -rf $DOCKER_DATA_DIR

exit $rc
