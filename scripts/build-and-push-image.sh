#!/bin/bash
set -ex

#
# Builds and pushes operator, bundle and index images for a given version of an operator
#
# NOTE: Requires make, podman, opm and operator-sdk to be installed!
#

if ! podman login --get-login registry.redhat.io &> /dev/null; then
  echo "Please run podman login registry.redhat.io before running this script."
  exit 1
fi

VERSION=${1:-"0.0.1"}
REPO=${2:-"quay.io/redhat-cne"}
OP_NAME=${3:-"hw-event-proxy-operator"}
IMG="$REPO/$OP_NAME":$VERSION
IMAGE_TAG_BASE="$REPO/$OP_NAME"
INDEX_IMG="$REPO/$OP_NAME-index:$VERSION"

BUNDLE_IMG="$REPO/$OP_NAME-bundle:$VERSION"

# Base operator image
make bin
make manifests
make generate
IMG=${IMG} make docker-build docker-push

rm -Rf bundle
rm -Rf bundle.Dockerfile

# Generate bundle manifests
VERSION=${VERSION} IMG=${IMG} make bundle

# Build bundle image
VERSION=${VERSION} IMAGE_TAG_BASE=${IMAGE_TAG_BASE} make bundle-build

# Push bundle image
VERSION=${VERSION} IMAGE_TAG_BASE=${IMAGE_TAG_BASE} make bundle-push
#opm alpha bundle validate --tag ${BUNDLE_IMG} -b podman

# Index image
opm index add --bundles ${BUNDLE_IMG} --tag ${INDEX_IMG} -u podman --pull-tool podman
podman push ${INDEX_IMG}