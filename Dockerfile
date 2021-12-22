# Build the manager binary
FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.17-openshift-4.10 AS builder
WORKDIR /go/src/github.com/redhat-cne/hw-event-proxy-operator
COPY . .
ENV GO111MODULE=off
RUN make
# Build

FROM registry.ci.openshift.org/ocp/4.10:base
WORKDIR /
COPY --from=builder /go/src/github.com/redhat-cne/hw-event-proxy-operator/build/_output/bin/hw-event-proxy-operator /
COPY --from=builder /go/src/github.com/redhat-cne/hw-event-proxy-operator/bindata /bindata
USER 65532:65532

ENTRYPOINT ["/hw-event-proxy-operator"]
