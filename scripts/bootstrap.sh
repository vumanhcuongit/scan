#!/bin/bash

function install_protoc() {
  # install protoc
  command -v protoc >/dev/null 2>&1 || {
    command -v brew >/dev/null 2>&1 || {
      echo ""
      if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "You could install homebrew by youself if you are using MacOSX by using this command line"
        echo "/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install.sh)\""
      else
        # install protoc if it is ubuntu
        command -v apt-get && {
          apt-get update
          apt-get install unzip -y
          export PROTOC_VERSION="3.12.3"
          PB_REL="https://github.com/protocolbuffers/protobuf/releases"
          curl -LO $PB_REL/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip
          unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d $HOME/.local
          export PATH="$PATH:$HOME/.local/bin"
          exit 0
        }

        echo "Please install protoc by following this document"
        echo "https://grpc.io/docs/quickstart/go/"
      fi
      exit 0
    }

    command -v brew && {
      echo ""
      echo "tkit-cli is installing protoc"
      brew install protobuf@3.12
    }
  }
}

function install_protoc_gen_go() {
  command -v protoc-gen-go >/dev/null 2>&1 || {
    echo ""
    echo "tkit-cli is installing protoc-gen-go"
    cd ..
    go get -u github.com/golang/protobuf/protoc-gen-go
    cd -
  }
}

function install_grpc_gateway() {
  [ -d "$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway" ] || {
    echo ""
    echo "tkit-cli is installing grpc-gateway"
    mkdir -p $GOPATH/src/github.com/grpc-ecosystem
    cd $GOPATH/src/github.com/grpc-ecosystem
    git clone https://github.com/grpc-ecosystem/grpc-gateway.git
    cd -
  }
}

function install_mockgen() {
  command -v mockgen >/dev/null 2>&1 || {
    echo ""
    echo "tkit-cli is installing mockgen"
    go get github.com/golang/mock/mockgen
  }
}

function install_golangci_lint() {
  command -v golangci-lint >/dev/null 2>&1 || {
    echo ""
    echo "tkit-cli is installing golangci-lint"
    curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.40.1
  }
}

function install_goimports() {
  command -v goimports >/dev/null 2>&1 || {
    echo ""
    echo "tkit-cli is installing goimports"
    go get golang.org/x/tools/cmd/goimports
  }
}

function install_gqlgen() {
  command -v gqlgen >/dev/null 2>&1 || {
    echo ""
    echo "tkit-cli is installing gqlgen"
    go get -u github.com/99designs/gqlgen@v0.10.2
  }
}

function install_migrate() {
  command -v migrate >/dev/null 2>&1 || {
    echo ""
    echo "tkit-cli is installing migrate tool"
    go get -u -d github.com/golang-migrate/migrate/cmd/migrate
  }
}

function init_git_repo() {
  echo "init git repository"
  git init
  git add .
  git commit -m "initialize project by tkit-cli"
}

function init() {
  export GO111MODULE=on

  rm -rf pkg/graphql/{admin,api}/schema/\{\{.Service\}\}
  mkdir -p pkg/graphql/{admin,api}/schema/shipping
  chmod +x ./scripts/bin.sh
  install_migrate
  install_gqlgen
  install_golangci_lint
  install_mockgen
  install_grpc_gateway
  install_protoc_gen_go
  install_goimports
  ./scripts/bin.sh code_format
  init_git_repo
  install_protoc
}

if [ ${1} != "--source-only" ]; then
  case $1 in
  init)
    init
    ;;
  *)
    echo "./scripts/bootstrap.sh [init]"
    ;;
  esac
fi
