FROM golang:alpine as builder

WORKDIR /build

COPY . .

RUN set -ex \
	&& apk update \
	&& apk upgrade \
	&& apk add --no-cache build-base \
	&& apk add --no-cache ca-certificates \
	&& update-ca-certificates \
	&& go mod download \
	&& go mod verify \
	&& GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o app . \
	&& chmod +x /build/app

FROM alpine

EXPOSE 8080

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/app ./run

RUN set -ex \
	&& mkdir -p /data \
	&& ln -s /data/pushbits.db /app/pushbits.db \
	&& ln -s /data/config.yml /app/config.yml

USER 1000

ENTRYPOINT ["./run"]
