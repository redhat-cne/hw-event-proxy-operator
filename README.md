# Hardware Event Proxy Operator
## Table of Contents

- [HW-Event-Proxy Operator](#hw-event-proxy-operator)
- [HardwareEvent](#hardwareevent)
- [Quick Start](#quick-start)

## HW-Event-Proxy Operator
hw-event-proxy operator, runs in single namespace, manages baremetal hardware event framework.
It offers `HardwareEvent` CRDs and deploys `hw-event-proxy` container along with `cloud-event-proxy` cloud event framework.


## HardwareEvent
To start using Hardware event proxy, install the operator and then create the following HardwareEvent:
You can only create one instance of HardwareEvent in this release.
```
apiVersion: "event.redhat-cne.org/v1alpha1"
kind: "HardwareEvent"
metadata:
  name: "hardware-event"
spec:
 nodeSelector: {}
 transportHost: "amqp://amq-router-service-name.amq-namespace.svc.cluster.local"
```
The amqp host specified is needed to deliver the events at transport layer using AMQP protocol.

This operator needs a secret to be created to access Redfish
Message Registry.
The secret name and spec the operand expects are under the same namespace
the operand is deployed on. For example:
```
apiVersion: v1
kind: Secret
metadata:
  name: redfish-basic-auth
type: Opaque
data:
  username: cm9vdA==
  password: Y2Fsdmlu
stringData:
  # BMC host DNS or IP address
  hostaddr: 10.46.61.142
```

Below is an example of updating `nodeSelector` to select a specific nodes:

```
apiVersion: "event.redhat-cne.org/v1alpha1"
kind: "HardwareEvent"
metadata:
  name: "hardware-event"
spec:
  nodeSelector:
    node-role.kubernetes.io/hw-event: ""
  transportHost: "amqp://amq-router-service-name.amq-namespace.svc.cluster.local"
```

## Quick Start

To install hw-event-proxy Operator:
```
export IMG=quay.io/openshift/origin-baremetal-hardware-event-proxy-operator:4.10
$ make deploy
```

To un-install:
```
$ make undeploy
```
Refer to [Developer Guide](docs/development.md) for more details.
