#!/bin/bash
set -x

TARGET_NAMESPACE=${TARGET_NAMESPACE:-"hw-event"}
CSV_VERSION=${CSV_VERSION:-"0.0.1"}

oc delete -n ${TARGET_NAMESPACE} csv hw-event-proxy-operator.${CSV_VERSION}
oc delete -n ${TARGET_NAMESPACE} subscription hw-event-proxy-operator-subscription
oc delete -n ${TARGET_NAMESPACE} catalogsource hw-event-proxy-operator-index