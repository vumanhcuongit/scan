#!/usr/bin/env bash

SCRIPTPATH="$(
    cd "$(dirname "$0")"
    pwd -P
)"

CURRENT_DIR=$SCRIPTPATH
ROOT_DIR="$(dirname $CURRENT_DIR)"

. ${CURRENT_DIR}/bootstrap.sh --source-only

function dc_infra() {
    PROJECT_NAME="$(basename $ROOT_DIR)"
    INFRA_COMPOSE_FILE=$ROOT_DIR/deployments/docker-compose.yml
    DEV_COMPOSE_FILE=$ROOT_DIR/deployments/docker-compose.infra.dev.yml
    docker-compose -p $PROJECT_NAME -f $INFRA_COMPOSE_FILE -f $DEV_COMPOSE_FILE $@
}

function infra() {
    case $1 in
    up)
        dc_infra up ${@:2}
        ;;
    down)
        dc_infra down ${@:2}
        ;;
    build)
        dc_infra build ${@:2}
        ;;
    *)
        echo "up|down|build [docker-compose command arguments]"
        ;;
    esac
}

function api_start() {
    echo "Starting infrastructure..."
    infra up -d
    setup_env_variables
    echo "Start api app config file: $CONFIG_FILE"
    ENTRY_FILE="$ROOT_DIR/cmd/app/main.go"
    go run $ENTRY_FILE --config-file=$CONFIG_FILE
}

# Add more command 'migrate' for migrate tool
function api_migrate() {
    echo "Starting migration..."
    infra up -d
    setup_env_variables
    ENTRY_FILE="$ROOT_DIR/cmd/migrate/main.go"
    go run $ENTRY_FILE
}

# Add command to create migraton
function api_create_migration() {
    migrate create -digits 4 -ext sql -dir ${ROOT_DIR}/migrations/sql/ -seq $1
}

# Setup variables environment for app
function setup_ci_env_variables() {
    setup_env_variables
    set -a
    export $(grep -v '^#' "$ROOT_DIR/deployments/.ci.env" | xargs -0) >/dev/null 2>&1
    . $ROOT_DIR/deployments/.ci.env
    set +a
}

# Setup variables environment for app
function setup_env_variables() {
    set -a
    export $(grep -v '^#' "$ROOT_DIR/deployments/.base.env" | xargs -0) >/dev/null 2>&1
    . $ROOT_DIR/deployments/.base.env
    set +a
    export CONFIG_FILE=$ROOT_DIR/configs/app.yaml
    export LOCALE_VI_PATH=$ROOT_DIR/locale/vi/message.yaml
    export LOCALE_EN_PATH=$ROOT_DIR/locale/en/message.yaml
    export GOOGLE_APPLICATION_CREDENTIALS=$ROOT_DIR/configs/service-account.json
}

