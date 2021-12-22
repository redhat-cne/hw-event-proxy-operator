REPO_DIR="$(dirname $0)/.."
OPERATOR_EXEC=oc

export RELEASE_VERSION=v0.0.1
export IMAGE_TAG=latest
export OPERATOR_NAME=hw-event-proxy-operator

HW_EVENT_PROXY_IMAGE_DIGEST=$(skopeo inspect docker://quay.io/redhat-cne/hw-event-proxy | jq --raw-output '.Digest')
CLOUD_EVENT_PROXY_IMAGE_DIGEST=$(skopeo inspect docker://quay.io/openshift/origin-cloud-event-proxy | jq --raw-output '.Digest')
KUBE_RBAC_PROXY_DIGEST=$(skopeo inspect docker://quay.io/openshift/origin-kube-rbac-proxy | jq --raw-output '.Digest')

export HW_EVENT_PROXY_IMAGE=${HW_EVENT_PROXY_IMAGE:-quay.io/redhat-cne/hw-event-proxy@${HW_EVENT_PROXY_IMAGE_DIGEST}}
export KUBE_RBAC_PROXY_IMAGE=${KUBE_RBAC_PROXY_IMAGE:-quay.io/openshift/origin-kube-rbac-proxy@${KUBE_RBAC_PROXY_DIGEST}}
export CLOUD_EVENT_PROXY_IMAGE=${SIDECAR_EVENT_IMAGE:-quay.io/openshift/origin-cloud-event-proxy@${CLOUD_EVENT_PROXY_IMAGE_DIGEST}}
