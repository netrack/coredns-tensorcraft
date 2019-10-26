ARG GOVERSION=1.12

FROM golang:${GOVERSION}-buster

# All args after each FROM command are no longer available.
ARG COREDNSVERSION=v1.6.4

RUN apt-get update && apt-get -uy upgrade
RUN apt-get -y install ca-certificates && update-ca-certificates

ENV COREDNSPATH github.com/coredns/coredns
ENV DNSTUNPATH github.com/netrack/dnstun
ENV GO111MODULE on
ENV CGO_ENABLED 0

RUN curl -fsSL https://${COREDNSPATH}/archive/${COREDNSVERSION}.tar.gz -o coredns.tar.gz \
    && mkdir -p coredns \
    && tar -xzf coredns.tar.gz --strip-components=1 -C coredns \
    && rm -rf coredns.tar.gz

COPY . ${GOPATH}/src/${DNSTUNPATH}
COPY plugin.cfg coredns/plugin.cfg

WORKDIR coredns

RUN go mod edit -require ${DNSTUNPATH}@v0.0.0
RUN go mod edit -replace ${DNSTUNPATH}@v0.0.0=${GOPATH}/src/${DNSTUNPATH}

RUN go generate && go build -o /bin/coredns


FROM scratch
COPY --from=0 /etc/ssl/certs /etc/ssl/certs
COPY --from=0 /bin/coredns /bin/coredns
COPY Corefile /etc/coredns/Corefile
VOLUME /etc/coredns

EXPOSE 53 53/udp
CMD ["/bin/coredns", "-conf", "/etc/coredns/Corefile"]
