FROM registry.ci.openshift.org/openshift/release:golang-1.16 as builder

WORKDIR /hypershift

ENV PUSHGATEWAY_VERSION="1.3.0"

COPY . .

RUN cd /tmp && \
    curl -OL https://github.com/prometheus/pushgateway/releases/download/v${PUSHGATEWAY_VERSION}/pushgateway-${PUSHGATEWAY_VERSION}.linux-amd64.tar.gz && \
    tar -xzf pushgateway-${PUSHGATEWAY_VERSION}.linux-amd64.tar.gz && \
    cp pushgateway-${PUSHGATEWAY_VERSION}.linux-amd64/pushgateway /pushgateway
RUN make build

FROM quay.io/openshift/origin-base:4.9
COPY --from=builder /hypershift/bin/ignition-server /usr/bin/ignition-server
COPY --from=builder /hypershift/bin/hypershift /usr/bin/hypershift
COPY --from=builder /hypershift/bin/hypershift-operator /usr/bin/hypershift-operator
COPY --from=builder /hypershift/bin/control-plane-operator /usr/bin/control-plane-operator
COPY --from=builder /hypershift/bin/hosted-cluster-config-operator /usr/bin/hosted-cluster-config-operator
COPY --from=builder /hypershift/bin/roks-metrics/roksmetrics /usr/bin/roks-metrics
COPY --from=builder /hypershift/bin/roks-metrics/metric-pusher /usr/bin/metric-pusher
COPY --from=builder /hypershift/bin/konnectivity-socks5-proxy /usr/bin/konnectivity-socks5-proxy

ENTRYPOINT /usr/bin/hypershift
