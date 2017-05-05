FROM alpine:latest
MAINTAINER CodeHat <dev@codehat.de>

LABEL linx-server_version="1.2.3" architecture="amd64"

RUN apk add --update --no-cache curl

RUN curl -L https://github.com/andreimarcu/linx-server/releases/download/v1.2.3/linx-server-v1.2.3_linux-amd64 -o /usr/bin/linx-server \
    && chmod 0755 /usr/bin/linx-server \
    && mkdir -p /srv/config /srv/storage/files /srv/storage/meta

EXPOSE 8080
WORKDIR /srv

COPY config.ini /srv/config/

ENTRYPOINT ["/usr/bin/linx-server"]
CMD ["-config", "/srv/config/config.ini"]