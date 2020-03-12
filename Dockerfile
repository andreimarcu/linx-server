FROM golang:1.14-alpine3.11 AS build

COPY . /go/src/github.com/andreimarcu/linx-server
WORKDIR /go/src/github.com/andreimarcu/linx-server

RUN set -ex \
        && apk add --no-cache --virtual .build-deps git \
        && go get -v . \
        && apk del .build-deps

FROM alpine:3.11

COPY --from=build /go/bin/linx-server /usr/local/bin/linx-server

ENV GOPATH /go
ENV SSL_CERT_FILE /etc/ssl/cert.pem

COPY static /go/src/github.com/andreimarcu/linx-server/static/
COPY templates /go/src/github.com/andreimarcu/linx-server/templates/

RUN mkdir -p /data/files && mkdir -p /data/meta && chown -R 65534:65534 /data

VOLUME ["/data/files", "/data/meta"]

EXPOSE 8080
USER nobody
ENTRYPOINT ["/usr/local/bin/linx-server", "-bind=0.0.0.0:8080", "-filespath=/data/files/", "-metapath=/data/meta/"]
CMD ["-sitename=linx", "-allowhotlink"]
