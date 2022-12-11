FROM docker.io/library/golang:alpine as builder

ARG PB_BUILD_VERSION

ARG CLI_VERSION=0.0.6
ARG CLI_PLATFORM=linux_amd64

WORKDIR /build

COPY . .

RUN set -ex \
	&& apk add --no-cache build-base ca-certificates curl \
	&& go mod download \
	&& go mod verify \
	&& PB_BUILD_VERSION="$PB_BUILD_VERSION" make build \
	&& chmod +x /build/out/pushbits \
	&& curl -q -s -S -L -o /tmp/pbcli_${CLI_VERSION}.tar.gz https://github.com/pushbits/cli/releases/download/v${CLI_VERSION}/pbcli_${CLI_VERSION}_${CLI_PLATFORM}.tar.gz \
	&& tar -C /usr/local/bin -xvf /tmp/pbcli_${CLI_VERSION}.tar.gz pbcli \
	&& chown root:root /usr/local/bin/pbcli \
	&& chmod +x /usr/local/bin/pbcli

FROM docker.io/library/alpine

ARG USER_ID=1000

ENV PUSHBITS_HTTP_PORT="8080"

EXPOSE 8080

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/out/pushbits ./run
COPY --from=builder /usr/local/bin/pbcli /usr/local/bin/pbcli

RUN set -ex \
	&& apk add --no-cache ca-certificates curl \
	&& update-ca-certificates \
	&& mkdir -p /data \
	&& ln -s /data/pushbits.db /app/pushbits.db \
	&& ln -s /data/config.yml /app/config.yml

USER ${USER_ID}

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s CMD curl --fail http://localhost:$PUSHBITS_HTTP_PORT/health || exit 1

ENTRYPOINT ["./run"]
