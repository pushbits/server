FROM golang:latest as builder

WORKDIR /build

COPY . .

RUN set -ex \
	&& go get -d -v \
	&& CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM scratch

WORKDIR /

COPY --from=builder /build/app .

USER 1000

CMD ["./app"]
