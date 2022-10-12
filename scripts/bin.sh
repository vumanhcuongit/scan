#!/usr/bin/env bash

SCRIPTPATH="$(
    cd "$(dirname "$0")"
    pwd -P
)"

CURRENT_DIR=$SCRIPTPATH
ROOT_DIR="$(dirname $CURRENT_DIR)"

. ${CURRENT_DIR}/bootstrap.sh --source-only


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

function run_test() {
    echo 'Running unit test'
    # note we use -p 1 to make sure that we only test 1 package at the same time
    # if we test >= 2 packages at same time
    # the code to cleanup db could affect each other tests
    go test -p 1 ./... -coverprofile cover.out.tmp || {
        echo 'unit testing failed'
        exit 1
    } && report_test_coverage
}

function api_test() {    
    setup_env_variables
    run_test
}

function report_test_coverage() {
    echo 'Test coverage report'
    cat cover.out.tmp | grep -v ".mock.go" > cover.out
    go tool cover -func cover.out
    coverage=$(go tool cover -func cover.out | tail -1 | tail -c 6 | cut -b 1-2)
    echo $coverage
    if [ "$coverage" -lt 40 ]; then
        echo "Test coverage is too low"
        exit 1
    fi
}

function api() {
    case $1 in    
    test)
        api_test
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
*)
    echo "./scripts/bin.sh [infra|api|lint|add_version]"
    ;;
esac
