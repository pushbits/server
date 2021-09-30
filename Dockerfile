FROM docker.io/library/golang:alpine as builder

WORKDIR /build

COPY . .

RUN set -ex \
	&& apk add --no-cache build-base \
	&& go mod download \
	&& go mod verify \
	&& make build \
	&& chmod +x /build/out/pushbits

FROM docker.io/library/alpine

ARG USER_ID=1000

ENV PUSHBITS_HTTP_PORT="8080"

EXPOSE 8080

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/out/pushbits ./run

RUN set -ex \
	&& apk add --no-cache ca-certificates curl \
	&& update-ca-certificates \
	&& mkdir -p /data \
	&& ln -s /data/pushbits.db /app/pushbits.db \
	&& ln -s /data/config.yml /app/config.yml

USER ${USER_ID}

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s CMD curl --fail http://localhost:$PUSHBITS_HTTP_PORT/health || exit 1

ENTRYPOINT ["./run"]
