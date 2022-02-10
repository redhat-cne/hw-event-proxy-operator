# Developer Guide

## Build image
```
make build
make bin
make image
```

## Update manifests and bundle
Configure images used by the operator in [config/manager/env.yaml](../config/manager/env.yaml).
```
make generate && make manifests
make bundle
```

## Deploy hw-event-proxy using operator
```
# under hw-event-proxy-operator repo:
export IMG=quay.io/openshift/origin-baremetal-hardware-event-proxy-operator:latest
make deploy

# check operator logs:
oc -n hw-event-proxy-operator-system logs -f hw-event-proxy-operator-controller-manager-fd95dfff4-7cngx -c manager

# find the worker node name
oc get nodes

# under hw-event-proxy repo:
oc label --overwrite node <name of worker node> app=local
make deploy-amq

oc apply -f tests/e2e/manifest/secret.yaml -n hw-event-proxy-operator-system

# deploy hw-event-proxy. Update the amqp host address
#   transportHost: "amqp://router.router.svc.cluster.local"
oc apply -f config/samples/event_v1alpha1_hardwareevent.yaml -n hw-event-proxy-operator-system

```