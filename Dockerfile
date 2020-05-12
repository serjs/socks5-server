ARG GOLANG_VERSION="1.14.2"

FROM golang:$GOLANG_VERSION-alpine as builder
WORKDIR /go/src/github.com/serjs/socks5-server
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w" -a -installsuffix cgo -o socks5 .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/serjs/socks5/socks5 /
ENTRYPOINT ["/socks5"]
