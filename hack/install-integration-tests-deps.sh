#!/bin/bash

set -e

_GOOS=$(go env GOOS)
_GOARCH=$(go env GOARCH)
_ARCH=$(arch)

mkdir -p ./bin

# Install kubectl if needed.
if ! [ -x "$(command -v kubectl)" ]; then
    echo "Installing kubectl..."
    curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/$_GOOS/$_GOARCH/kubectl
    mv ./kubectl ./bin/
    chmod 755 ./bin/kubectl
else
    echo "No need to install kubectl, continue..."
fi

# Install KUTTL if needed
if ! [ -x "$(command -v kubectl-kuttl)" ]; then
    echo "Installing kubectl-kuttl..."
    curl -LO https://github.com/kudobuilder/kuttl/releases/download/v0.11.0/kubectl-kuttl_0.11.0_${_GOOS}_${_ARCH}
    mv ./kubectl-kuttl_0.11.0_${_GOOS}_${_ARCH} ./bin/kubectl-kuttl
    ls ./bin/
    chmod +x ./bin/kubectl-kuttl
else
    echo "No need to install kubectl-kuttl, continue..."
fi
