#!/bin/bash

set -e

cp bundle/manifests/* manifests/stable
mv manifests/stable/hw-event-proxy-operator.clusterserviceversion.yaml manifests/stable/bare-metal-event-relay.clusterserviceversion.yaml
sed -i -e 's/name: hw-event-proxy-operator.v4.10.0/name: bare-metal-event-relay.v4.10.0/' \
       -e 's/\(description:.*\) hw-event-proxy/\1 Baremetal Hardware Event Proxy/' \
       -e 's/displayName: Hardware Event Proxy Operator/displayName: Bare Metal Event Relay/' \
       manifests/stable/bare-metal-event-relay.clusterserviceversion.yaml
