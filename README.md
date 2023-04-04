# Bare Metal Event Relay Operator

> **_NOTE:_** `Bare Metal Event Relay` was renamed from `Hardware Event Proxy`.

The Bare Metal Event Relay Operator, runs in a single namespace, manages bare metal hardware event framework.
It offers `HardwareEvent` CRDs and deploys [Bare Metal Event Relay](https://github.com/redhat-cne/hw-event-proxy) container along with [Cloud Event Proxy](https://github.com/redhat-cne/cloud-event-proxy) framework.

 [![go-doc](https://godoc.org/github.com/redhat-cne/hw-event-proxy-operator?status.svg)](https://godoc.org/github.com/redhat-cne/hw-event-proxy-operator)
 [![Go Report Card](https://goreportcard.com/badge/github.com/redhat-cne/hw-event-proxy-operator)](https://goreportcard.com/report/github.com/redhat-cne/hw-event-proxy-operator)
 [![LICENSE](https://img.shields.io/github/license/redhat-cne/hw-event-proxy-operator.svg)](https://github.com/redhat-cne/hw-event-proxy-operator/blob/main/LICENSE)

## HardwareEvent
To start using Bare Metal Event Relay, install the operator and then create the following HardwareEvent CR.
You can only create one instance of HardwareEvent in this release. Examples:

### With HTTP Transport
```
apiVersion: "event.redhat-cne.org/v1alpha1"
kind: "HardwareEvent"
metadata:
  name: "hardware-event"
spec:
  nodeSelector: {}
  transportHost: "http://hw-event-publisher-service.openshift-bare-metal-events.svc.cluster.local:9043"
  storageType: "example-storage-class"
```
Here the transport is set to `hw-event-publisher-service` service in the `openshift-bare-metal-events` namespace.

The `storageType` is for HTTP transport only. HTTP transport requires persistent storage for storing subscription data. The `storageType` must be set to the name of StorageClass providing the persist storage.

If `transportHost` is missing or empty, the default transportHost `"http://hw-event-publisher-service.openshift-bare-metal-events.svc.cluster.local:9043"` is used. In this case a valid `storageType` is still required.

A special storageType `emptyDir` is used for developers only. It provides ephemeral storage for HTTP transport and is used for developer testings.

### With AMQP Transport
```
apiVersion: "event.redhat-cne.org/v1alpha1"
kind: "HardwareEvent"
metadata:
  name: "hardware-event"
spec:
  nodeSelector: {}
  transportHost: "amqp://<amq-router-service-name>.<amq-namespace>.svc.cluster.local"
```
The AMQP service `<amq-router-service-name>` should be available in `<amq-namespace>` namespace prior to deploying HardwareEvent.

### Select Node
Below is an example of updating `nodeSelector` to select a specific node:
```
apiVersion: "event.redhat-cne.org/v1alpha1"
kind: "HardwareEvent"
metadata:
  name: "hardware-event"
spec:
  nodeSelector:
    node-role.kubernetes.io/worker: ""
  transportHost: "http://hw-event-publisher-service.openshift-bare-metal-events.svc.cluster.local:9043"
```

### Create Secret for Redfish Authentication
This operator needs a secret to be created to access Redfish Message Registry.

```
oc -n openshift-bare-metal-events create secret generic redfish-basic-auth \
  --from-literal=username=${BMC_USER} \
  --from-literal=password=${BMC_PASSWORD} \
  --from-literal=hostaddr=${BMC_HOST}
```

The secret is created under the same namespace as the operator. For example:
```
apiVersion: v1
kind: Secret
metadata:
  name: redfish-basic-auth
type: Opaque
stringData:
  username: <bmc_username>
  password: <bmc_password>
  # BMC host DNS or IP address
  hostaddr: <bmc_host_ip_address>
```

## Quick Start

To install the Bare Metal Event Relay Operator:
```
export IMG=quay.io/openshift/origin-baremetal-hardware-event-proxy-operator:4.13
$ make deploy
```

To un-install:
```
$ make undeploy
```
Refer to [Developer Guide](docs/development.md) for more details.
