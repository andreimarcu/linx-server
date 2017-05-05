FROM golang:alpine

RUN set -ex \
        && apk add --no-cache --virtual .build-deps git \
        && go get github.com/andreimarcu/linx-server \
        && apk del .build-deps

VOLUME ["/data/files", "/data/meta"]

EXPOSE 8080
USER nobody
ENTRYPOINT ["/go/bin/linx-server", "-bind=0.0.0.0:8080", "-filespath=/data/files/", "-metapath=/data/meta/"]
CMD ["-sitename=linx", "-allowhotlink"]