# generate code from proto
function api_proto_gen() {
    echo ""

    echo "Validating proto changes..."
    echo ""

    if ! command -v buf &> /dev/null; then
        brew tap bufbuild/buf
        brew install buf
    fi

    if ! command -v buf &> /dev/null; then
        echo "Please intall buf command https://buf.build/docs/installation"
        exit 1
    fi

    if [ -f descriptors.bin ]; then
      buf check breaking --against-input descriptors.bin || {
          echo ''
          echo '-> Failed! Updates BROKE backward compatibility.'
          echo ''
          exit 1
      }
    fi

    echo "Compiling protobuf file"

    cd api/v1
    mkdir -p pb
    # generate go code
    protoc \
        -I. \
        -I/usr/local/include \
        -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
        -I$GOPATH/src/github.com/envoyproxy/protoc-gen-validate \
        --descriptor_set_out=./descriptors.bin --include_source_info --include_imports -I. \
        --go_out=plugins=grpc:. \
        proto/*.proto
    mv proto/*.pb.go pb/
    cd -

    cd api/v1/pb
    mockgen -source=shipping.pb.go -destination=shipping_mock.go -package=pb
    cd -

    echo ""
    echo "Done!"
}

function run_gqlgen() {
  command -v gqlgen >/dev/null 2>&1 || {
    go get -u github.com/99designs/gqlgen@v0.10.2
  }

  gqlgen
  sed -i"any" "s/}var/}\\$(echo -e '\n\r')var/g" generated.go
  gofmt -w generated.go
  goimports -w generated.go
  rm -rf generated.goany
}

function api_gqlgen() {
  cd pkg/graphql/api
  run_gqlgen || {
    echo "generate graphql code for api failed"
  }
  cd -
}

function admin_gqlgen() {
  cd pkg/graphql/admin
  run_gqlgen || {
    echo "generate graphql code for api failed"
  }
  cd -
}

# generate code from gqlgen
function gql_gen() {
   api_gqlgen
   admin_gqlgen
}

function test() {
    go clean -testcache ./...

    setup_ci_env_variables
    ENTRY_FILE="$ROOT_DIR/cmd/migrate/main.go"
    go run $ENTRY_FILE
    run_test
}

function api() {
    case $1 in
    proto_gen)
        api_proto_gen
        ;;
    gql_gen)
        gql_gen
        ;;
    test)
        api_test
        ;;
    ci_test)
        api_ci_test
        ;;
    start)
        api_start
        ;;
    graphql)
        api_graphql_start
        ;;
    migrate)
        api_migrate
        ;;
    create_migration)
        api_create_migration ${@:2}
        ;;
    *)
        echo "[build|test|start|graphql|migrate]"
        ;;
    esac
}

function lint() {
    install_golangci_lint
    golangci-lint run ./internal/... ./pkg/... ./cmd/...
}

function code_format() {
    gofmt -w internal/ pkg/ cmd/

    goimports -w internal/ pkg/ cmd/
}

function init() {
    cd $CURRENT_DIR/..
    goimports -w ./..
    go fmt ./...
}

function add_version() {
  DOCKER_FILE=$ROOT_DIR/Dockerfile.prod
  export APP_VERSION="$(cat VERSION).$(git rev-parse --short HEAD)"
  envsubst '${APP_VERSION}' <${DOCKER_FILE}.tmpl >${DOCKER_FILE}
}

function e2e_test() {
    export BUILD_ENV=$1
    curl -XPOST \
        -u "${CIRCLE_CI_USER_TOKEN}:" \
        -H 'Content-Type: application/json' \
        -d "{\"build_parameters\": {\"CIRCLE_JOB\": \"test_${BUILD_ENV}\"}}" \
        "https://circleci.com/api/v1.1/project/github/tboxvn/api-systest/tree/master"
}

function ci() {
    case $1 in
    build)
        ci_build
        ;;
    ignore_if_image_is_build)
        ci_ignore_if_image_is_build
        ;;
    update_image_is_build)
        ci_update_image_is_build
        ;;
    *)
        echo "[ignore_if_image_is_build | update_image_is_build]"
        ;;
    esac
}


function ci_ignore_if_image_is_build() {
    git_sha=$(git rev-parse --short HEAD)
    status=$(curl -XGET "https://keyvalue.immanuel.co/api/KeyVal/GetValue/oze6702m/tbox-shipping-${git_sha}")
    echo $git_sha $status
    if [ $status != '""' ]; then
        echo "Image with commit ${git_sha} is already build. We skip this job"
        circleci-agent step halt
    fi
}

function ci_update_image_is_build() {
    git_sha=$(git rev-parse --short HEAD)
    curl -XPOST "https://keyvalue.immanuel.co/api/KeyVal/UpdateValue/oze6702m/tbox-shipping-${git_sha}/${git_sha}" -d ''
}

case $1 in
init)
    init
    ;;
infra)
    infra ${@:2}
    ;;
api)
    api ${@:2}
    ;;
test)
    test
    ;;
lint)
    lint
    ;;
code_format)
    code_format
    ;;
add_version)
    add_version
    ;;
e2e_test)
    e2e_test ${@:2}
    ;;
ci)
    ci ${@:2}
    ;;
*)
    echo "./scripts/bin.sh [infra|api|lint|add_version]"
    ;;
esac
