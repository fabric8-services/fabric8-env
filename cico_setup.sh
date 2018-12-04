#!/bin/bash

# functions to generate client
source <(curl -s https://raw.githubusercontent.com/fabric8-services/fabric8-common/master/.scripts/create_push_client.sh)

set -x
set -e

function prepare() {
  make docker-start
  make docker-check-go-format
  make docker-deps
  make docker-generate
  make docker-build
  echo 'CICO: Preparation complete'
}

function run_tests_with_coverage() {
  # Run unit test
  make docker-test-unit

  # Setup for integration-test
  make integration-test-env-prepare
  trap cleanup_env EXIT
  check_postgres_healthiness

  # Run the integration tests
  make docker-test-migration
  make docker-test-integration

  # Run the remote tests
  make docker-test-remote

  # Output coverage
  make docker-coverage-all

  # Upload coverage to codecov.io
  # -t <upload_token> copy from https://codecov.io/gh/fabric8-services/fabric8-env/settings
  bash <(curl -s https://codecov.io/bash) -t 1b0a3526-5545-411d-860d-7afab0fba3e6

  echo "CICO: ran tests and uploaded coverage"
}

function generate_client() {
  SERVICE_NAME=fabric8-env
  PKG_NAME=env
  TOOL_DIR=tool
  GHORG_NAME=fabric8-services
  GHREPO_NAME=fabric8-env-client

  # function in create_push_client.sh
  generate_client_setup ${SERVICE_NAME} ${PKG_NAME} ${TOOL_DIR} ${GHORG_NAME} ${GHREPO_NAME}
}

# Source environment variables of the jenkins slave
function load_jenkins_vars() {
  if [ -e "jenkins-env.json" ]; then
    eval "$(./env-toolkit load -f jenkins-env.json \
              DEVSHIFT_TAG_LEN \
              QUAY_USERNAME \
              QUAY_PASSWORD \
              JENKINS_URL \
              GIT_BRANCH \
              GIT_COMMIT \
              BUILD_NUMBER \
              BUILD_TAG \
              ghprbSourceBranch \
              ghprbActualCommit \
              BUILD_URL \
              ghprbPullId)"
  fi
}

function install_deps() {
  # We need to disable selinux for now, XXX
  /usr/sbin/setenforce 0 || :

  # Get all the deps in
  yum -y install --quiet \
    docker \
    make \
    git \
    curl \
    nc

  service docker start

  echo 'CICO: Dependencies installed'
}

function cleanup_env {
  EXIT_CODE=$?
  echo "CICO: Cleanup environment: Tear down test environment"
  make integration-test-env-tear-down
  echo "CICO: Exiting with $EXIT_CODE"
}

function check_postgres_healthiness(){
  echo "CICO: Waiting for postgresql container to be healthy...";
  while ! docker ps | grep postgres_integration_test | grep -q healthy; do
    printf .;
    sleep 1 ;
  done;
  echo "CICO: postgresql container is HEALTHY!";
}

function deploy() {
  # Login first
  registry="quay.io"

  if [ -n "${QUAY_USERNAME}" -a -n "${QUAY_PASSWORD}" ]; then
    docker login -u ${QUAY_USERNAME} -p ${QUAY_PASSWORD} ${registry}
  else
    echo "Could not login, missing credentials for the registry"
  fi

  # Build fabric8-env-deploy
  make docker-image-deploy

  TAG=$(echo $GIT_COMMIT | cut -c1-${DEVSHIFT_TAG_LEN})

  if [ "$TARGET" = "rhel" ]; then
    tag_push ${registry}/openshiftio/rhel-fabric8-services-fabric8-env:$TAG
    tag_push ${registry}/openshiftio/rhel-fabric8-services-fabric8-env:latest
  else
    tag_push ${registry}/openshiftio/fabric8-services-fabric8-env:$TAG
    tag_push ${registry}/openshiftio/fabric8-services-fabric8-env:latest
  fi

  echo 'CICO: Image pushed, ready to update deployed app'
}

function tag_push() {
  local target=$1
  docker tag fabric8-env-deploy ${target}
  docker push ${target}
}

function cico_setup() {
  load_jenkins_vars;
  install_deps;
  prepare;
}
