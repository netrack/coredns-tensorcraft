ARG GOVERSION=1.14

FROM golang:${GOVERSION}-buster

# All args after each FROM command are no longer available.
ARG COREDNSVERSION=v1.6.4
ARG TENSORFLOWVERSION=1.15.0

RUN apt-get update && apt-get -uy upgrade
RUN apt-get -y install ca-certificates && update-ca-certificates

ENV COREDNSPATH github.com/coredns/coredns
ENV DNSTUNPATH github.com/netrack/dnstun
ENV TENSORFLOWPATH storage.googleapis.com/tensorflow/libtensorflow
ENV GO111MODULE on

RUN curl -fsSL https://${COREDNSPATH}/archive/${COREDNSVERSION}.tar.gz -o coredns.tar.gz \
    && mkdir -p coredns \
    && tar -xzf coredns.tar.gz --strip-components=1 -C coredns \
    && rm -rf coredns.tar.gz


RUN curl -fsSL https://${TENSORFLOWPATH}/libtensorflow-cpu-linux-x86_64-${TENSORFLOWVERSION}.tar.gz -o tensorflow.tar.gz \
    && tar -xzf tensorflow.tar.gz -C /usr/ \
    && rm -rf tensorflow.tar.gz \
    && ldconfig

COPY . ${GOPATH}/src/${DNSTUNPATH}
COPY plugin.cfg coredns/plugin.cfg

WORKDIR coredns

RUN go mod edit -require ${DNSTUNPATH}@v0.0.0
RUN go mod edit -replace ${DNSTUNPATH}@v0.0.0=${GOPATH}/src/${DNSTUNPATH}

RUN go generate && go build -o /bin/coredns


FROM debian:buster-slim
COPY --from=0 /etc/ssl/certs /etc/ssl/certs
COPY --from=0 /usr/lib/libtensorflow* /usr/lib/
COPY --from=0 /bin/coredns /bin/coredns
COPY Corefile /etc/coredns/Corefile
VOLUME /etc/coredns

EXPOSE 53 53/udp
CMD ["/bin/coredns", "-conf", "/etc/coredns/Corefile"]
