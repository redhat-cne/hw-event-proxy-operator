# Developer Guide

## Build image
```
make build
make bin
podman build . -f Dockerfile.dev -t quay.io/redhat-cne/hw-event-proxy-operator:latest
podman push quay.io/redhat-cne/hw-event-proxy-operator:latest
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
export IMG=quay.io/redhat-cne/hw-event-proxy-operator:latest
make deploy

# check operator logs:
oc -n openshift-bare-metal-events logs -f hw-event-proxy-operator-controller-manager-fd95dfff4-7cngx -c manager

# deploy amqp if AMQP transport is used. Under hw-event-proxy repo:
make deploy-amq

# Create redfish secret
# replace the following with real Redfish credentials and BMC ip address
export REDFISH_USERNAME=<user>
export REDFISH_PASSWORD=<password>
export REDFISH_HOSTADDR=<ip address>
oc -n openshift-bare-metal-events create secret generic redfish-basic-auth \
--from-literal=username=$REDFISH_USERNAME \
--from-literal=password=$REDFISH_PASSWORD \
--from-literal=hostaddr="$REDFISH_HOSTADDR"

# deploy hw-event-proxy. Update the transportHost before apply.
oc apply -f config/samples/event_v1alpha1_hardwareevent.yaml -n openshift-bare-metal-events
```

## Deploy example consumer
under [hw-event-proxy](https://github.com/redhat-cne/hw-event-proxy) repo:
```
make deploy-consumer
```
